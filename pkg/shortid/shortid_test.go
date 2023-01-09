package shortid

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		got := Generate()
		require.NotEmpty(t, got)
	})
	t.Run("url friendly", func(t *testing.T) {
		got := Generate()
		require.Equal(t, got, url.QueryEscape(got))
	})
	t.Run("uniqueness", func(t *testing.T) {
		got1 := Generate()
		got2 := Generate()
		require.NotEqual(t, got1, got2)
	})
}
