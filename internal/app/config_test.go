package app

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_normalizeURL(t *testing.T) {
	t.Run("add slash", func(t *testing.T) {
		got := normalizeURL("https://example.com")
		require.Equal(t, "https://example.com/", got)
	})
	t.Run("no slash", func(t *testing.T) {
		got := normalizeURL("https://example.com/")
		require.Equal(t, "https://example.com/", got)
	})
}
