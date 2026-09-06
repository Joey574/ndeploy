package dbu

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"ndeploy/v2/internal/db"
	"os"
	"strings"

	"github.com/Joey574/sink/v2/pkg/sink"
)

type Dbu struct {
	path string
	sink *sink.Sink
	db   *sql.DB
}

func New(path string, options ...func(*Dbu)) *Dbu {
	dbu := &Dbu{
		path: path,
		sink: sink.New(sink.EnableStdOut()),
	}

	for _, o := range options {
		o(dbu)
	}

	return dbu
}

func (d *Dbu) DatabaseExists() bool {
	if _, err := os.Open(d.path); err != nil {
		return false
	}

	return true
}

func (d *Dbu) CreateIfNotExists(schema embed.FS) error {
	if !d.DatabaseExists() {
		if err := d.Create(schema); err != nil {
			return err
		}
	} else {
		d.sink.Printf(sink.DEBUG, "database found: %s\n", d.path)
	}

	return nil
}

func (d *Dbu) Connect() error {
	if d.db != nil {
		return nil
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rw", d.path))
	if err != nil {
		return err
	}

	d.db = db
	return nil
}

func (d *Dbu) Create(schema embed.FS) error {
	d.sink.Println(sink.TRACE, "creating database")
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=rwc", d.path))
	if err != nil {
		return err
	}
	defer db.Close()

	err = fs.WalkDir(schema, "sql/schema", func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}

		if !strings.HasSuffix(entry.Name(), ".sql") {
			return nil
		}

		bytes, err := fs.ReadFile(schema, fmt.Sprintf("sql/schema/%s", entry.Name()))
		if err != nil {
			d.sink.Printf(sink.ERROR, "%v\n", err)
			return err
		}

		_, err = db.Exec(string(bytes))
		if err != nil {
			d.sink.Printf(sink.ERROR, "%v\n", err)
			return err
		}

		return nil
	})

	return err
}

func (d *Dbu) Queries() *db.Queries {
	return db.New(d.db)
}
