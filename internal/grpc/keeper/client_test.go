package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("wrong server address", func(t *testing.T) {
		_, err := NewClient("")
		require.Error(t, err)
	})

	t.Run("close already closed client", func(t *testing.T) {
		c, err := NewClient(":3200")
		require.NoError(t, err)

		err = c.Close()
		require.NoError(t, err)

		err = c.Close()
		require.Error(t, err)
	})
}
