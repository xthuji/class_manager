package main

import (
	"context"
	"embed"
	"log"

	"github.com/class_manager/pkg/api"
	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/utils"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	studentAPI        *api.StudentAPI
	courseAPI         *api.CourseAPI
	hourRecordAPI     *api.HourRecordAPI
	hourRechargeAPI   *api.HourRechargeAPI
	scheduleAPI       *api.ScheduleAPI
	thresholdAPI      *api.ThresholdAPI
	notificationAPI   *api.NotificationAPI
}

func NewApp() *App {
	return &App{
		studentAPI:      api.NewStudentAPI(),
		courseAPI:       api.NewCourseAPI(),
		hourRecordAPI:   api.NewHourRecordAPI(),
		hourRechargeAPI: api.NewHourRechargeAPI(),
		scheduleAPI:     api.NewScheduleAPI(),
		thresholdAPI:    api.NewThresholdAPI(),
		notificationAPI: api.NewNotificationAPI(),
	}
}

func main() {
	utils.InitLogger()

	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Class Manager",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: options.NewRGBA(255, 255, 255, 1),
		OnStartup: func(ctx context.Context) {
			log.Println("Class Manager started")
		},
		OnShutdown: func(ctx context.Context) {
			log.Println("Class Manager shutting down")
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}

func (a *App) ExportAllData() (string, error) {
	return db.ExportAllData()
}

func (a *App) ImportAllData(jsonData string) (bool, error) {
	err := db.ImportAllData(jsonData)
	return err == nil, err
}

func (a *App) GetAppInfo() map[string]string {
	return map[string]string{
		"name":    "Class Manager",
		"version": "1.0.0",
	}
}