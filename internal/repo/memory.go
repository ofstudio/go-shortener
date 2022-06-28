package repo

import (
	"github.com/ofstudio/go-shortener/internal/models"
	"sync"
)

// MemoryRepo - реализация Repo для хранения данных в памяти.
type MemoryRepo struct {
	shortURLs     map[string]models.ShortURL
	users         map[uint]models.User
	userShortURLs map[uint][]string
	nextUserID    uint
	sync.RWMutex
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		shortURLs:     make(map[string]models.ShortURL),
		users:         make(map[uint]models.User),
		userShortURLs: make(map[uint][]string),
		nextUserID:    1,
	}
}

// UserCreate - добавляет нового пользователя в репозиторий.
// Если пользователь с таким id уже существует, возвращает ошибку ErrDuplicate.
func (r *MemoryRepo) UserCreate(user *models.User) error {
	r.Lock()
	defer r.Unlock()
	if user == nil {
		return ErrInvalidModel
	}
	autoIncrement(&user.ID, &r.nextUserID)
	if _, exist := r.users[user.ID]; exist {
		return ErrDuplicate
	}
	r.users[user.ID] = *user
	r.userShortURLs[user.ID] = []string{}
	return nil
}

// UserGetByID - возвращает пользователя по его id либо ErrNotFound.
func (r *MemoryRepo) UserGetByID(id uint) (*models.User, error) {
	r.RLock()
	defer r.RUnlock()
	if user, ok := r.users[id]; ok {
		return &user, nil
	}
	return nil, ErrNotFound
}

// ShortURLCreate - создает новую короткую ссылку в репозитории.
// Если короткая ссылка с таким id уже существует, возвращает ErrDuplicate.
func (r *MemoryRepo) ShortURLCreate(shortURL *models.ShortURL) error {
	r.Lock()
	defer r.Unlock()
	if shortURL == nil {
		return ErrInvalidModel
	}
	if _, exist := r.shortURLs[shortURL.ID]; exist {
		return ErrDuplicate
	}
	r.shortURLs[shortURL.ID] = *shortURL
	r.userShortURLs[shortURL.UserID] = append(r.userShortURLs[shortURL.UserID], shortURL.ID)
	return nil
}

// ShortURLGetByID - возвращает короткую ссылку по ее id либо ErrNotFound.
func (r *MemoryRepo) ShortURLGetByID(id string) (*models.ShortURL, error) {
	r.RLock()
	defer r.RUnlock()
	if shortURL, ok := r.shortURLs[id]; ok {
		return &shortURL, nil
	}
	return nil, ErrNotFound
}

// ShortURLGetByUserID - возвращает список коротких ссылок пользователя.
// Если пользователь не существует, возвращает ошибку ErrNotFound.
// Если у пользователя нет коротких ссылок, возвращает пустой слайс.
func (r *MemoryRepo) ShortURLGetByUserID(userID uint) ([]models.ShortURL, error) {
	r.RLock()
	defer r.RUnlock()
	index, ok := r.userShortURLs[userID]
	if !ok {
		return nil, ErrNotFound
	}
	result := make([]models.ShortURL, len(index))
	for i, id := range index {
		result[i] = r.shortURLs[id]
	}
	return result, nil
}

// Close - закрывает репозиторий для записи.
// В этой реализации не делает ничего.
func (r *MemoryRepo) Close() error {
	return nil
}

// userDelete - удаляет пользователя, в тч из индекса ссылок пользователя.
// Вызывается при неудачной попытке создания пользователя в AOFRepo.UserCreate.
func (r *MemoryRepo) userDelete(id uint) {
	r.Lock()
	defer r.Unlock()
	delete(r.users, id)
	delete(r.userShortURLs, id)
}

// shortURLDelete - удаляет короткую ссылку, в тч из индекса ссылок пользователя.
// Вызывается при неудачной попытке создания короткой ссылки в AOFRepo.ShortURLCreate.
func (r *MemoryRepo) shortURLDelete(id string) {
	r.Lock()
	defer r.Unlock()
	if shortURL, exist := r.shortURLs[id]; exist {
		// Удаляем из индекса ссылок пользователя
		userID := shortURL.UserID
		r.userShortURLs[userID] = findAndDelete(r.userShortURLs[userID], shortURL.ID)
		// Удаляем короткую ссылку
		delete(r.shortURLs, id)
	}
}

// autoIncrement - устанавливает значение id и next
// таким образом, чтобы next всегда был больше id.
//
//    - Если id больше next, то next будет установлен в id + 1.
//    - Если id = 0, то id будет установлен в next, а next увеличится на 1.
//    - Если id меньше next, то id и next не изменяются.
//    - Ситуации с id < 0 или next <= 0 не обрабатываются.
//
func autoIncrement(id, next *uint) {
	if id == nil || next == nil || *id < 0 || *next <= 0 {
		return
	}
	switch {
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
