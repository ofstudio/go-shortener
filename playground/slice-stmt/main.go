package main

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"strings"
	"text/tabwriter"
	"time"
)

const dsn = "postgres://playground:playground@localhost:6432/playground"

func main() {
	db := connect()
	defer db.Close()
	migrate(db)
	defer cleanup(db)

	insert(db, 1, "A")
	insert(db, 1, "B")
	insert(db, 1, "C")
	insert(db, 2, "X")
	insert(db, 2, "Y")
	insert(db, 2, "Z")
	fmt.Println("Initial")
	fmt.Println(selectAll(db))

	q := `UPDATE slice_stmt SET deleted = true WHERE user_id = $1 AND payload = ANY($2)`
	stmt, err := db.Prepare(q)
	if err != nil {
		panic(err)
	}

	res, err := stmt.Exec(1, []string{"A", "B"})
	if err != nil {
		panic(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	fmt.Println("After query", q, "with", count, "rows affected")
	fmt.Println(selectAll(db))
}

func insert(db *sql.DB, userID int, payload string) {
	_, err := db.Exec(`INSERT INTO slice_stmt (user_id, payload) VALUES ($1, $2)`, userID, payload)
	if err != nil {
		panic(err)
	}
}

func selectAll(db *sql.DB) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprintf(w, "ID\tUSER_ID\tPAYLOAD\tDELETED\tCREATED_AT\n")
	rows, err := db.Query(`SELECT * FROM slice_stmt ORDER BY created_at ASC`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var userID int
		var payload string
		var deleted bool
		var createdAt time.Time
		err = rows.Scan(&id, &userID, &payload, &deleted, &createdAt)
		if err != nil {
			panic(err)
		}
		_, _ = fmt.Fprintf(w, "%d\t%d\t%s\t%t\t%s\n", id, userID, payload, deleted, createdAt.Format("2006-01-02 15:04:05"))
	}

	_ = w.Flush()
	return b.String()
}

func migrate(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS slice_stmt (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		payload TEXT NOT NULL,
		deleted BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	)
	`)
	if err != nil {
		panic(err)
	}
}

func cleanup(db *sql.DB) {
	_, err := db.Exec(`DROP TABLE IF EXISTS slice_stmt`)
	if err != nil {
		panic(err)
	}
}

func connect() *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}
	return db
}
