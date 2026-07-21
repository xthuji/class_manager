package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type ExportData struct {
	Students     []map[string]interface{} `json:"students"`
	Courses      []map[string]interface{} `json:"courses"`
	HourRecords  []map[string]interface{} `json:"hour_records"`
	HourRecharges []map[string]interface{} `json:"hour_recharges"`
	Schedules    []map[string]interface{} `json:"schedules"`
	Threshold    []map[string]interface{} `json:"threshold"`
	Notifications []map[string]interface{} `json:"notifications"`
}

var DB *sql.DB

func InitDB() error {
	if customPath := os.Getenv("CLASS_MANAGER_DB_PATH"); customPath != "" {
		dbPath := customPath
		return initDBWithPath(dbPath)
	}

	appDataDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Failed to get user config dir, using temp directory: %v", err)
		appDataDir = os.TempDir()
	}

	dbDir := filepath.Join(appDataDir, "ClassManager")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Printf("Failed to create database directory, using temp directory: %v", err)
		dbDir = os.TempDir()
	}

	dbPath := filepath.Join(dbDir, "class_manager.sqlite")
	return initDBWithPath(dbPath)
}

func initDBWithPath(dbPath string) error {
	var exists bool
	if _, err := os.Stat(dbPath); err == nil {
		exists = true
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	DB = db

	if !exists {
		InitSchema()
	}

	if err := runMigrations(); err != nil {
		return err
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

func GetDBPath() (string, error) {
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDataDir, "ClassManager", "class_manager.sqlite"), nil
}

func GetTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

func ExportAllData() (string, error) {
	data := ExportData{}

	tables := []struct {
		name   string
		result *[]map[string]interface{}
	}{
		{"students", &data.Students},
		{"courses", &data.Courses},
		{"hour_records", &data.HourRecords},
		{"hour_recharges", &data.HourRecharges},
		{"schedules", &data.Schedules},
		{"threshold", &data.Threshold},
		{"notifications", &data.Notifications},
	}

	for _, table := range tables {
		rows, err := DB.Query(`SELECT * FROM ` + table.name)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return "", err
		}

		for rows.Next() {
			values := make([]interface{}, len(columns))
			scanArgs := make([]interface{}, len(columns))
			for i := range values {
				scanArgs[i] = &values[i]
			}

			if err := rows.Scan(scanArgs...); err != nil {
				return "", err
			}

			record := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					record[col] = string(b)
				} else {
					record[col] = val
				}
			}
			*table.result = append(*table.result, record)
		}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func ImportAllData(jsonData string) error {
	var data ExportData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return err
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM notifications`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM threshold`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM schedules`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM hour_recharges`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM hour_records`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM student_course`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM courses`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM students`)
	if err != nil {
		return err
	}

	if len(data.Students) > 0 {
		columns := getKeys(data.Students[0])
		placeholders := make([]string, len(columns))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		query := `INSERT INTO students (` + joinWithComma(columns) + `) VALUES (` + joinWithComma(placeholders) + `)`

		for _, record := range data.Students {
			values := getValues(record, columns)
			_, err := tx.Exec(query, values...)
			if err != nil {
				return err
			}
		}
	}

	if len(data.Courses) > 0 {
		columns := getKeys(data.Courses[0])
		placeholders := make([]string, len(columns))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		query := `INSERT INTO courses (` + joinWithComma(columns) + `) VALUES (` + joinWithComma(placeholders) + `)`

		for _, record := range data.Courses {
			values := getValues(record, columns)
			_, err := tx.Exec(query, values...)
			if err != nil {
				return err
			}
		}
	}

	if len(data.HourRecords) > 0 {
		columns := getKeys(data.HourRecords[0])
		placeholders := make([]string, len(columns))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		query := `INSERT INTO hour_records (` + joinWithComma(columns) + `) VALUES (` + joinWithComma(placeholders) + `)`

		for _, record := range data.HourRecords {
			values := getValues(record, columns)
			_, err := tx.Exec(query, values...)
			if err != nil {
				return err
			}
		}
	}

	if len(data.HourRecharges) > 0 {
		columns := getKeys(data.HourRecharges[0])
		placeholders := make([]string, len(columns))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		query := `INSERT INTO hour_recharges (` + joinWithComma(columns) + `) VALUES (` + joinWithComma(placeholders) + `)`

		for _, record := range data.HourRecharges {
			values := getValues(record, columns)
			_, err := tx.Exec(query, values...)
			if err != nil {
				return err
			}
		}
	}

	if len(data.Schedules) > 0 {
		columns := getKeys(data.Schedules[0])
		placeholders := make([]string, len(columns))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		query := `INSERT INTO schedules (` + joinWithComma(columns) + `) VALUES (` + joinWithComma(placeholders) + `)`

		for _, record := range data.Schedules {
			values := getValues(record, columns)
			_, err := tx.Exec(query, values...)
			if err != nil {
				return err
			}
		}
	}

	if len(data.Threshold) > 0 {
		columns := getKeys(data.Threshold[0])
		placeholders := make([]string, len(columns))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		query := `INSERT INTO threshold (` + joinWithComma(columns) + `) VALUES (` + joinWithComma(placeholders) + `)`

		for _, record := range data.Threshold {
			values := getValues(record, columns)
			_, err := tx.Exec(query, values...)
			if err != nil {
				return err
			}
		}
	}

	if len(data.Notifications) > 0 {
		columns := getKeys(data.Notifications[0])
		placeholders := make([]string, len(columns))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		query := `INSERT INTO notifications (` + joinWithComma(columns) + `) VALUES (` + joinWithComma(placeholders) + `)`

		for _, record := range data.Notifications {
			values := getValues(record, columns)
			_, err := tx.Exec(query, values...)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func joinWithComma(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}

func getValues(m map[string]interface{}, keys []string) []interface{} {
	values := make([]interface{}, len(keys))
	for i, k := range keys {
		values[i] = m[k]
	}
	return values
}
