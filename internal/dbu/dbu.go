package dbu

import (
	"database/sql"
	"fmt"
	"ndeploy/v2/internal/database"
	"os"
	"path/filepath"
)

const (
	dbName = "db.sqlite"
	schema = `CREATE TABLE nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user TEXT NOT NULL,
    host TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    `
)

var (
	_path string
)

func path(dir string) string {
	if _path == "" {
		_path = filepath.Join(dir, dbName)
	}

	return _path
}

func DatabaseExists(dir string) bool {
	if _, err := os.Open(path(dir)); err != nil {
		return false
	}

	return true
}

func ConnectTo(dir string) (*database.Queries, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rw", path(dir)))
	if err != nil {
		return nil, err
	}

	return database.New(db), nil

}

func InitDb(dir string) error {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rwc", path(dir)))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(schema)
	return err
}
