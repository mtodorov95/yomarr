package db

import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
)

func Init(path string) *sql.DB {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}

	// WAL for concurrency
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatal(err)
	}

	createTables(db)
	runMigrations(db)
	return db
}

func createTables(DB *sql.DB) {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS series (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			anilist_id TEXT UNIQUE,
			mangadex_id TEXT UNIQUE,
			title TEXT NOT NULL,
			alt_titles TEXT DEFAULT '[]',
			path TEXT NOT NULL,
			status TEXT,
			monitored INTEGER DEFAULT 1,
            downloading INTEGER DEFAULT 0,
			total_chapters INTEGER DEFAULT 0,
			thumbnail TEXT DEFAULT '',
			historical_covers TEXT DEFAULT '[]',
			author TEXT DEFAULT NULL,
			genres TEXT DEFAULT '[]',
			description TEXT DEFAULT NULL,
			artist TEXT DEFAULT NULL,
			year INTEGER DEFAULT NULL,
			last_chapter TEXT DEFAULT NULL,
			last_volume TEXT DEFAULT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS chapters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			series_id INTEGER REFERENCES series(id) ON DELETE CASCADE,
			number REAL NOT NULL,
			volume INTEGER,
			file_path TEXT,
			status TEXT,
			release_date DATETIME,
			language TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS indexers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			url TEXT NOT NULL,
			api_key TEXT,
			priority INTEGER DEFAULT 0,
			enable_rss INTEGER DEFAULT 1,
			enable_search INTEGER DEFAULT 1,
			additional_parameters TEXT,
			minimum_seeders INTEGER DEFAULT 1,
			seed_time INTEGER DEFAULT 0
		);`,

		`CREATE TABLE IF NOT EXISTS download_clients (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			host TEXT NOT NULL,
			port INTEGER NOT NULL,
			use_ssl INTEGER DEFAULT 0,
			username TEXT,
			password TEXT,
			category TEXT DEFAULT 'yomarr'
		);`,

		`CREATE TABLE IF NOT EXISTS download_queue (
			torrent_hash TEXT PRIMARY KEY,
			series_id INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			release_type TEXT NOT NULL,
			start_num REAL NOT NULL,
			end_num REAL NOT NULL,
			language TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'downloading',
			error_message TEXT DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, tableSchema := range tables {
		if _, err := DB.Exec(tableSchema); err != nil {
			log.Fatalf("[Database] Initialization failed while executing block: %v\nError details: %v", tableSchema, err)
		}
	}
}

func runMigrations(db *sql.DB) {
    _, err := db.Exec(`ALTER TABLE series ADD COLUMN monitored INTEGER DEFAULT 1;`)
    if err != nil {
        log.Printf("[Migration] Note: monitored column might already exist: %v", err)
    }

    _, err = db.Exec(`ALTER TABLE series ADD COLUMN downloading INTEGER DEFAULT 0;`)
    if err != nil {
        log.Printf("[Migration] Note: downloading column might already exist: %v", err)
    }
}
