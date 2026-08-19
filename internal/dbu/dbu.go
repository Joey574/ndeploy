package dbu

import (
	"database/sql"
	"fmt"
	"ndeploy/v2/internal/database"
	"ndeploy/v2/internal/sink"
	"os"
	"path/filepath"
)

const (
	dbName     = "db.sqlite"
	schemaPath = "sql/schema"
	schema     = `CREATE TABLE nodes (
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

func CreateIfNotExists(dir string) error {
	if !DatabaseExists(dir) {
		if err := Init(dir); err != nil {
			return err
		}
	} else {
		sink.Printf(sink.DEBUG, "database found: %s\n", path(dir))
	}

	return nil
}

func ConnectTo(dir string) (*database.Queries, error) {
	sink.Println(sink.TRACE, "connecting to database")
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rw", path(dir)))
	if err != nil {
		return nil, err
	}

	return database.New(db), nil
}

func Init(dir string) error {
	sink.Println(sink.TRACE, "creating database")
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rwc", path(dir)))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(schema)
	return err
}
