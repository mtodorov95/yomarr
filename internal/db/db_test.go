package db

import (
	"database/sql"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	tmpFile := "test_yomarr.db"
	
	var testDB *sql.DB

	defer func() {
		if testDB != nil {
			testDB.Close()
		}
		os.Remove(tmpFile)
		os.Remove(tmpFile + "-shm")
		os.Remove(tmpFile + "-wal")
	}()

	testDB = Init(tmpFile)
	if testDB == nil {
		t.Fatal("DB handle is nil")
	}

	runMigrations(testDB)

	rows, err := testDB.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='series';")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Error("Table 'series' was not properly created")
	}
}
