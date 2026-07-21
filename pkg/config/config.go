package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// Config 应用配置，从 config.json 读取
type Config struct {
	Port              int  `json:"port"`                // HTTP 服务端口
	AllowPortFallback bool `json:"allow_port_fallback"` // 端口被占用时是否自动寻找可用端口（开发时设为 false 便于调试，交付时设为 true 避免端口冲突）
	Width             int  `json:"width"`               // 窗口宽度
	Height            int  `json:"height"`              // 窗口高度
}

var defaultConfig = Config{
	Port:              3100,
	AllowPortFallback: false,
	Width:             1280,
	Height:            800,
}

// Load 从可执行文件同目录或当前工作目录加载 config.json，
// 若都未找到则使用硬编码的默认配置。
func Load() Config {
	for _, dir := range candidateDirs() {
		path := filepath.Join(dir, "config.json")
		if data, err := os.ReadFile(path); err == nil {
			cfg := defaultConfig
			if err := json.Unmarshal(data, &cfg); err != nil {
				log.Printf("Warning: failed to parse %s: %v, using defaults", path, err)
				return defaultConfig
			}
			applyDefaults(&cfg)
			log.Printf("Config loaded from %s: port=%d, allow_port_fallback=%v, width=%d, height=%d", path, cfg.Port, cfg.AllowPortFallback, cfg.Width, cfg.Height)
			return cfg
		}
	}

	// 使用硬编码默认配置
	cfg := defaultConfig
	applyDefaults(&cfg)
	log.Printf("No external config.json found, using default: port=%d, allow_port_fallback=%v, width=%d, height=%d", cfg.Port, cfg.AllowPortFallback, cfg.Width, cfg.Height)

	// 将默认配置写入可执行文件同目录，方便用户后续修改
	writeDefaultToDisk()

	return cfg
}

// writeDefaultToDisk 将默认配置写入可执行文件同目录（仅在文件不存在时写入）
func writeDefaultToDisk() {
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		
		// 不在临时目录中创建配置文件（防止构建过程中在临时目录创建配置）
		if isTempDir(exeDir) {
			log.Printf("Skipping writeDefaultToDisk in temp directory: %s", exeDir)
			return
		}
		
		path := filepath.Join(exeDir, "config.json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if data, err := json.MarshalIndent(defaultConfig, "", "  "); err == nil {
				if err := os.WriteFile(path, data, 0644); err == nil {
					log.Printf("Default config written to %s", path)
				}
			}
		}
	}
}

// isTempDir 判断目录是否为临时目录
func isTempDir(dir string) bool {
	tempDir := os.TempDir()
	return filepath.HasPrefix(dir, tempDir) || filepath.HasPrefix(dir, "/var/folders/")
}

func candidateDirs() []string {
	var dirs []string

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		// macOS .app bundle: 资源文件位于 Contents/Resources/
		// exeDir = .../Contents/MacOS, 需要向上两级到 .../Contents/Resources
		// 放在最前面，确保 .app bundle 内的配置优先
		resourcesDir := filepath.Join(filepath.Dir(exeDir), "Resources")
		dirs = append(dirs, resourcesDir)
		dirs = append(dirs, exeDir)
	}

	if cwd, err := os.Getwd(); err == nil {
		// 项目根目录下的 configs/ 子目录优先于当前目录
		dirs = append(dirs, filepath.Join(cwd, "configs"))
		dirs = append(dirs, cwd)
	}

	return dirs
}

func applyDefaults(cfg *Config) {
	if cfg.Port <= 0 {
		cfg.Port = defaultConfig.Port
	}
	if cfg.Width <= 0 {
		cfg.Width = defaultConfig.Width
	}
	if cfg.Height <= 0 {
		cfg.Height = defaultConfig.Height
	}
}
