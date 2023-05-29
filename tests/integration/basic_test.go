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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ImpressionableRaccoon/GophKeeper/internal/grpc/keeper"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/keys"
	pb "github.com/ImpressionableRaccoon/GophKeeper/proto"
)

var (
	data    = []byte{1, 2, 3, 4, 5}
	newData = []byte{5, 4, 3, 2, 1}
	id      string
)

func TestBasic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := keeper.NewClient(os.Getenv("SERVER_ADDRESS"))
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	key, fileName, err := keys.GenRSAKey(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, fileName)

	t.Run("create entry", func(t *testing.T) {
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

		id = resp.Id
	})

	t.Run("get entry", func(t *testing.T) {
		var resp *pb.GetResponse
		resp, err = c.Get(ctx, &pb.GetRequest{Id: id})
		require.NoError(t, err)

		require.NoError(t, err)
		assert.Equal(t, data, resp.Data)
	})

	t.Run("get entry from all entries", func(t *testing.T) {
		var resp *pb.GetAllResponse
		resp, err = c.GetAll(ctx, &pb.GetAllRequest{PublicKey: key.PublicKey.N.Bytes()})
		require.NoError(t, err)

		assert.Len(t, resp.Entries, 1)
		assert.Equal(t, id, resp.Entries[0].Id)
		assert.Equal(t, data, resp.Entries[0].Data)
	})

	t.Run("update entry", func(t *testing.T) {
		oldHash := sha256.Sum256(data)
		oldHash2 := sha256.Sum256(oldHash[:])
		var signOld []byte
		signOld, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, oldHash2[:])
		require.NoError(t, err)

		newHash := sha256.Sum256(newData)
		var signNew []byte
		signNew, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, newHash[:])
		require.NoError(t, err)

		_, err = c.Update(ctx, &pb.UpdateRequest{
			Id:      id,
			Data:    newData,
			SignOld: signOld,
			SignNew: signNew,
		})
		require.NoError(t, err)
	})

	t.Run("check if entry really updated", func(t *testing.T) {
		var resp *pb.GetResponse
		resp, err = c.Get(ctx, &pb.GetRequest{Id: id})
		require.NoError(t, err)

		require.NoError(t, err)
		assert.Equal(t, newData, resp.Data)
	})

	t.Run("delete entry", func(t *testing.T) {
		hash := sha256.Sum256(newData)
		hash2 := sha256.Sum256(hash[:])
		var sign []byte
		sign, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash2[:])
		require.NoError(t, err)

		_, err = c.Delete(ctx, &pb.DeleteRequest{
			Id:   id,
			Sign: sign,
		})
		require.NoError(t, err)
	})

	t.Run("check is entry deleted", func(t *testing.T) {
		_, err = c.Get(ctx, &pb.GetRequest{Id: id})
		require.Error(t, err)
	})

	t.Run("check if no more entries left", func(t *testing.T) {
		var resp *pb.GetAllResponse
		resp, err = c.GetAll(ctx, &pb.GetAllRequest{PublicKey: key.PublicKey.N.Bytes()})
		require.NoError(t, err)

		assert.Len(t, resp.Entries, 0)
	})
}
