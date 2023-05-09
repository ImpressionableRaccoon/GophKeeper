// Package storage содержит хранилища для клиентов и серверов.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres init for golang-migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file init for golang-migrate
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("entry not found")

type entry struct {
	ID        uuid.UUID
	PublicKey []byte
	Payload   []byte
}

// ServerStorage - хранилище для сервера.
type ServerStorage struct {
	db *sql.DB
}

// NewServerStorage - создаем новое хранилище для сервера.
func NewServerStorage(dsn string) (*ServerStorage, error) {
	s := &ServerStorage{}

	var err error
	s.db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = s.doMigrate(dsn)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Get - получить запись по ID.
func (s *ServerStorage) Get(ctx context.Context, id uuid.UUID) (e entry, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	e = entry{
		ID: id,
	}

	row := s.db.QueryRowContext(ctx, `SELECT public_key, payload FROM entries WHERE id = $1`, id)
	err = row.Scan(&e.PublicKey, &e.Payload)
	if errors.Is(err, sql.ErrNoRows) {
		return entry{}, ErrNotFound
	}
	if err != nil {
		return entry{}, err
	}

	return e, nil
}

// GetAll - получить все записи по publicKey.
func (s *ServerStorage) GetAll(ctx context.Context, publicKey []byte) ([]entry, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT id, payload FROM entries WHERE public_key = $1`, publicKey)
	if err != nil {
		return nil, err
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	entries := make([]entry, 0)
	for rows.Next() {
		e := entry{}
		err = rows.Scan(&e.ID, &e.Payload)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// Create - добавить запись и вернуть ID.
func (s *ServerStorage) Create(ctx context.Context, publicKey []byte, data []byte) (id uuid.UUID, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	row := s.db.QueryRowContext(ctx,
		`INSERT INTO entries (public_key, payload) VALUES ($1, $2) RETURNING id`,
		publicKey, data,
	)
	err = row.Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// Delete - удалить запись по ID.
func (s *ServerStorage) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM entries WHERE id = $1`, id)
	return err
}

// Close - закрываем соединение с базой данных.
func (s *ServerStorage) Close() error {
	return s.db.Close()
}

func (s *ServerStorage) doMigrate(dsn string) error {
	m, err := migrate.New("file://migrations/server/postgres", dsn)
	if err != nil {
		return err
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	return err
}
