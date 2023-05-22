// Package service содержит слой для обработки действий клиента.
package service

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chzyer/readline"

	"github.com/ImpressionableRaccoon/GophKeeper/internal/dataverse"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/grpc/keeper"
	pb "github.com/ImpressionableRaccoon/GophKeeper/proto"
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

// Get - получить запись по ID.
func (s Service) Get(ctx context.Context, id string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("service Service Get: context: %w", err)
	}

	resp, err := s.c.Get(ctx, &pb.GetRequest{
		Id: id,
	})
	if err != nil {
		return "", fmt.Errorf("service Service Get: client: %w", err)
	}

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, s.key, resp.Data)
	if err != nil {
		return "", fmt.Errorf("service Service Get: decrypt: %w", err)
	}

	e, err := dataverse.ParseEntry(decrypted)
	if err != nil {
		return "", fmt.Errorf("service Service Get: parse entry: %w", err)
	}

	b := strings.Builder{}
	_, _ = fmt.Fprintf(&b, "Type: %s\n", e.GetType())
	_, _ = fmt.Fprintf(&b, "Name: %s\n", e.GetName())
	b.WriteString(e.GetContent())

	return b.String(), nil
}

// Add - добавить новую запись.
func (s Service) Add(ctx context.Context, line string, l *readline.Instance) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("service Service Add: context: %w", err)
	}

	t := strings.TrimSpace(line)

	e, err := dataverse.GenDatabaseEntry(t, l)
	if err != nil {
		return "", fmt.Errorf("service Service Add: gen entry: %w", err)
	}

	data, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("service Service Add: marshal json: %w", err)
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, &s.key.PublicKey, data)
	if err != nil {
		return "", fmt.Errorf("service Service Add: encrypt: %w", err)
	}

	hash := sha256.Sum256(encrypted)

	sign, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("service Service Add: sign: %w", err)
	}

	resp, err := s.c.Create(ctx, &pb.CreateRequest{
		PublicKey: s.key.PublicKey.N.Bytes(),
		Data:      encrypted,
		Sign:      sign,
	})
	if err != nil {
		return "", fmt.Errorf("service Service Add: client: %w", err)
	}

	return fmt.Sprintf("Entry ID: %s", resp.Id), nil
}

// All - получить все записи пользователя.
func (s Service) All(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("service Service All: context: %w", err)
	}

	resp, err := s.c.GetAll(ctx, &pb.GetAllRequest{
		PublicKey: s.key.PublicKey.N.Bytes(),
	})
	if err != nil {
		return "", fmt.Errorf("service Service All: client: %w", err)
	}

	b := strings.Builder{}
	for _, entry := range resp.Entries {
		var decrypted []byte
		decrypted, err = rsa.DecryptPKCS1v15(rand.Reader, s.key, entry.Data)
		if err != nil {
			_, _ = fmt.Fprintf(&b, "%s\tdecrypt failed\n", entry.Id)
			continue
		}

		var e dataverse.Entry
		e, err = dataverse.ParseEntry(decrypted)
		if err != nil {
			_, _ = fmt.Fprintf(&b, "%s\tparsing failed\t%s\n", entry.Id, err)
			continue
		}

		_, _ = fmt.Fprintf(&b, "%s\t%s\t%s\n", entry.Id, e.GetType(), e.GetName())
	}

	return b.String(), nil
}

// Delete - удалить запись по ID.
func (s Service) Delete(ctx context.Context, id string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("service Service Delete: context: %w", err)
	}

	getResp, err := s.c.Get(ctx, &pb.GetRequest{
		Id: id,
	})
	if err != nil {
		return "", fmt.Errorf("service Service Delete: client get: %w", err)
	}

	hash := sha256.Sum256(getResp.Data)
	hash2 := sha256.Sum256(hash[:])

	sign2, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, hash2[:])
	if err != nil {
		return "", fmt.Errorf("service Service Delete: sign: %w", err)
	}

	_, err = s.c.Delete(ctx, &pb.DeleteRequest{
		Id:   id,
		Sign: sign2,
	})
	if err != nil {
		return "", fmt.Errorf("service Service Delete: client delete: %w", err)
	}

	return fmt.Sprintf("Entry %s successfully deleted", id), nil
}

// Update - обновить запись по ID.
func (s Service) Update(ctx context.Context, line string, l *readline.Instance) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("service Service Update: context: %w", err)
	}

	splitted := strings.Split(strings.TrimSpace(line), " ")
	id := splitted[0]
	t := splitted[1]

	getResp, err := s.c.Get(ctx, &pb.GetRequest{
		Id: id,
	})
	if err != nil {
		return "", fmt.Errorf("service Service Update: client get: %w", err)
	}

	oldHash := sha256.Sum256(getResp.Data)
	oldHash2 := sha256.Sum256(oldHash[:])
	oldSign, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, oldHash2[:])
	if err != nil {
		return "", fmt.Errorf("service Service Update: sign old: %w", err)
	}

	e, err := dataverse.GenDatabaseEntry(t, l)
	if err != nil {
		return "", fmt.Errorf("service Service Update: gen entry: %w", err)
	}

	data, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("service Service Update: marshal json: %w", err)
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, &s.key.PublicKey, data)
	if err != nil {
		return "", fmt.Errorf("service Service Update: encrypt: %w", err)
	}

	newHash := sha256.Sum256(encrypted)
	newSign, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, newHash[:])
	if err != nil {
		return "", fmt.Errorf("service Service Update: sign new: %w", err)
	}

	_, err = s.c.Update(ctx, &pb.UpdateRequest{
		Id:      id,
		Data:    encrypted,
		SignOld: oldSign,
		SignNew: newSign,
	})
	if err != nil {
		return "", fmt.Errorf("service Service Update: client update: %w", err)
	}

	return "update ok", nil
}
