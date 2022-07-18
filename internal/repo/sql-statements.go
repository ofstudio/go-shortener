package repo

import "database/sql"

type (
	stmt       uint8
	statements map[stmt]*sql.Stmt
)

const (
	stmtUserCreate stmt = iota
	stmtUserGetByID
	stmtShortURLCreate
	stmtShortURLGetByID
	stmtShortURLGetByUserID
	stmtShortURLGetByOriginalURL
)

var queries = map[stmt]string{
	stmtUserCreate: `
		INSERT INTO users (id)
		VALUES (DEFAULT)
		RETURNING id
	`,
	stmtUserGetByID: `
		SELECT id FROM users 
	  	WHERE id = $1
	`,
	stmtShortURLCreate: `	
		INSERT INTO short_urls (id, original_url, user_id)
		VALUES ($1, $2, $3)
	`,
	stmtShortURLGetByID: `
		SELECT id, original_url, user_id FROM short_urls 
		WHERE id = $1
	`,
	stmtShortURLGetByUserID: `
		SELECT id, original_url, user_id FROM short_urls 
		WHERE user_id = $1
	`,
	stmtShortURLGetByOriginalURL: `
		SELECT id, original_url, user_id FROM short_urls
		WHERE original_url = $1
	`,
}

// prepareStmts - подготавливает запросы к БД
func prepareStmts(db *sql.DB) (statements, error) {
	stmts := make(statements)
	for id, query := range queries {
		s, err := db.Prepare(query)
		if err != nil {
			return nil, err
		}
		stmts[id] = s
	}
	return stmts, nil
}
