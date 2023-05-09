// Package storage содержит хранилища для клиентов и серверов.
package storage

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres init for golang-migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file init for golang-migrate
)

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
