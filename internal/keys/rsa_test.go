package keys

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRSA(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	generatedKey, filePath, err := GenRSAKey(ctx)
	require.NoError(t, err)

	loadedKey, err := LoadRSAKey(ctx, filePath)
	require.NoError(t, err)

	assert.Equal(t, generatedKey, loadedKey)
}

func TestGenRSAKey(t *testing.T) {
	t.Run("context done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, _, err := GenRSAKey(ctx)
		require.Error(t, err)
	})

	t.Run("wrong key size", func(t *testing.T) {
		oldKeySize := keySize
		keySize = 0
		defer func() { keySize = oldKeySize }()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, _, err := GenRSAKey(ctx)
		require.Error(t, err)
	})
}

func TestLoadRSAKey(t *testing.T) {
	t.Run("context done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := LoadRSAKey(ctx, "")
		require.Error(t, err)
	})

	t.Run("wrong file name", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := LoadRSAKey(ctx, "")
		require.Error(t, err)
	})

	t.Run("empty file", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		fileName := uuid.New().String() + ".pem"
		err := os.WriteFile(fileName, []byte{}, 0o600)
		require.NoError(t, err)

		_, err = LoadRSAKey(ctx, fileName)
		require.Error(t, err)
	})

	t.Run("wrong key in correct pem file", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		content := `-----BEGIN RSA PRIVATE KEY-----
dGVzdA==
-----END RSA PRIVATE KEY-----
`

		fileName := uuid.New().String() + ".pem"
		err := os.WriteFile(fileName, []byte(content), 0o600)
		require.NoError(t, err)

		_, err = LoadRSAKey(ctx, fileName)
		require.Error(t, err)
	})
}
