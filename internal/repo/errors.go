package repo

import "errors"

// ErrInvalidModel - невалидная модель данных
var ErrInvalidModel = errors.New("invalid model")

// ErrDuplicate - дубликат id
var ErrDuplicate = errors.New("duplicate id")

// ErrNotFound - не найдено
var ErrNotFound = errors.New("not found")

// ErrAOFOpen - ошибка открытия AOF-файла
var ErrAOFOpen = errors.New("aof open error")

// ErrAOFRead - ошибка чтения AOF-файла
var ErrAOFRead = errors.New("aof read error")

// ErrAOFWrite - ошибка записи AOF-файла
var ErrAOFWrite = errors.New("aof write error")

// ErrAOFStructure - ошибка структуры JSON в AOF-файле
var ErrAOFStructure = errors.New("aof json structure error")

// ErrDBNotInitialized - база данных не инициализирована
var ErrDBNotInitialized = errors.New("db not initialized")
