package postgres

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/Necromemeser/Cringearium-go/internal/testsupport"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/domain"
)

func TestConcurrentDraftCreationReusesOnlyOwnersEmptyConversation(t *testing.T) {
	conn := testsupport.Postgres(t)
	db := &DB{conn: conn, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	ctx := context.Background()
	create := func(user int64) (domain.Conversation, error) {
		return repo.CreateOrReuseEmptyConversation(ctx, domain.Conversation{UserID: user, Name: "Новый чат", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	}
	const requests = 24
	ids := make(chan int64, requests)
	errs := make(chan error, requests)
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := create(7)
			if err != nil {
				errs <- err
				return
			}
			ids <- c.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	var first int64
	for id := range ids {
		if first == 0 {
			first = id
		}
		if id != first {
			t.Fatal("concurrent clicks created duplicate drafts")
		}
	}
	var count int
	if err := conn.Get(&count, `SELECT count(*) FROM conversations WHERE user_id=7`); err != nil || count != 1 {
		t.Fatal("unexpected draft count", count, err)
	}
	other, err := create(8)
	if err != nil || other.ID == first {
		t.Fatal("draft crossed ownership boundary", other, err)
	}
	user := int64(7)
	if _, err := repo.SaveMessage(ctx, domain.Message{ConversationID: first, UserID: &user, Content: "Уже начатый разговор"}); err != nil {
		t.Fatal(err)
	}
	next, err := create(7)
	if err != nil || next.ID == first {
		t.Fatal("conversation with messages reused", next, err)
	}
	again, err := create(7)
	if err != nil || again.ID != next.ID {
		t.Fatal("empty draft not reused", again, err)
	}
	if err := repo.DeleteConversation(ctx, next.ID, 7); err != nil {
		t.Fatal(err)
	}
	replacement, err := create(7)
	if err != nil || replacement.ID == next.ID {
		t.Fatal("deleted draft reused", replacement, err)
	}
	messages, err := repo.ListMessages(ctx, first, 7)
	if err != nil || len(messages) != 1 || messages[0].Content != "Уже начатый разговор" {
		t.Fatal("existing messages changed", messages, err)
	}
	if err := conn.Get(&count, `SELECT count(*) FROM ai_daily_usage`); err != nil || count != 0 {
		t.Fatal("draft creation consumed AI quota", count, err)
	}
}
