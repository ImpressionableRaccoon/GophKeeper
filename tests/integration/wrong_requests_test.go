package integration

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ImpressionableRaccoon/GophKeeper/internal/grpc/keeper"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/keys"
	pb "github.com/ImpressionableRaccoon/GophKeeper/proto"
)

func TestWrongRequests(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := keeper.NewClient(os.Getenv("SERVER_ADDRESS"))
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	key, fileName, err := keys.GenRSAKey(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, fileName)

	t.Run("get with wrong id", func(t *testing.T) {
		_, err = c.Get(ctx, &pb.GetRequest{Id: "dasdasda"})
		require.Error(t, err)
	})

	t.Run("create with wrong sign", func(t *testing.T) {
		_, err = c.Create(ctx, &pb.CreateRequest{
			PublicKey: []byte{1, 2, 3},
			Data:      []byte{1, 2, 3},
			Sign:      []byte{1, 2, 3},
		})
	})

	t.Run("delete with wrong id", func(t *testing.T) {
		_, err = c.Delete(ctx, &pb.DeleteRequest{Id: "dasdasda"})
		require.Error(t, err)
	})

	t.Run("delete with not found id", func(t *testing.T) {
		_, err = c.Delete(ctx, &pb.DeleteRequest{Id: uuid.New().String()})
		require.Error(t, err)
	})

	t.Run("delete with wrong sign", func(t *testing.T) {
		hash := sha256.Sum256(data)

		var sign []byte
		sign, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
		require.NoError(t, err)

		var resp *pb.CreateResponse
		resp, err = c.Create(ctx, &pb.CreateRequest{
			PublicKey: key.PublicKey.N.Bytes(),
			Data:      data,
			Sign:      sign,
		})
		require.NoError(t, err)

		_, err = c.Delete(ctx, &pb.DeleteRequest{
			Id:   resp.Id,
			Sign: []byte{1, 2, 3},
		})
		require.Error(t, err)
	})

	t.Run("update with wrong id", func(t *testing.T) {
		_, err = c.Update(ctx, &pb.UpdateRequest{Id: "dasdasda"})
		require.Error(t, err)
	})

	t.Run("update with not found id", func(t *testing.T) {
		_, err = c.Update(ctx, &pb.UpdateRequest{Id: uuid.New().String()})
		require.Error(t, err)
	})

	t.Run("update with wrong old sign", func(t *testing.T) {
		hash := sha256.Sum256(data)

		var sign []byte
		sign, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
		require.NoError(t, err)

		var resp *pb.CreateResponse
		resp, err = c.Create(ctx, &pb.CreateRequest{
			PublicKey: key.PublicKey.N.Bytes(),
			Data:      data,
			Sign:      sign,
		})
		require.NoError(t, err)

		_, err = c.Update(ctx, &pb.UpdateRequest{
			Id:      resp.Id,
			SignOld: []byte{1, 2, 3},
		})
		require.Error(t, err)
	})

	t.Run("update with wrong new sign", func(t *testing.T) {
		hash := sha256.Sum256(data)

		var sign []byte
		sign, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
		require.NoError(t, err)

		var resp *pb.CreateResponse
		resp, err = c.Create(ctx, &pb.CreateRequest{
			PublicKey: key.PublicKey.N.Bytes(),
			Data:      data,
			Sign:      sign,
		})
		require.NoError(t, err)

		hash2 := sha256.Sum256(hash[:])

		var signOld []byte
		signOld, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash2[:])
		require.NoError(t, err)

		_, err = c.Update(ctx, &pb.UpdateRequest{
			Id:      resp.Id,
			Data:    newData,
			SignOld: signOld,
			SignNew: []byte{1, 2, 3},
		})
		require.Error(t, err)
	})
}
