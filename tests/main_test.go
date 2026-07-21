package tests

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/class_manager/pkg/db"
)

// TestMain 在所有测试执行前初始化一个隔离的临时 SQLite 数据库。
// 通过覆盖 db.DB，避免污染用户正式数据库。
func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "class_manager_test_*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.sqlite")

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)

	db.DB = database
	db.InitSchema()

	code := m.Run()

	database.Close()
	os.Exit(code)
}

// resetTables 清空所有表数据，确保每个测试用例从干净状态开始。
func resetTables(t *testing.T) {
	t.Helper()
	tables := []string{
		"notifications",
		"schedules",
		"hour_recharges",
		"hour_records",
		"student_course",
		"courses",
		"students",
		"operation_logs",
	}
	for _, tbl := range tables {
		if _, err := db.DB.Exec("DELETE FROM " + tbl); err != nil {
			t.Fatalf("Failed to clear table %s: %v", tbl, err)
		}
	}
}
