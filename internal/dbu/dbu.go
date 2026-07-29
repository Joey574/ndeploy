package dbu

import (
	"os"
	"path/filepath"
)

const (
	dbName = "db.sqlite"
)

func DatabaseExists(dir string) bool {
	path := filepath.Join(dir, dbName)
	if _, err := os.Open(path); err != nil {
		return false
	}

	return true
}

func ConnectToDb(dir string) {

}

func InitDb(dir string) error {
	return nil
}
