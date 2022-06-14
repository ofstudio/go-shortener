package storage

import "sync"

// MemoryStorage - реализация KV-стораджа storage.Interface в памяти
type MemoryStorage struct {
	data map[string]string
	sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
}

func (s *MemoryStorage) Get(key string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	if val, ok := s.data[key]; ok {
		return val, nil
	}
	return "", ErrNotFound
}

func (s *MemoryStorage) Set(key, val string) error {
	s.Lock()
	defer s.Unlock()
	s.data[key] = val
	return nil
}
