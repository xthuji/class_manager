package db

import (
	"fmt"
	"log"
)

func InitSchema() {
	createTables := []string{
		`CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			student_id TEXT UNIQUE,
			contact TEXT,
			subjects TEXT,
			total_hours REAL DEFAULT 0,
			completed_hours REAL DEFAULT 0,
			created_at TEXT,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS courses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			subject TEXT,
			description TEXT,
			created_at TEXT,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS student_course (
			student_id INTEGER,
			course_id INTEGER,
			enrolled_at TEXT,
			PRIMARY KEY (student_id, course_id),
			FOREIGN KEY (student_id) REFERENCES students(id),
			FOREIGN KEY (course_id) REFERENCES courses(id)
		)`,
		`CREATE TABLE IF NOT EXISTS hour_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER,
			course_id INTEGER,
			hours REAL NOT NULL,
			record_date TEXT NOT NULL,
			description TEXT,
			created_at TEXT,
			FOREIGN KEY (student_id) REFERENCES students(id),
			FOREIGN KEY (course_id) REFERENCES courses(id)
		)`,
		`CREATE TABLE IF NOT EXISTS hour_recharges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			hours REAL NOT NULL,
			recharge_date TEXT NOT NULL,
			description TEXT,
			created_at TEXT,
			FOREIGN KEY (student_id) REFERENCES students(id)
		)`,
	}

	for _, query := range createTables {
		_, err := DB.Exec(query)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}

	fmt.Println("Database schema initialized successfully")
}