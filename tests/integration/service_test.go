package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chzyer/readline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ImpressionableRaccoon/GophKeeper/internal/grpc/keeper"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/keys"
	"github.com/ImpressionableRaccoon/GophKeeper/internal/service"
)

func TestService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := keeper.NewClient(os.Getenv("SERVER_ADDRESS"))
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	key, fileName, err := keys.GenRSAKey(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, fileName)

	s, err := service.New(c, key)
	require.NoError(t, err)

	var entryID string

	t.Run("create entry", func(t *testing.T) {
		b := &bytes.Buffer{}
		b.Write([]byte("name\n"))
		b.Write([]byte("content\n"))

		var l *readline.Instance
		l, err = readline.NewEx(&readline.Config{
			Stdin: io.NopCloser(b),
		})
		require.NoError(t, err)
		defer func() { _ = l.Close() }()

		var resp string
		resp, err = s.Add(ctx, "text", l)
		require.NoError(t, err)

		entryID = strings.TrimSpace(strings.Split(resp, ":")[1])
	})

	t.Run("get entry", func(t *testing.T) {
		var resp string
		resp, err = s.Get(ctx, entryID)
		require.NoError(t, err)

		target := `Type: TextData
Name: name
content`

		assert.Equal(t, target, strings.TrimSpace(resp))
	})

	t.Run("get entry from all entries", func(t *testing.T) {
		var resp string
		resp, err = s.All(ctx)
		require.NoError(t, err)

		target := fmt.Sprintf("%s\tTextData\tname", entryID)

		assert.Equal(t, target, strings.TrimSpace(resp))
	})

	t.Run("update entry", func(t *testing.T) {
		b := &bytes.Buffer{}
		b.Write([]byte("nameUpdated\n"))
		b.Write([]byte("contentUpdated\n"))

		var l *readline.Instance
		l, err = readline.NewEx(&readline.Config{
			Stdin: io.NopCloser(b),
		})
		require.NoError(t, err)
		defer func() { _ = l.Close() }()

		line := fmt.Sprintf("%s text", entryID)

		var resp string
		resp, err = s.Update(ctx, line, l)
		require.NoError(t, err)

		assert.Equal(t, "update ok", strings.TrimSpace(resp))
	})

	t.Run("check if entry really updated", func(t *testing.T) {
		var resp string
		resp, err = s.Get(ctx, entryID)
		require.NoError(t, err)

		target := `Type: TextData
Name: nameUpdated
contentUpdated`

		assert.Equal(t, target, strings.TrimSpace(resp))
	})

	t.Run("delete entry", func(t *testing.T) {
		var resp string
		resp, err = s.Delete(ctx, entryID)
		require.NoError(t, err)

		target := fmt.Sprintf("Entry %s successfully deleted", entryID)

		assert.Equal(t, target, strings.TrimSpace(resp))
	})

	t.Run("check is entry deleted", func(t *testing.T) {
		_, err = s.Get(ctx, entryID)
		require.Error(t, err)
	})

	t.Run("check if no more entries left", func(t *testing.T) {
		var resp string
		resp, err = s.All(ctx)
		require.NoError(t, err)

		assert.Len(t, strings.TrimSpace(resp), 0)
	})

	ctxDone, doneCancel := context.WithCancel(context.Background())
	doneCancel()

	t.Run("get with context done", func(t *testing.T) {
		_, err = s.Get(ctxDone, "")
		require.Error(t, err)
	})

	t.Run("add with context done", func(t *testing.T) {
		_, err = s.Add(ctxDone, "", nil)
		require.Error(t, err)
	})

	t.Run("all with context done", func(t *testing.T) {
		_, err = s.All(ctxDone)
		require.Error(t, err)
	})

	t.Run("delete with context done", func(t *testing.T) {
		_, err = s.Delete(ctxDone, "")
		require.Error(t, err)
	})

	t.Run("update with context done", func(t *testing.T) {
		_, err = s.Update(ctxDone, "", nil)
		require.Error(t, err)
	})

	t.Run("add with wrong type", func(t *testing.T) {
		_, err = s.Add(ctx, "", nil)
		require.Error(t, err)
	})

	t.Run("delete with wrong id", func(t *testing.T) {
		_, err = s.Delete(ctx, "")
		require.Error(t, err)
	})

	t.Run("update with wrong id", func(t *testing.T) {
		_, err = s.Update(ctx, "", nil)
		require.Error(t, err)
	})
}
