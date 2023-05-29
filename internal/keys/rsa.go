// Package keys содержит методы, для взаимодействия с ключами пользователей.
package keys

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

var keySize = 4096

const keyType = "RSA PRIVATE KEY"

// GenRSAKey генерирует и сохраняет RSA-ключ в pem-файл.
func GenRSAKey(ctx context.Context) (_ *rsa.PrivateKey, fileName string, _ error) {
	if err := ctx.Err(); err != nil {
		return nil, "", fmt.Errorf("rsa GenRSAKey: context: %w", err)
	}

	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, "", fmt.Errorf("rsa GenRSAKey: generate key: %w", err)
	}

	bytes := x509.MarshalPKCS1PrivateKey(key)

	fileName = fmt.Sprintf("%0x.pem", bytes[len(bytes)-16:])

	privateKeyPEM := &pem.Block{
		Type:  keyType,
		Bytes: bytes,
	}

	pemFile, err := os.Create(fileName)
	if err != nil {
		return nil, "", fmt.Errorf("rsa GenRSAKey: create file: %w", err)
	}
	defer func() { _ = pemFile.Close() }()

	err = pem.Encode(pemFile, privateKeyPEM)
	if err != nil {
		return nil, "", fmt.Errorf("rsa GenRSAKey: encode: %w", err)
	}

	return key, fileName, nil
}

// LoadRSAKey загружает RSA ключ из pem-файла.
func LoadRSAKey(ctx context.Context, keyPath string) (*rsa.PrivateKey, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("rsa LoadRSAKey: context: %w", err)
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("rsa LoadRSAKey: read file: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != keyType {
		return nil, errors.New("rsa LoadRSAKey: decode: wrong block format")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("rsa LoadRSAKey: parse key: %w", err)
	}

	return key, nil
}
