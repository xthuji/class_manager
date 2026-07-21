package db

import "log"

var migrations = []string{}

func runMigrations() error {
	for i, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			log.Printf("Failed to run migration %d: %v", i+1, err)
			return err
		}
		log.Printf("Migration %d executed successfully", i+1)
	}
	return nil
}
