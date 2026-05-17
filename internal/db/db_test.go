package db

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	tmpFile := "test_yomarr.db"
	defer func() {
		if DB != nil {
			DB.Close()
		}
		os.Remove(tmpFile)
		os.Remove(tmpFile + "-shm")
		os.Remove(tmpFile + "-wal")
	}()

	Init(tmpFile)
	if DB == nil {
		t.Fatal("DB handle is nil")
	}

	rows, err := DB.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='series';")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Error("Table 'series' not created")
	}
}
