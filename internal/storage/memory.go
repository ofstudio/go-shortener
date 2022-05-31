package storage

// Memory - реализация KV-стораджа storage.Interface в памяти
type Memory struct {
	data map[string]string
}

func NewMemory() *Memory {
	return &Memory{data: make(map[string]string)}
}

func (m Memory) Get(key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", ErrNotFound
}

func (m Memory) Set(key, val string) error {
	m.data[key] = val
	return nil
}
