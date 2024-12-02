package repository

import (
	"AuthService/pkg"
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

type testAdd struct {
	mock   sqlmock.Sqlmock
	ctx    context.Context
	repo   Repository
	user   pkg.User
	hash   string
	testId int64
}

func AddInMock(t *testing.T, tt *testAdd) {
	t.Helper()

	tt.mock.ExpectBegin()

	tt.mock.ExpectQuery("INSERT INTO sessions").
		WithArgs(tt.hash, tt.user.IP).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tt.testId))
	tt.mock.ExpectExec("INSERT INTO user_session").
		WithArgs(tt.user.UserId, tt.testId).
		WillReturnResult(sqlmock.NewResult(tt.testId, 1))

	tt.mock.ExpectCommit()

	err := tt.repo.AddSession(tt.ctx, tt.user, tt.hash)
	require.NoError(t, err)

	require.NoError(t, tt.mock.ExpectationsWereMet())
}

func TestRepository_AddSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	repo := New(db)
	ctx := context.Background()

	insertError := errors.New("insert error")

	t.Run("RollBack_1", func(t *testing.T) {
		hash := "hash"
		ip := "0.0.0.0"
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO sessions").
			WithArgs(hash, ip).
			WillReturnError(insertError)
		mock.ExpectRollback()

		err = repo.AddSession(ctx, pkg.User{UserId: "1", IP: ip}, hash)
		require.ErrorIs(t, err, insertError)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RollBack_2", func(t *testing.T) {
		mock.ExpectBegin()
		hash := "hash"
		ip := "0.0.0.0"
		mock.ExpectQuery("INSERT INTO sessions").
			WithArgs(hash, ip).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectExec("INSERT INTO user_session").
			WithArgs("1", 1).
			WillReturnError(insertError)
		mock.ExpectRollback()

		err = repo.AddSession(ctx, pkg.User{UserId: "1", IP: ip}, hash)
		require.ErrorIs(t, err, insertError)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Add", func(t *testing.T) {
		for testId := 1; testId <= 100; testId++ {
			userId := strconv.Itoa(testId)
			AddInMock(t, &testAdd{
				mock:   mock,
				ctx:    ctx,
				repo:   repo,
				user:   pkg.User{UserId: userId, IP: userId},
				hash:   "hash" + userId,
				testId: int64(testId),
			})
		}
	})
}
