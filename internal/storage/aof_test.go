package storage

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestAOFStorage(t *testing.T) {
	dir, err := os.MkdirTemp("", "aof_test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Файл не существует
	t.Run("file does not exist", func(t *testing.T) {
		filePath := dir + "/not-exist.aof"
		s, err := NewAOFStorage(filePath)
		require.NoError(t, err)
		defer s.Close()
		require.Equal(t, 0, len(s.data))
	})

	// Недопустимый файл для записи
	t.Run("invalid file path", func(t *testing.T) {
		filePath := dir + "/:/*"
		_, err = NewAOFStorage(filePath)
		require.Error(t, err)
		require.Equal(t, ErrAOFOpen, err)
	})

	// Успешная запись и чтение
	t.Run("successful write and read", func(t *testing.T) {
		filePath := dir + "/shortener.aof"

		// Записываем 2 записи в сторадж s1
		s1, err := NewAOFStorage(filePath)
		require.NoError(t, err)
		defer s1.Close()
		require.NoError(t, s1.Set("key1", "val1"))
		require.NoError(t, s1.Set("key2", "val2"))
		val, err := s1.Get("key1")
		require.NoError(t, err)
		require.Equal(t, "val1", val)
		val, err = s1.Get("key2")
		require.NoError(t, err)
		require.Equal(t, "val2", val)

		// Читаем 2 записи из стораджа s2
		s2, err := NewAOFStorage(filePath)
		require.NoError(t, err)
		defer s2.Close()
		require.Equal(t, 2, len(s2.data))
		val, err = s2.Get("key1")
		require.NoError(t, err)
		require.Equal(t, "val1", val)
		val, err = s2.Get("key2")
		require.NoError(t, err)
		require.Equal(t, "val2", val)
	})

	// Запись не найдена
	t.Run("not found", func(t *testing.T) {
		s, err := NewAOFStorage(dir + "/shortener-nf.aof")
		require.NoError(t, err)
		defer s.Close()
		_, err = s.Get("key1")
		require.Equal(t, ErrNotFound, err)
	})

	// Невалидные строки в файле
	t.Run("invalid aof file", func(t *testing.T) {
		filePath := dir + "/invalid.aof"
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		require.NoError(t, err)
		_, err = f.Write([]byte(`{"invalid": "key", "wrong": "value"}`))
		require.NoError(t, err)
		defer f.Close()

		_, err = NewAOFStorage(filePath)
		require.Error(t, err)
		require.Equal(t, ErrAOFRead, err)
	})

	// Попытка записи в закрытый сторадж
	t.Run("write to closed storage", func(t *testing.T) {
		s, err := NewAOFStorage(dir + "/shortener-closed.aof")
		require.NoError(t, err)
		require.NoError(t, s.Close())
		err = s.Set("key1", "val1")
		require.Error(t, err)
		require.Equal(t, ErrAOFWrite, err)
	})
}
