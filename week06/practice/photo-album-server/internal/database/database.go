package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		avatar_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	`CREATE TABLE IF NOT EXISTS albums (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		is_public INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`,
	`CREATE TABLE IF NOT EXISTS photos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		album_id INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		file_size INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (album_id) REFERENCES albums(id)
	);`,
}

func New(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath+"?_foreign_keys=on"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	for _, statement := range schemaStatements {
		if err := db.Exec(statement).Error; err != nil {
			return nil, err
		}
	}

	return db, nil
}
