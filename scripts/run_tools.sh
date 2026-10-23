#!/usr/bin/env bash
# =============================================================================
# ClassManager - 构建与运行工具 (跨平台: macOS / Linux / Windows Git Bash)
# =============================================================================
# 命令:
#   1|b,build     构建桌面应用 (当前平台)
#   2|r,run       构建并前台运行 (Ctrl+C 关闭)
#   3|d,dev       开发模式: go run 启动内嵌 HTTP 服务 + Wails 窗口
#   4|t,test      运行 Go 单元测试
#   5|c,clean     清理构建产物
#   0|q|exit        退出
# =============================================================================
# 薄入口：公共实现全部在 run_tools_master.sh（跨项目逐字节一致，禁止单项目
# 私改），项目差异全部写在本文件的「常量与命令表」与「项目专属实现」内联区；
# 全部构建/打包/测试逻辑集中在这里，CI workflow 只做 runner 引导与产物上传
#（规范 4.2）。
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

# ============================================================
# 公共模块：母版实现（与其余桌面应用逐字节一致，请勿在此修改）
# 先于常量区加载：母版只定义函数不执行业务，差异常量声明在下一区
# ============================================================
source "$SCRIPT_DIR/run_tools_master.sh"
# ============================================================
# 项目差异区 · 常量与命令表（脚本唯一需要按项目调整的部分）
# ============================================================

# --- 项目标识 ---
APP_NAME="ClassManager"                    # 可执行文件 / Wails outputfilename
BUNDLE_NAME=""              # .app bundle 名（留空同 APP_NAME）
RELEASE_PREFIX=""        # release/ 产物命名前缀（留空同 APP_NAME）
APP_DISPLAY_NAME="ClassManager"
PROJECT_BUNDLE_ID="com.xthuji.ClassManager"
export PROJECT_BUNDLE_ID

# --- 目录与产物 ---
BUILD_DIR="$PROJECT_DIR/build"
FRONTEND_DIR="$PROJECT_DIR/frontend"
# 产物目录：默认 <frontend_dir>/dist；Vite outDir 指到别处（如 easy_diff 直写 embed
# 目录 frontend/dist）时必须按项目覆盖，否则母版 ensure_frontend 会误判为未构建
FRONTEND_DIST="$FRONTEND_DIR/dist"
WAILS_CMD=""
VERSION=""

# --- 构建策略 ---
IS_WAILS=1                     # 0=纯 Go 桌面（自行覆盖 build_impl / get_app_bundle）
PLATFORM_SCOPE="all"        # darwin=仅面向 macOS 产物
MACOS_MIN_VERSION="12.0"
MACOS_UNIVERSAL_DEFAULT=1     # 1=恒产 universal；0=本地单架构 + 环境变量开关
UNIVERSAL_ENV_VAR=""     # 例：EASYDIFF_UNIVERSAL
WAILS_BUILD_CLEAN=0
WAILS_PIN_VERSION=""                    # 留空=对齐 go.mod 的 wails 版本（真源）
NEEDS_CGO=0                   # 1=Windows 构建需要 MinGW-w64
ADHOC_SIGN=1                      # 1=macOS 构建后 ad-hoc 签名
GO_BUILD_LDFLAGS="-s -w"
VERSION_LD_VAR="main.Version"        # 版本注入的 Go 变量（ldflags -X）
LDFLAGS_EXTRA=""
GOTOOLCHAIN_VER="go1.22.12"

# --- 资源与配置 ---
CONFIG_SRC="$PROJECT_DIR/configs/config.json"
CONFIG_DEST_DIR=""
ICON_SRC="$PROJECT_DIR/assets/appicon.png"

# --- 菜单与分发 ---
DEFAULT_CMD="run"
FALLBACK_CMD="cmd_dev"
SRC_FIND_EXPR=( -name '*.go' -o -path '*/pkg/server/static/*' -o -name 'VERSION' -o -name 'wails.json' -o -name 'config.json' )             # needs_build 的源文件匹配表达式
TEST_PKGS=( ./tests/... )
CLEAN_EXTRA=(  )
DEV_CMD=( go run . )                    # 留空=母版默认 wails dev

# --- 命令表：序号|别名(逗号分隔)|说明|处理函数（菜单/帮助/分发唯一真源）---
commands=(
  "1|b,build|构建桌面应用 (当前平台)|cmd_build"
  "2|r,run|构建并前台运行 (Ctrl+C 关闭)|cmd_run"
  "3|d,dev|开发模式: go run 启动内嵌 HTTP 服务 + Wails 窗口|cmd_dev"
  "4|t,test|运行 Go 单元测试|cmd_test"
  "5|c,clean|清理构建产物|cmd_clean"
)

# ============================================================
# 项目专属实现（覆盖母版默认 / 追加项目专属命令）
# 内联区位于常量表之后、main "$@" 之前：顶层可安全展开差异常量
# ============================================================
# 前端为 Go 内嵌静态资源（pkg/server/server.go 的 //go:embed static/* 经
# wails.Run 的 AssetServer.Handler 直出），frontend/ 下无 package.json 也非
# 构建输入。母版 ensure_frontend 会因 frontend/dist/*.js 不存在而误触发
# npm install -> ENOENT，故此处整体跳过前端构建（同 easy_download/launchpad）。
ensure_frontend() { :; }

# Go 1.22 链接器默认不生成 Mach-O LC_UUID，新版 macOS (26+) 的 dyld 会直接
# abort 加载缺失 LC_UUID 的二进制，故 darwin 构建追加 -B gobuildid；
# Linux/Windows 无此问题，保持精简 ldflags。
build_impl() {
  local ldflags="$1"
  # 前端为 Go 内嵌静态资源，脚本侧无 npm 步骤；-s 让 wails 跳过内部
  # frontend:install / build（当前 wails.json 未配置，属双保险）
  if [ "$HOST_OS" = "darwin" ]; then
    run "$WAILS_CMD" build -platform "$(get_wails_platform)" -s -ldflags "$ldflags -B gobuildid"
  else
    run "$WAILS_CMD" build -platform "$(get_wails_platform)" -s -ldflags "$ldflags"
  fi
}

cmd_test() {
  check_go
  log_info "运行 Go 单元测试..."
  GOTOOLCHAIN="$GOTOOLCHAIN_VER" go test -v ./tests/...
  log_success "全部测试通过"
}

main "$@"
