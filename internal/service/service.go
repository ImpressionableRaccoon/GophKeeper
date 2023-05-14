// Package service содержит слой для обработки действий клиента.
package service

import (
	"crypto/rsa"

	"github.com/ImpressionableRaccoon/GophKeeper/internal/grpc/keeper"
)

// Service - структура, которая обрабатывает действия клиента.
type Service struct {
	c   *keeper.Client
	key *rsa.PrivateKey
}

// New - создать новый Service.
func New(client *keeper.Client, key *rsa.PrivateKey) (*Service, error) {
	return &Service{
		c:   client,
		key: key,
	}, nil
}

func (s Service) Get(id string) (string, error) {
	return "get method not implemented", nil
}

func (s Service) Add(line string) (string, error) {
	return "add method not implemented", nil
}

func (s Service) All() (string, error) {
	return "all method not implemented", nil
}

func (s Service) Delete(id string) (string, error) {
	return "delete method not implemented", nil
}
