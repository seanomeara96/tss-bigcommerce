package internal

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func Database(path *string) (*sql.DB, error) {
	defaultPath := "data/main.db"
	if path == nil {
		path = &defaultPath
	}

	db, err := sql.Open("sqlite3", *path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS orders(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL,
		xml_file_created DATETIME NOT NULL,
		website TEXT NOT NULL
	)`); err != nil {
		return nil, err
	}

	return db, nil
}

func SaveFileCreation(db *sql.DB, orderID int, website string) error {
	if _, err := db.Exec(`INSERT INTO ORDERS(order_id, xml_file_created, website) VALUES(?, ?, ?)`, orderID, time.Now().UTC(), website); err != nil {
		return err
	}
	return nil
}
