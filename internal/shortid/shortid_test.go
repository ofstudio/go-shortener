package shortid

import (
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		got := New()
		require.NotEmpty(t, got)
	})
	t.Run("url friendly", func(t *testing.T) {
		got := New()
		require.Equal(t, got, url.QueryEscape(got))
	})
	t.Run("uniqueness", func(t *testing.T) {
		got1 := New()
		got2 := New()
		require.NotEqual(t, got1, got2)
	})
}
