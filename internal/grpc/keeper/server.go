// Package keeper хранит grpc сервер и клиент для GophKeeper.
package keeper

import (
	"github.com/ImpressionableRaccoon/GophKeeper/internal/storage"
	pb "github.com/ImpressionableRaccoon/GophKeeper/proto"
)

type server struct {
	pb.UnimplementedKeeperServer

	s *storage.ServerStorage
}

// NewServer - конструктор для grpc сервера GophKeeper.
func NewServer(s *storage.ServerStorage) *server {
	return &server{
		s: s,
	}
}
