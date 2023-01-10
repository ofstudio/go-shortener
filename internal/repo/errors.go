package repo

import "errors"

var (
	ErrInvalidModel     = errors.New("invalid model")            // ErrInvalidModel - невалидная модель данных
	ErrDuplicate        = errors.New("duplicate id")             // ErrDuplicate - дубликат id
	ErrNotFound         = errors.New("not found")                // ErrNotFound - не найдено
	ErrAOFOpen          = errors.New("aof open error")           // ErrAOFOpen - ошибка открытия AOF-файла
	ErrAOFRead          = errors.New("aof read error")           // ErrAOFRead - ошибка чтения AOF-файла
	ErrAOFWrite         = errors.New("aof write error")          // ErrAOFWrite - ошибка записи AOF-файла
	ErrAOFStructure     = errors.New("aof json structure error") // ErrAOFStructure - ошибка структуры JSON в AOF-файле
	ErrDBNotInitialized = errors.New("db not initialized")       // ErrDBNotInitialized - база данных не инициализирована
)
