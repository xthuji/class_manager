package backup

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/class_manager/pkg/db"
)

type BackupFile struct {
	Filename    string    `json:"filename"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	Checksum    string    `json:"checksum"`
}

func GetBackupDir() (string, error) {
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDataDir, "ClassManager", "backups"), nil
}

func PerformBackup() (string, error) {
	dbPath, err := db.GetDBPath()
	if err != nil {
		return "", err
	}

	backupDir, err := GetBackupDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("backup_%s.zip", timestamp)
	backupFilePath := filepath.Join(backupDir, backupFileName)

	dbFile, err := os.Open(dbPath)
	if err != nil {
		return "", err
	}
	defer dbFile.Close()

	zipFile, err := os.Create(backupFilePath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fileInfo, _ := dbFile.Stat()
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return "", err
	}
	header.Name = "class_manager.sqlite"
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(writer, dbFile)
	if err != nil {
		return "", err
	}

	return backupFilePath, nil
}

func RestoreFromBackup(backupPath string) error {
	dbPath, err := db.GetDBPath()
	if err != nil {
		return err
	}

	zipFile, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipReader, err := zip.NewReader(zipFile, 0)
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		if file.Name == "class_manager.sqlite" {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			dbDir := filepath.Dir(dbPath)
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return err
			}

			dbFile, err := os.Create(dbPath)
			if err != nil {
				return err
			}
			defer dbFile.Close()

			_, err = io.Copy(dbFile, rc)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("database file not found in backup")
}

func GetBackupList() ([]BackupFile, error) {
	backupDir, err := GetBackupDir()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupFile{}, nil
		}
		return nil, err
	}

	var backupFiles []BackupFile

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".zip" {
			continue
		}

		filePath := filepath.Join(backupDir, file.Name())
		fileInfo, err := file.Info()
		if err != nil {
			continue
		}

		checksum, _ := calculateChecksum(filePath)

		backupFiles = append(backupFiles, BackupFile{
			Filename:  file.Name(),
			Path:      filePath,
			Size:      fileInfo.Size(),
			CreatedAt: fileInfo.ModTime(),
			Checksum:  checksum,
		})
	}

	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[j].CreatedAt.Before(backupFiles[i].CreatedAt)
	})

	return backupFiles, nil
}

func CleanupOldBackups(keepDays int) error {
	backupDir, err := GetBackupDir()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cutoff := time.Now().Add(-time.Duration(keepDays) * 24 * time.Hour)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(backupDir, file.Name())
		fileInfo, err := file.Info()
		if err != nil {
			continue
		}

		if fileInfo.ModTime().Before(cutoff) {
			os.Remove(filePath)
		}
	}

	return nil
}

func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
