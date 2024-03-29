package repo

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/ofstudio/go-shortener/internal/models"
)

// aofRecord - структура одной JSON-записи для хранения в AOF-файле.
type aofRecord struct {
	UserCreate     *models.User     `json:"user_create,omitempty"`
	ShortURLCreate *models.ShortURL `json:"short_url_create,omitempty"`
	ShortURLDelete *models.ShortURL `json:"short_url_update,omitempty"`
}

// AOFRepo - реализация IRepo для хранения данных в append-only файле (AOF).
// При создании репозитория производится загрузка данных из файла в память.
// При чтении из репозитория используются данные из памяти.
// При записи в репозиторий, данные сохраняются в память, а также записываются в файл в виде JSON-строк.
// После завершения работы необходимо закрывать репозиторий AOFRepo.Close.
type AOFRepo struct {
	aof     *os.File
	encoder *json.Encoder
	*MemoryRepo
	mu sync.Mutex
}

// NewAOFRepo - конструктор репозитория AOFRepo.
func NewAOFRepo(filePath string) (*AOFRepo, error) {
	// Считываем данные из файла в память
	memoryRepo, err := loadRepoFromFile(filePath)
	if err != nil {
		return nil, err
	}
	// Открываем файл для записи
	aof, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, ErrAOFOpen
	}
	return &AOFRepo{aof: aof, encoder: json.NewEncoder(aof), MemoryRepo: memoryRepo}, nil
}

// UserCreate - добавляет нового пользователя в репозиторий.
// Если пользователь с таким id уже существует, возвращает ErrDuplicate.
// При ошибке записи в файл, возвращает ErrAOFWrite.
func (r *AOFRepo) UserCreate(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.MemoryRepo.UserCreate(ctx, user); err != nil {
		return err
	}
	if err := r.encoder.Encode(aofRecord{UserCreate: user}); err != nil {
		r.MemoryRepo.userPurge(user.ID)
		return ErrAOFWrite
	}
	return nil
}

// ShortURLCreate - создает новую короткую ссылку в репозитории.
// Если короткая ссылка с таким id уже существует, возвращает ErrDuplicate.
// При ошибке записи в файл, возвращает ErrAOFWrite.
func (r *AOFRepo) ShortURLCreate(ctx context.Context, shortURL *models.ShortURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.MemoryRepo.ShortURLCreate(ctx, shortURL); err != nil {
		return err
	}
	if err := r.encoder.Encode(aofRecord{ShortURLCreate: shortURL}); err != nil {
		r.MemoryRepo.shortURLPurge(shortURL.ID)
		return ErrAOFWrite
	}
	return nil
}

// ShortURLDelete - помечает удаленной короткую ссылку пользователя по ее id.
func (r *AOFRepo) ShortURLDelete(ctx context.Context, userID uint, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.MemoryRepo.ShortURLDelete(ctx, userID, id); err != nil {
		return err
	}
	if err := r.encoder.Encode(aofRecord{ShortURLDelete: &models.ShortURL{ID: id, UserID: userID}}); err != nil {
		r.MemoryRepo.shortURLRestore(id)
		return ErrAOFWrite
	}
	return nil
}

// ShortURLDeleteBatch - помечает удаленными несколько сокращенных ссылок пользователя по их id.
// Принимает на вход список каналов для передачи идентификаторов.
// Возвращает количество удаленных сокращенных ссылок.
func (r *AOFRepo) ShortURLDeleteBatch(ctx context.Context, userID uint, chans ...chan string) (int64, error) {
	// Мультиплексируем каналы chans в один канал ch.
	ch := fanIn(ctx, chans...)
	n := 0
	// Читаем значения из канала и помечаем ссылки как удаленные
loop:
	for {
		select {
		// Если контекст завершился, выходим из цикла.
		case <-ctx.Done():
			break loop
		case id, ok := <-ch:
			// Если канал закрыт, выходим из цикла.
			if !ok {
				break loop
			}
			// Помечаем ссылку как удаленную.
			if err := r.ShortURLDelete(ctx, userID, id); err != nil {
				continue
			}
			// Если ссылка успешно помечена как удаленная, увеличиваем счетчик.
			n++
		}
	}
	return int64(n), nil
}

// Close - закрывает репозиторий для записи.
func (r *AOFRepo) Close() error {
	return r.aof.Close()
}

// loadRepoFromFile - считывает данные из файла в MemoryRepo.
// При ошибке чтения или парсинга JSON возвращает ErrAOFRead.
// При несоответствии структуры данных возвращает ErrAOFStructure.
func loadRepoFromFile(aofPath string) (*MemoryRepo, error) {
	f, err := os.OpenFile(aofPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, ErrAOFOpen
	}
	//goland:noinspection ALL
	defer f.Close()
	repo := NewMemoryRepo()
	decoder := json.NewDecoder(f)

	for {
		record := &aofRecord{}
		err = decoder.Decode(record)
		if errors.Is(err, io.EOF) { // Конец файла
			break
		} else if err != nil { // Ошибка чтения
			return nil, ErrAOFRead
		}
		if err = loadRecord(record, repo); err != nil {
			return nil, err
		}
	}
	return repo, nil
}

// loadRecord - загружает одну JSON-запись aofRecord в MemoryRepo.
// При несоответствии структуры данных возвращает ErrAOFStructure.
func loadRecord(r *aofRecord, repo *MemoryRepo) error {
	switch {
	case r.UserCreate != nil:
		if err := repo.UserCreate(context.Background(), r.UserCreate); err != nil {
			return err
		}
	case r.ShortURLCreate != nil:
		if err := repo.ShortURLCreate(context.Background(), r.ShortURLCreate); err != nil {
			return err
		}
	case r.ShortURLDelete != nil:
		if err := repo.ShortURLDelete(context.Background(), r.ShortURLDelete.UserID, r.ShortURLDelete.ID); err != nil {
			return err
		}
	default:
		return ErrAOFStructure
	}
	return nil
}
