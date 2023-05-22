// Package storage содержит хранилище для сервера.
package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres init for golang-migrate
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
)

//go:embed migrations
var migrationsFS embed.FS

// ErrNotFound - запись не найдена.
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
		return nil, fmt.Errorf("storage NewServerStorage: sql open error: %w", err)
	}

	err = s.doMigrate(dsn)
	if err != nil {
		return nil, fmt.Errorf("storage NewServerStorage: migrate error: %w", err)
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
		return entry{}, fmt.Errorf("ServerStorage Get: query row: %w", ErrNotFound)
	}
	if err != nil {
		return entry{}, fmt.Errorf("ServerStorage Get: query row: %w", err)
	}

	return e, nil
}

// GetAll - получить все записи по publicKey.
func (s *ServerStorage) GetAll(ctx context.Context, publicKey []byte) ([]entry, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT id, payload FROM entries WHERE public_key = $1`, publicKey)
	if errors.Is(err, sql.ErrNoRows) {
		return []entry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ServerStorage GetAll: query: %w", err)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ServerStorage GetAll: query rows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := make([]entry, 0)
	for rows.Next() {
		e := entry{}
		err = rows.Scan(&e.ID, &e.Payload)
		if err != nil {
			return nil, fmt.Errorf("ServerStorage GetAll: query rows scan: %w", err)
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
		return uuid.Nil, fmt.Errorf("ServerStorage Create: query row scan: %w", err)
	}

	return id, nil
}

// Delete - удалить запись по ID.
func (s *ServerStorage) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM entries WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("ServerStorage Delete: exec: %w", err)
	}

	return nil
}

// Update - обновляем запись по ID.
func (s *ServerStorage) Update(ctx context.Context, id uuid.UUID, data []byte) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `UPDATE entries SET payload = $1 WHERE id = $2`, data, id)
	if err != nil {
		return fmt.Errorf("ServerStorage Update: exec: %w", err)
	}

	return nil
}

// Close - закрываем соединение с базой данных.
func (s *ServerStorage) Close() error {
	err := s.db.Close()
	if err != nil {
		return fmt.Errorf("ServerStorage Close: %w", err)
	}

	return nil
}

func (s *ServerStorage) doMigrate(dsn string) error {
	d, err := iofs.New(migrationsFS, "migrations/server")
	if err != nil {
		return fmt.Errorf("ServerStorage doMigrate: iofs: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("ServerStorage doMigrate: new migrate: %w", err)
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("ServerStorage doMigrate: migrate up: %w", err)
	}

	return nil
}
