package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/class_manager/pkg/config"
)

// 已知的默认配置值（与 config.defaultConfig 保持一致）。
// 由于 defaultConfig 是未导出的，测试中直接使用这些硬编码值。
const (
	defaultPort   = 3100
	defaultWidth  = 1280
	defaultHeight = 800
)

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config.json failed: %v", err)
	}
}

// removeExeDirConfig 删除可执行文件同目录下的 config.json，
// 防止先前测试中 writeDefaultToDisk 写入的默认配置干扰后续测试。
func removeExeDirConfig(t *testing.T) {
	t.Helper()
	if exe, err := os.Executable(); err == nil {
		path := filepath.Join(filepath.Dir(exe), "config.json")
		_ = os.Remove(path)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// 清理可能存在的可执行文件目录 config.json，确保此测试触发 writeDefaultToDisk
	removeExeDirConfig(t)

	// 使用一个空目录作为 CWD，确保没有 config.json
	tmpDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	// 由于 Load 会同时检查可执行文件目录和 CWD，
	// 测试中我们直接验证默认值在无 config.json 时仍合理
	cfg := config.Load()
	if cfg.Port <= 0 {
		t.Fatalf("expected default port > 0, got %d", cfg.Port)
	}
	if cfg.Width <= 0 {
		t.Fatalf("expected default width > 0, got %d", cfg.Width)
	}
	if cfg.Height <= 0 {
		t.Fatalf("expected default height > 0, got %d", cfg.Height)
	}

	// 测试结束后清理可执行文件目录中的 config.json，避免影响其他测试
	t.Cleanup(func() { removeExeDirConfig(t) })
}

func TestLoad_FromCWD(t *testing.T) {
	removeExeDirConfig(t)

	tmpDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	writeConfig(t, tmpDir, `{"port": 9999, "width": 1024, "height": 768}`)

	// 切换 CWD 到临时目录
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	// os.Executable() 在测试环境下通常指向 go test 二进制，
	// 但 candidateDirs 会先尝试 executable dir，再尝试 CWD。
	// 这里我们靠 CWD 命中。
	cfg := config.Load()
	if cfg.Port != 9999 {
		t.Fatalf("expected port 9999, got %d", cfg.Port)
	}
	if cfg.Width != 1024 {
		t.Fatalf("expected width 1024, got %d", cfg.Width)
	}
	if cfg.Height != 768 {
		t.Fatalf("expected height 768, got %d", cfg.Height)
	}
}

func TestLoad_InvalidJSONFallsBackToDefaults(t *testing.T) {
	removeExeDirConfig(t)

	tmpDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	writeConfig(t, tmpDir, `{not valid json}`)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cfg := config.Load()
	// 解析失败应回退到默认值
	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %d, got %d", defaultPort, cfg.Port)
	}
}

func TestLoad_MissingFieldsApplyDefaults(t *testing.T) {
	removeExeDirConfig(t)

	tmpDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 只提供 port，不提供 width/height
	writeConfig(t, tmpDir, `{"port": 7777}`)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cfg := config.Load()
	if cfg.Port != 7777 {
		t.Fatalf("expected port 7777, got %d", cfg.Port)
	}
	if cfg.Width != defaultWidth {
		t.Fatalf("expected default width, got %d", cfg.Width)
	}
	if cfg.Height != defaultHeight {
		t.Fatalf("expected default height, got %d", cfg.Height)
	}
}
