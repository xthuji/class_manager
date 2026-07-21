package db

import (
	"log"
	"strings"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS schedules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		course_id INTEGER,
		day_of_week INTEGER,
		period TEXT,
		start_time TEXT,
		end_time TEXT,
		hours_consumed REAL,
		created_at TEXT,
		updated_at TEXT,
		FOREIGN KEY (course_id) REFERENCES courses(id)
	)`,
	`CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER,
		course_name TEXT,
		current_hours REAL,
		threshold REAL,
		status TEXT DEFAULT 'unread',
		created_at TEXT,
		FOREIGN KEY (student_id) REFERENCES students(id)
	)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_courses_name ON courses(name)`,
	// 为已有数据库添加 threshold 列（新数据库已通过 schema 创建）
	`ALTER TABLE courses ADD COLUMN threshold REAL DEFAULT 2`,
	// 去除科目概念：删除 threshold 表（阈值已迁移到课程级别）
	`DROP TABLE IF EXISTS threshold`,
	// 去除科目概念：删除 students.subjects 列
	`ALTER TABLE students DROP COLUMN subjects`,
	// 去除科目概念：删除 courses.subject 列
	`ALTER TABLE courses DROP COLUMN subject`,
	// 通知表 subject 列重命名为 course_name
	`ALTER TABLE notifications RENAME COLUMN subject TO course_name`,
	// 通知简化为未读/已读：删除 processed_at 列
	`ALTER TABLE notifications DROP COLUMN processed_at`,
	// 操作日志表（数据库存储用户操作记录）
	`CREATE TABLE IF NOT EXISTS operation_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		operation_type TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id INTEGER,
		description TEXT,
		created_at TEXT
	)`,
	// 添加学生退学状态字段
	`ALTER TABLE students ADD COLUMN is_dropped INTEGER DEFAULT 0`,
	// 通知表添加 type 列，区分通知类型
	`ALTER TABLE notifications ADD COLUMN type TEXT DEFAULT 'threshold'`,
}

func runMigrations() error {
	for i, migration := range migrations {
		_, err := DB.Exec(migration)
		if err != nil {
			errMsg := err.Error()
			// 这些错误是正常的，跳过即可：
			// - duplicate column: ALTER TABLE ADD COLUMN 列已存在
			// - no such column: ALTER TABLE DROP/RENAME 列已不存在
			// - no such table: DROP TABLE 表已不存在
			if strings.Contains(errMsg, "duplicate column") ||
				strings.Contains(errMsg, "no such column") ||
				strings.Contains(errMsg, "no such table") {
				log.Printf("Migration %d skipped (already applied)", i+1)
			} else {
				log.Printf("Warning: migration %d skipped: %v", i+1, err)
			}
		} else {
			log.Printf("Migration %d executed successfully", i+1)
		}
	}
	return nil
}
