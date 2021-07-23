package maria

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func New(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	return db, err
}
