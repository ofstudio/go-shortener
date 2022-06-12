package storage

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

// aofItem - структура для хранения JSON-строк в AOF-файле
type aofItem struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

// AOFStorage - реализация KV-стораджа storage.Interface в append-only файле (AOF).
// При вызове AOFStorage.Set данные записываются в файл в виде JSON-объектов, а также сохраняются в памяти.
// При вызове AOFStorage.Get для возврата значения используются данные из памяти.
// При создании стораджа вызовом NewAOFStorage производится попытка загрузки данных из файла в память.
// После завершения работы необходимо закрывать сторадж вызовом AOFStorage.Close.
type AOFStorage struct {
	aof     *os.File
	encoder *json.Encoder
	data    map[string]string
	sync.RWMutex
}

// NewAOFStorage - создает новый сторадж. При создании стораджа производится попытка загрузки данных из файла в память.
// Если файл отсутствует, то происходит создание нового файла.
// После завершения работы необходимо закрывать сторадж вызовом AOFStorage.Close.
func NewAOFStorage(filePath string) (*AOFStorage, error) {
	// Считываем данные из файла в память
	data := make(map[string]string)
	if err := loadDataFromAOF(filePath, data); err != nil {
		return nil, err
	}
	// Открываем файл для записи
	aof, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, ErrAOFOpen
	}

	return &AOFStorage{aof: aof, encoder: json.NewEncoder(aof), data: data}, nil
}

// Get - возвращает значение по ключу
func (s *AOFStorage) Get(key string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	if val, ok := s.data[key]; ok {
		return val, nil
	}
	return "", ErrNotFound
}

// Set - записывает данные в файл и в память
func (s *AOFStorage) Set(key, val string) error {
	s.Lock()
	defer s.Unlock()
	// Сначала сохраняем данные в файл
	item := aofItem{Key: key, Val: val}
	if err := s.encoder.Encode(item); err != nil {
		return ErrAOFWrite
	}
	// Теперь сохраняем данные в память
	s.data[key] = val
	return nil
}

// Close - закрывает сторадж для записи
func (s *AOFStorage) Close() error {
	return s.aof.Close()
}

// loadDataFromAOF - загружает данные из файла в память
func loadDataFromAOF(aofPath string, data map[string]string) error {
	f, err := os.OpenFile(aofPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return ErrAOFOpen
	}
	//goland:noinspection ALL
	defer f.Close()
	decoder := json.NewDecoder(f)
	for {
		var item aofItem
		err = decoder.Decode(&item)
		if err == io.EOF {
			// Конец файла
			break
		} else if err != nil {
			// Ошибка чтения
			return ErrAOFRead
		}
		// Проверяем корректность данных в файле
		if item.Key == "" || item.Val == "" {
			return ErrAOFRead
		}
		data[item.Key] = item.Val
	}
	return nil
}
