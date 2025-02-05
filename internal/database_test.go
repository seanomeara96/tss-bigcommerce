package internal

import "testing"

func TestDatabaseConnection(t *testing.T) {
	dbURL := "../data/test.db"
	db, err := Database(&dbURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := SaveFileCreation(db, 1, "ch"); err != nil {
		t.Fatal(err)
	}
}
