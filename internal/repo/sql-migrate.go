package repo

// migrate - создает таблицы в базе данных
func (r *SQLRepo) migrate() error {
	_, err := r.db.Exec(`
		-- Создаем таблицу пользователей
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY
		);
		
		-- Создаем таблицу коротких ссылок
		CREATE TABLE IF NOT EXISTS short_urls (
			id TEXT PRIMARY KEY,
			original_url TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT false,
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		);

		-- Создаем уникальный индекс для поля original_url
		CREATE UNIQUE INDEX IF NOT EXISTS short_urls_original_url_idx ON short_urls (original_url);
		
`)

	return err
}
