package keeper

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"math/big"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/ImpressionableRaccoon/GophKeeper/internal/storage"
	pb "github.com/ImpressionableRaccoon/GophKeeper/proto"
)

const publicE = 65537

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

// Get - обработчик для получения записи.
func (s server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to parse UUID: %s", err)
	}

	entry, err := s.s.Get(ctx, id)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "entry not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "storage error: %s", err)
	}

	return &pb.GetResponse{
		Data: entry.Payload,
	}, nil
}

// GetAll - обработчик для получения всех данных пользователя.
func (s server) GetAll(ctx context.Context, req *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	entries, err := s.s.GetAll(ctx, req.PublicKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "storage error: %s", err)
	}

	e := make([]*pb.GetAllResponse_Entry, 0, len(entries))
	for _, entry := range entries {
		e = append(e, &pb.GetAllResponse_Entry{
			Id:   entry.ID.String(),
			Data: entry.Payload,
		})
	}

	return &pb.GetAllResponse{
		Entries: e,
	}, nil
}

// Create - обработчик для сохранения новой записи.
func (s server) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	publicN := big.Int{}
	publicN.SetBytes(req.PublicKey)
	public := rsa.PublicKey{
		N: &publicN,
		E: publicE,
	}

	hash := sha256.Sum256(req.Data)

	err := rsa.VerifyPKCS1v15(&public, crypto.SHA256, hash[:], req.Sign)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "sign verify failed: %s", err)
	}

	id, err := s.s.Create(ctx, req.PublicKey, req.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "storage error: %s", err)
	}

	return &pb.CreateResponse{
		Id: id.String(),
	}, nil
}

// Delete - обработчик для удаления записи.
func (s server) Delete(ctx context.Context, req *pb.DeleteRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to parse UUID: %s", err)
	}

	entry, err := s.s.Get(ctx, id)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "entry not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "storage error on get: %s", err)
	}

	publicN := big.Int{}
	publicN.SetBytes(entry.PublicKey)
	public := rsa.PublicKey{
		N: &publicN,
		E: publicE,
	}

	hash := sha256.Sum256(entry.Payload)
	hash2 := sha256.Sum256(hash[:])

	err = rsa.VerifyPKCS1v15(&public, crypto.SHA256, hash2[:], req.Sign)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "sign verify failed: %s", err)
	}

	err = s.s.Delete(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "storage error on delete: %s", err)
	}

	return &emptypb.Empty{}, nil
}
