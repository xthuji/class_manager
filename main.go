package main

import (
	"context"
	"os"
	"strconv"

	"github.com/class_manager/pkg/api"
	"github.com/class_manager/pkg/config"
	"github.com/class_manager/pkg/db"
	"github.com/class_manager/pkg/server"
	"github.com/class_manager/pkg/utils"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func main() {
	utils.InitLogger()

	cfg := config.Load()
	utils.LogInfo("Loaded config: port=" + strconv.Itoa(cfg.Port) + ", allow_port_fallback=" + strconv.FormatBool(cfg.AllowPortFallback) + ", width=" + strconv.Itoa(cfg.Width) + ", height=" + strconv.Itoa(cfg.Height))

	if err := db.InitDB(); err != nil {
		utils.LogErrorf("Failed to initialize database: %v", err)
		panic(err)
	}
	defer db.CloseDB()

	srv := server.NewServer()

	// 在独立 goroutine 中启动 HTTP 服务器，提供 RESTful API 和静态文件服务
	utils.LogInfo("Starting HTTP server on port " + strconv.Itoa(cfg.Port) + " (allow_port_fallback=" + strconv.FormatBool(cfg.AllowPortFallback) + ")")
	go srv.Start(cfg.Port, cfg.AllowPortFallback)

	// 启动时检查学生剩余课时是否达到阈值并生成通知
	notifAPI := api.NewNotificationAPI()
	if _, err := notifAPI.CheckThresholds(); err != nil {
		utils.LogErrorf("Startup threshold check failed: %v", err)
	} else {
		utils.LogInfo("Startup threshold check completed")
	}

	fileService := &FileService{}

	err := wails.Run(&options.App{
		Title:  "Class Manager",
		Width:  cfg.Width,
		Height: cfg.Height,
		AssetServer: &assetserver.Options{
			// Wails 内嵌窗口使用同一个 handler 提供 API 和静态页面
			Handler: srv.Handler(),
		},
		BackgroundColour: options.NewRGBA(255, 255, 255, 1),
		Bind: []interface{}{
			fileService,
		},
		OnStartup: func(ctx context.Context) {
			utils.LogInfo("Class Manager started")
			fileService.ctx = ctx
		},
		OnShutdown: func(ctx context.Context) {
			utils.LogInfo("Class Manager shutting down")
		},
	})

	if err != nil {
		utils.LogErrorf("Failed to run app: %v", err)
		panic(err)
	}
}

// FileService 提供文件操作相关的 Go 函数，可绑定到 JavaScript
type FileService struct {
	ctx context.Context
}

// ExportDataToFile 打开保存文件对话框并写入内容
func (fs *FileService) ExportDataToFile(content string, defaultFilename string) (string, error) {
	path, err := runtime.SaveFileDialog(fs.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON 文件 (*.json)", Pattern: "*.json"},
			{DisplayName: "所有文件", Pattern: "*"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

// ExportToFile 打开保存文件对话框并写入内容（支持多种格式）
func (fs *FileService) ExportToFile(content string, defaultFilename string, fileType string) (string, error) {
	var filters []runtime.FileFilter
	switch fileType {
	case "csv":
		filters = []runtime.FileFilter{
			{DisplayName: "CSV 文件 (*.csv)", Pattern: "*.csv"},
			{DisplayName: "所有文件", Pattern: "*"},
		}
	case "xlsx":
		filters = []runtime.FileFilter{
			{DisplayName: "Excel 文件 (*.xlsx)", Pattern: "*.xlsx"},
			{DisplayName: "所有文件", Pattern: "*"},
		}
	default:
		filters = []runtime.FileFilter{
			{DisplayName: "所有文件", Pattern: "*"},
		}
	}

	path, err := runtime.SaveFileDialog(fs.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Filters:         filters,
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

// ImportDataFromFile 打开文件选择对话框并读取文件内容
func (fs *FileService) ImportDataFromFile() (string, error) {
	path, err := runtime.OpenFileDialog(fs.ctx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON 文件 (*.json)", Pattern: "*.json"},
			{DisplayName: "所有文件", Pattern: "*"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
