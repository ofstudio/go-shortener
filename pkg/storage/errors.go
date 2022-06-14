package storage

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrAOFOpen  = errors.New("aof open error")
	ErrAOFRead  = errors.New("aof read error")
	ErrAOFWrite = errors.New("aof write error")
)

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
