package repo

import "errors"

var (
	// ErrInvalidModel - невалидная модель данных
	ErrInvalidModel = errors.New("invalid model")
	// ErrDuplicate - дубликат id
	ErrDuplicate = errors.New("duplicate id")
	// ErrNotFound - не найдено
	ErrNotFound = errors.New("not found")
	// ErrAOFOpen - ошибка открытия AOF-файла
	ErrAOFOpen = errors.New("aof open error")
	// ErrAOFRead - ошибка чтения AOF-файла
	ErrAOFRead = errors.New("aof read error")
	// ErrAOFWrite - ошибка записи AOF-файла
	ErrAOFWrite = errors.New("aof write error")
	// ErrAOFStructure - ошибка структуры JSON в AOF-файле
	ErrAOFStructure = errors.New("aof json structure error")
	// ErrDBNotInitialized - база данных не инициализирована
	ErrDBNotInitialized = errors.New("db not initialized")
)
