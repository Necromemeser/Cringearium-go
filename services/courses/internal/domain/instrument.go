package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

// Fingerprints exclude database IDs, so equivalent forms in course variants match.
func InstrumentSnapshot(test *Test) (string, json.RawMessage) {
	type option struct {
		Text    string
		Correct bool
	}
	type question struct {
		Text    string
		Options []option
	}
	normalize := func(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }
	var canonical []string
	for _, q := range test.Questions {
		item := question{Text: normalize(q.Question)}
		for _, a := range q.Answers {
			item.Options = append(item.Options, option{normalize(a.Text), a.IsCorrect})
		}
		sort.Slice(item.Options, func(i, j int) bool {
			if item.Options[i].Text != item.Options[j].Text {
				return item.Options[i].Text < item.Options[j].Text
			}
			return !item.Options[i].Correct && item.Options[j].Correct
		})
		data, _ := json.Marshal(item)
		canonical = append(canonical, string(data))
	}
	sort.Strings(canonical)
	raw, _ := json.Marshal(canonical)
	sum := sha256.Sum256(raw)
	snapshot, _ := json.Marshal(struct {
		Version int
		Test    *Test
	}{1, test})
	return hex.EncodeToString(sum[:]), snapshot
}
