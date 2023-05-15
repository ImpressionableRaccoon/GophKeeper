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

const (
	keySize = 4096
	keyType = "RSA PRIVATE KEY"
)

// GenRSAKey генерирует и сохраняет RSA-ключ в pem-файл.
func GenRSAKey(ctx context.Context) (_ *rsa.PrivateKey, fileName string, _ error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}

	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, "", err
	}

	bytes := x509.MarshalPKCS1PrivateKey(key)

	fileName = fmt.Sprintf("%0x.pem", bytes[len(bytes)-16:])

	privateKeyPEM := &pem.Block{
		Type:  keyType,
		Bytes: bytes,
	}

	pemFile, err := os.Create(fileName)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = pemFile.Close() }()

	err = pem.Encode(pemFile, privateKeyPEM)
	if err != nil {
		return nil, "", err
	}

	return key, fileName, nil
}

// LoadRSAKey загружает RSA ключ из pem-файла.
func LoadRSAKey(ctx context.Context, keyPath string) (*rsa.PrivateKey, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != keyType {
		return nil, errors.New("wrong block format")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, err
}
