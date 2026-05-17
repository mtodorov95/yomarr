package db

import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(path string) {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}

	// WAL for concurrency
	_, err = DB.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatal(err)
	}

	createTables()
}

func createTables() {
	schema := `
	CREATE TABLE IF NOT EXISTS series (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		anilist_id TEXT UNIQUE,
		mangadex_id TEXT UNIQUE,
		title TEXT NOT NULL,
		path TEXT NOT NULL,
		status TEXT
	);

	CREATE TABLE IF NOT EXISTS chapters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		series_id INTEGER REFERENCES series(id) ON DELETE CASCADE,
		number REAL NOT NULL,
		volume INTEGER,
		file_path TEXT,
		status TEXT,
		release_date DATETIME
	);

	CREATE TABLE IF NOT EXISTS indexers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		api_key TEXT,
		priority INTEGER DEFAULT 0
	);
	`

	if _, err := DB.Exec(schema); err != nil {
		log.Fatal(err)
	}
}
