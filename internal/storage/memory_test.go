package storage

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemory_Get(t *testing.T) {
	m := NewMemory()
	t.Run("successful read", func(t *testing.T) {
		key, val := "k1", "v1"
		err := m.Set(key, val)
		require.NoError(t, err)
		got, err := m.Get(key)
		require.NoError(t, err)
		require.Equal(t, val, got)
	})
	t.Run("not found", func(t *testing.T) {
		got, err := m.Get("not_found")
		require.Equal(t, ErrNotFound, err)
		require.Empty(t, got)
	})
}

func TestMemory_Set(t *testing.T) {
	m := NewMemory()
	t.Run("successful write", func(t *testing.T) {
		key, val := "k1", "v1"
		err := m.Set(key, val)
		require.NoError(t, err)
		got, err := m.Get(key)
		require.NoError(t, err)
		require.Equal(t, val, got)
	})
}
