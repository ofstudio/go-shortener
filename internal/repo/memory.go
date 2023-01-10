package repo

import (
	"context"
	"sync"

	"github.com/ofstudio/go-shortener/internal/models"
)

// MemoryRepo - реализация IRepo для хранения данных в памяти.
type MemoryRepo struct {
	shortURLs      map[string]*models.ShortURL
	users          map[uint]*models.User
	userShortURLs  map[uint][]string
	originalURLIdx map[string]string
	nextUserID     uint
	mu             sync.RWMutex
}

// NewMemoryRepo - конструктор MemoryRepo.
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		shortURLs:      make(map[string]*models.ShortURL),
		users:          make(map[uint]*models.User),
		userShortURLs:  make(map[uint][]string),
		originalURLIdx: make(map[string]string),
		nextUserID:     1,
	}
}

// UserCreate - добавляет нового пользователя в репозиторий.
// Если пользователь с таким id уже существует, возвращает ошибку ErrDuplicate.
func (r *MemoryRepo) UserCreate(_ context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if user == nil {
		return ErrInvalidModel
	}
	autoIncrement(&user.ID, &r.nextUserID)
	if _, exist := r.users[user.ID]; exist {
		return ErrDuplicate
	}
	r.users[user.ID] = user
	return nil
}

// UserGetByID - возвращает пользователя по его id либо ErrNotFound.
func (r *MemoryRepo) UserGetByID(_ context.Context, id uint) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, ErrNotFound
}

// ShortURLCreate - создает новую короткую ссылку в репозитории.
// Если короткая ссылка с таким id уже существует, возвращает ErrDuplicate.
func (r *MemoryRepo) ShortURLCreate(_ context.Context, shortURL *models.ShortURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if shortURL == nil {
		return ErrInvalidModel
	}
	if _, exist := r.shortURLs[shortURL.ID]; exist {
		return ErrDuplicate
	}
	if _, exist := r.originalURLIdx[shortURL.OriginalURL]; exist {
		return ErrDuplicate
	}
	r.shortURLs[shortURL.ID] = shortURL
	r.userShortURLs[shortURL.UserID] = append(r.userShortURLs[shortURL.UserID], shortURL.ID)
	r.originalURLIdx[shortURL.OriginalURL] = shortURL.ID
	return nil
}

// ShortURLGetByID - возвращает короткую ссылку по ее id либо ErrNotFound.
func (r *MemoryRepo) ShortURLGetByID(_ context.Context, id string) (*models.ShortURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if shortURL, ok := r.shortURLs[id]; ok {
		return shortURL, nil
	}
	return nil, ErrNotFound
}

// ShortURLGetByUserID - возвращает список коротких ссылок пользователя.
// Если пользователь не найден, или у пользователя нет ссылок возвращает nil.
func (r *MemoryRepo) ShortURLGetByUserID(_ context.Context, userID uint) ([]models.ShortURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	index, ok := r.userShortURLs[userID]
	if !ok {
		return nil, nil
	}
	result := make([]models.ShortURL, 0, len(index))
	for _, id := range index {
		result = append(result, *r.shortURLs[id])
	}
	return result, nil
}

// ShortURLGetByOriginalURL - возвращает сокращенную ссылку по ее оригинальному url.
func (r *MemoryRepo) ShortURLGetByOriginalURL(_ context.Context, originalURL string) (*models.ShortURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if id, ok := r.originalURLIdx[originalURL]; ok {
		if shortURL, ok := r.shortURLs[id]; ok {
			return shortURL, nil
		}
	}
	return nil, ErrNotFound
}

// ShortURLDeleteBatch - помечает удаленными несколько сокращенных ссылок пользователя по их id.
// Принимает на вход список каналов для передачи идентификаторов.
// Возвращает количество удаленных сокращенных ссылок.
func (r *MemoryRepo) ShortURLDeleteBatch(ctx context.Context, userID uint, chans ...chan string) (int64, error) {
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

// ShortURLDelete - помечает удаленной короткую ссылку пользователя по ее id.
func (r *MemoryRepo) ShortURLDelete(_ context.Context, userID uint, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	shortURL, exist := r.shortURLs[id]
	if !exist {
		return ErrNotFound
	}
	if shortURL.UserID != userID {
		return ErrNotFound
	}
	shortURL.Deleted = true
	r.shortURLs[id] = shortURL
	return nil
}

// Close - закрывает репозиторий для записи.
// В этой реализации не делает ничего.
func (r *MemoryRepo) Close() error {
	return nil
}

// userPurge - удаляет пользователя, в тч из индекса ссылок пользователя.
// Вызывается при неудачной попытке создания пользователя в AOFRepo.UserCreate.
func (r *MemoryRepo) userPurge(id uint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	delete(r.userShortURLs, id)
}

// shortURLPurge - удаляет короткую ссылку, в тч из индекса ссылок пользователя.
// Вызывается при неудачной попытке создания короткой ссылки в AOFRepo.ShortURLCreate.
func (r *MemoryRepo) shortURLPurge(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if shortURL, exist := r.shortURLs[id]; exist {
		// Удаляем из индекса ссылок пользователя
		userID := shortURL.UserID
		r.userShortURLs[userID] = findAndDelete(r.userShortURLs[userID], shortURL.ID)
		// Удаляем короткую ссылку
		delete(r.originalURLIdx, shortURL.OriginalURL)
		delete(r.shortURLs, id)
	}
}

// shortURLRestore - восстанавливает короткую ссылку.
// Вызывается при неудачной попытке создания короткой ссылки в AOFRepo.ShortURLDeleteBatch.
func (r *MemoryRepo) shortURLRestore(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if shortURL, exist := r.shortURLs[id]; exist {
		shortURL.Deleted = false
		r.shortURLs[id] = shortURL
	}
}

// autoIncrement - устанавливает значение id и next
// таким образом, чтобы next всегда был больше id.
//
//   - Если id больше next, то next будет установлен в id + 1.
//   - Если id = 0, то id будет установлен в next, а next увеличится на 1.
//   - Если id меньше next, то id и next не изменяются.
//   - Ситуации с next == 0 не обрабатываются.
func autoIncrement(id, next *uint) {
	switch {
	case id == nil || next == nil:
		return
	case *next == 0:
		return
	case *id >= *next:
		*next = *id + 1
	case *id == 0:
		*id = *next
		*next++
	}
}

// findAndDelete - ищет значение в слайсе и удаляет его.
func findAndDelete(slice []string, item string) []string {
	if slice == nil {
		return nil
	}
	i := 0
	for _, current := range slice {
		if current != item {
			slice[i] = current
			i++
		}
	}
	return slice[:i]
}
