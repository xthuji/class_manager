package scheduler

import (
	"log"

	"github.com/class_manager/pkg/backup"
	"github.com/class_manager/pkg/services"
	"github.com/robfig/cron/v3"
)

func StartScheduler() *cron.Cron {
	c := cron.New()

	_, err := c.AddFunc("0 2 * * *", func() {
		log.Println("Starting daily backup...")
		path, err := backup.PerformBackup()
		if err != nil {
			log.Printf("Backup failed: %v", err)
		} else {
			log.Printf("Backup completed: %s", path)
		}

		err = backup.CleanupOldBackups(30)
		if err != nil {
			log.Printf("Cleanup failed: %v", err)
		}
	})
	if err != nil {
		log.Printf("Failed to add backup cron job: %v", err)
	}

	_, err = c.AddFunc("0 * * * *", func() {
		log.Println("Checking thresholds...")
		notificationService := services.NewNotificationService()
		err := notificationService.CheckThresholds()
		if err != nil {
			log.Printf("Threshold check failed: %v", err)
		}
	})
	if err != nil {
		log.Printf("Failed to add threshold check cron job: %v", err)
	}

	c.Start()
	log.Println("Scheduler started")

	return c
}
