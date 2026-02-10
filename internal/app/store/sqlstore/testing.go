package sqlstore

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

func TestDB(t *testing.T, databaseURL string) (*sql.DB, func(...string)) {
	t.Helper()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	return db, func(tables ...string) {
		if len(tables) > 0 {
			query := fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", "))
			if _, err := db.Exec(query); err != nil {
				t.Fatal(err)
			}
		}

		db.Close()
	}
}
