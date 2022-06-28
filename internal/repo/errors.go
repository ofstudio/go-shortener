package repo

import "errors"

var (
	ErrInvalidModel = errors.New("invalid model")
	ErrDuplicate    = errors.New("duplicate id")
	ErrNotFound     = errors.New("not found")
	ErrAOFOpen      = errors.New("aof open error")
	ErrAOFRead      = errors.New("aof read error")
	ErrAOFWrite     = errors.New("aof write error")
	ErrAOFStructure = errors.New("aof json structure error")
)
