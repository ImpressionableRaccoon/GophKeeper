package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServerStorage(t *testing.T) {
	t.Run("wrong dsn", func(t *testing.T) {
		_, err := NewServerStorage("dsn")
		assert.Error(t, err)
	})
}

func TestServerStorage_Get(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		e := entry{
			ID:        uuid.New(),
			PublicKey: []byte{1, 2, 3, 4, 5},
			Payload:   []byte{6, 7, 8, 9, 0},
		}

		rows := sqlmock.NewRows([]string{"public_key", "payload"}).AddRow(e.PublicKey, e.Payload)
		mock.ExpectQuery("SELECT public_key, payload").WithArgs(e.ID).WillReturnRows(rows)

		s := ServerStorage{db: db}
		ctx := context.Background()
		res, err := s.Get(ctx, e.ID)

		assert.NoError(t, err)
		assert.Equal(t, e, res)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		id := uuid.New()

		mock.ExpectQuery("SELECT public_key, payload").WithArgs(id).WillReturnError(sql.ErrNoRows)

		s := ServerStorage{db: db}
		ctx := context.Background()
		_, err = s.Get(ctx, id)

		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("fail", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		id := uuid.New()

		mock.ExpectQuery("SELECT public_key, payload").WithArgs(id).WillReturnError(sql.ErrConnDone)

		s := ServerStorage{db: db}
		ctx := context.Background()
		_, err = s.Get(ctx, id)

		assert.Error(t, err)
	})
}

func TestServerStorage_GetAll(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		publicKey := []byte{1, 2, 3, 4, 5}

		data := map[uuid.UUID][]byte{
			uuid.New(): {1, 2, 3},
			uuid.New(): {4, 5, 6},
			uuid.New(): {7, 8, 9},
		}

		rows := sqlmock.NewRows([]string{"id", "payload"})
		for k, v := range data {
			rows = rows.AddRow(k, v)
		}

		mock.ExpectQuery("SELECT id, payload").WithArgs(publicKey).WillReturnRows(rows)

		s := ServerStorage{db: db}
		ctx := context.Background()
		res, err := s.GetAll(ctx, publicKey)

		assert.NoError(t, err)
		assert.Len(t, res, len(data))
		for _, e := range res {
			assert.Equal(t, data[e.ID], e.Payload)
		}
	})

	t.Run("no entries", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectQuery("SELECT id, payload").WithArgs(sqlmock.AnyArg()).WillReturnError(sql.ErrNoRows)

		s := ServerStorage{db: db}
		ctx := context.Background()
		res, err := s.GetAll(ctx, nil)

		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectQuery("SELECT id, payload").WithArgs(sqlmock.AnyArg()).WillReturnError(sql.ErrConnDone)

		s := ServerStorage{db: db}
		ctx := context.Background()
		_, err = s.GetAll(ctx, nil)

		assert.Error(t, err)
	})
}

func TestServerStorage_Create(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		e := entry{
			ID:        uuid.New(),
			PublicKey: []byte{1, 2, 3, 4, 5},
			Payload:   []byte{6, 7, 8, 9, 0},
		}

		rows := sqlmock.NewRows([]string{"id"}).AddRow(e.ID)
		mock.ExpectQuery("INSERT").WithArgs(e.PublicKey, e.Payload).WillReturnRows(rows)

		s := ServerStorage{db: db}
		ctx := context.Background()
		res, err := s.Create(ctx, e.PublicKey, e.Payload)

		assert.NoError(t, err)
		assert.Equal(t, e.ID, res)
	})

	t.Run("fail", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectQuery("INSERT").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		s := ServerStorage{db: db}
		ctx := context.Background()
		_, err = s.Create(ctx, nil, nil)

		assert.Error(t, err)
	})
}

func TestServerStorage_Delete(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		id := uuid.New()

		mock.ExpectExec("DELETE").WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

		s := ServerStorage{db: db}
		ctx := context.Background()
		err = s.Delete(ctx, id)

		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectExec("DELETE").WithArgs(sqlmock.AnyArg()).WillReturnError(sql.ErrConnDone)

		s := ServerStorage{db: db}
		ctx := context.Background()
		err = s.Delete(ctx, uuid.New())

		assert.Error(t, err)
	})
}

func TestServerStorage_Update(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		id := uuid.New()
		payload := []byte{1, 3, 5, 7, 9}

		mock.ExpectExec("UPDATE").WithArgs(payload, id).WillReturnResult(sqlmock.NewResult(0, 1))

		s := ServerStorage{db: db}
		ctx := context.Background()
		err = s.Update(ctx, id, payload)

		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectExec("UPDATE").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		s := ServerStorage{db: db}
		ctx := context.Background()
		err = s.Update(ctx, uuid.New(), nil)

		assert.Error(t, err)
	})
}

func TestServerStorage_Close(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectClose()

		s := ServerStorage{db: db}
		err = s.Close()

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("already closed", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		mock.ExpectClose().WillReturnError(sql.ErrConnDone)

		s := ServerStorage{db: db}
		err = s.Close()

		assert.ErrorIs(t, err, sql.ErrConnDone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
