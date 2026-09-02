package db

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
)

func newTestDB(t *testing.T) (*DB, sqlxmock.Sqlmock) {
	db, mock, err := sqlxmock.Newx()
	require.NoError(t, err)

	return &DB{
		log:  slog.Default(),
		conn: db,
	}, mock
}

func TestDB_Drop_Success(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectBegin()

	mock.ExpectExec("TRUNCATE TABLE").
		WillReturnResult(sqlxmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := db.Drop(context.Background())

	require.NoError(t, err)
}

func TestDB_Drop_Error(t *testing.T) {
	db, mock := newTestDB(t)

	mock.ExpectBegin()

	mock.ExpectExec("TRUNCATE TABLE").
		WillReturnError(errors.New("fail"))

	mock.ExpectRollback()

	err := db.Drop(context.Background())

	require.Error(t, err)
}
