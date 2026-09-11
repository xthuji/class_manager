#!/usr/bin/env bash
# =============================================================================
# Class Manager - 构建与运行工具 (跨平台: macOS / Linux / Windows Git Bash)
# =============================================================================
# 命令:
#   1|b|build    构建桌面应用 (Wails 生产构建, 自动适配当前平台)
#   2|r|run      构建并前台运行 (Ctrl+C 关闭)
#   3|d|dev      开发模式: 直接 go run 启动 HTTP 服务 + Wails 窗口
#   4|t|test     运行 Go 单元测试
#   5|c|clean    清理构建产物
#   0|q|exit     退出
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

# --- 项目常量 ---
APP_NAME="ClassManager"
BUILD_DIR="$PROJECT_DIR/build"
CONFIGS_DIR="$PROJECT_DIR/configs"
CONFIG_FILE="config.json"
ICON_FILE="$PROJECT_DIR/assets/appicon.png"
WAILS_CMD=""
VERSION=""
PROJECT_BUNDLE_ID="com.xthuji.ClassManager"
export PROJECT_BUNDLE_ID

# Go 1.22 内部链接器默认不生成 Mach-O LC_UUID，
# 新版 macOS (26+) 的 dyld 会直接 abort 加载缺失 LC_UUID 的二进制。
# -B gobuildid 显式生成 UUID
GO_BUILD_LDFLAGS="-s -w -B gobuildid"

# --- 平台检测 ---
HOST_OS=""
HOST_ARCH=""
EXT=""

detect_platform() {
  local os
  os="$(uname -s 2>/dev/null || echo "unknown")"
  case "$os" in
    Darwin*)  HOST_OS="darwin"  ;;
    Linux*)   HOST_OS="linux"   ;;
    MINGW*|MSYS*|CYGWIN*) HOST_OS="windows" ;;
    *)        HOST_OS="unknown" ;;
  esac

  local arch
  arch="$(uname -m 2>/dev/null || echo "unknown")"
  case "$arch" in
    x86_64|amd64)   HOST_ARCH="amd64" ;;
    arm64|aarch64)   HOST_ARCH="arm64" ;;
    *)               HOST_ARCH="amd64" ;;
  esac

  if [ "$HOST_OS" = "windows" ]; then
    EXT=".exe"
  else
    EXT=""
  fi
}

detect_platform

# --- 彩色输出 ---
BLUE='\033[0;34m'; GREEN='\033[0;32m'; RED='\033[0;31m'
YELLOW='\033[1;33m'; BOLD='\033[1m'; NC='\033[0m'
log_info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
log_success() { echo -e "${GREEN}[OK]${NC} $*"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

# ============================================================
# 依赖检查
# ============================================================
check_go() {
  command -v go &>/dev/null || log_error "Go 未安装 (https://go.dev/dl)"
}

check_wails() {
  if [ -n "$WAILS_CMD" ]; then return 0; fi
  if [ -x "$HOME/go/bin/wails" ]; then
    WAILS_CMD="$HOME/go/bin/wails"
  elif [ -n "${USERPROFILE:-}" ] && [ -x "$USERPROFILE/go/bin/wails.exe" ] 2>/dev/null; then
    WAILS_CMD="$USERPROFILE/go/bin/wails"
  elif command -v wails &>/dev/null; then
    WAILS_CMD="wails"
  else
    log_warn "Wails CLI 未安装，正在安装..."
    go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.0
    WAILS_CMD="$HOME/go/bin/wails"
  fi
  [ -x "$WAILS_CMD" ] || log_error "Wails CLI 安装失败"
}

# 非 root 且有 sudo 时才用 sudo（CI runner 上为免密 sudo）
SUDO=""
if [ "$(id -u 2>/dev/null || echo 0)" != "0" ] && command -v sudo &>/dev/null; then
  SUDO="sudo"
fi

APT_INDEX_REFRESHED=0
apt_refresh_once() {
  [ "${APT_INDEX_REFRESHED:-0}" = "1" ] && return 0
  if ! command -v apt-get &>/dev/null; then
    log_warn "未检测到 apt-get，跳过自动依赖安装。请确保已手动安装 Wails Linux 构建依赖"
    return 1
  fi
  APT_INDEX_REFRESHED=1
  log_info "刷新 apt 索引 ..."
  # shellcheck disable=SC2086
  $SUDO apt-get update -qq || log_warn "apt-get update 未完全成功，继续尝试安装"
  return 0
}

apt_install() {
  apt_refresh_once || return 1
  log_info "安装 Linux 构建依赖: $*"
  # shellcheck disable=SC2086
  DEBIAN_FRONTEND=noninteractive $SUDO apt-get install -y --no-install-recommends "$@" \
    || log_error "apt-get install 失败: $*"
}

# Linux: Wails v2 需要 CGO 工具链 + GTK3 + WebKit2GTK(4.0)；zip 用于打包发布产物
ensure_linux_deps() {
  local need=()
  command -v pkg-config &>/dev/null || need+=("pkg-config")
  command -v gcc &>/dev/null        || need+=("build-essential")
  command -v zip &>/dev/null        || need+=("zip")
  pkg-config --exists gtk+-3.0 2>/dev/null       || need+=("libgtk-3-dev")
  pkg-config --exists webkit2gtk-4.0 2>/dev/null \
    || need+=("libwebkit2gtk-4.0-dev" "libglib2.0-dev" "libsoup-3.0-dev" "javascriptcoregtk-4.0-dev")

  [ ${#need[@]} -eq 0 ] || apt_install "${need[@]}"

  if ! pkg-config --cflags --libs gtk+-3.0 webkit2gtk-4.0 >/dev/null 2>&1; then
    log_error "pkg-config 无法解析 gtk+-3.0/webkit2gtk-4.0：$(pkg-config --errors --exists gtk+-3.0 webkit2gtk-4.0 2>&1 | head -3)"
  fi
  log_success "Linux 构建依赖就绪 (webkit2gtk-4.0 → $(pkg-config --modversion webkit2gtk-4.0))"
}

# macOS: 仅需 Xcode Command Line Tools (clang)，CI runner 已内置
ensure_macos_deps() {
  command -v clang &>/dev/null \
    || log_error "缺少 clang，请安装 Xcode Command Line Tools: xcode-select --install"
}

# Windows: Wails v2 无需 CGO（WebView2 由系统提供），构建本身无额外系统依赖
ensure_windows_deps() {
  : # no-op
}

ensure_build_deps() {
  case "$HOST_OS" in
    linux)   ensure_linux_deps ;;
    darwin)  ensure_macos_deps ;;
    windows) ensure_windows_deps ;;
  esac
}

# ============================================================
# 版本管理
# ============================================================
get_version() {
  if [ -n "$VERSION" ]; then echo "$VERSION"; return; fi
  if [ -f "VERSION" ]; then
    VERSION=$(tr -d '[:space:]' < VERSION)
  else
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
  fi
  echo "$VERSION"
}

# 跨平台 sed -i: macOS 需要 -i ''，Linux/Git Bash 需要 -i
portable_sed_i() {
  if [ "$HOST_OS" = "darwin" ]; then
    sed -i '' "$@"
  else
    sed -i "$@"
  fi
}

sync_wails_version() {
  local version
  version=$(get_version)
  local wails_json="$PROJECT_DIR/wails.json"
  [ -f "$wails_json" ] || return 0

  local current
  current=$(grep '"productVersion"' "$wails_json" | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')
  if [ "$current" = "$version" ]; then
    return 0
  fi

  portable_sed_i "s|\"productVersion\": *\"[^\"]*\"|\"productVersion\": \"$version\"|" "$wails_json"
  portable_sed_i "s|\"version\": *\"[^\"]*\"|\"version\": \"$version\"|" "$wails_json"
  log_info "版本已同步到 wails.json: $version"
}

# ============================================================
# 平台特定路径
# ============================================================
get_build_binary() {
  if [ "$HOST_OS" = "windows" ]; then
    echo "$BUILD_DIR/bin/${APP_NAME}.exe"
  else
    echo "$BUILD_DIR/bin/${APP_NAME}"
  fi
}

get_wails_platform() {
  case "$HOST_OS" in
    darwin)  echo "darwin/universal" ;;
    linux)   echo "linux/${HOST_ARCH}" ;;
    windows) echo "windows/${HOST_ARCH}" ;;
    *)       echo "unknown" ;;
  esac
}

# macOS only: 同步 config.json 到 .app/Contents/Resources/
sync_config_macos() {
  [ "$HOST_OS" = "darwin" ] || return 0
  local app_path="$BUILD_DIR/bin/${APP_NAME}.app"
  local resources="$app_path/Contents/Resources"
  [ -d "$app_path" ] || return 0
  [ -f "$CONFIGS_DIR/$CONFIG_FILE" ] || return 0
  mkdir -p "$resources"
  cp -f "$CONFIGS_DIR/$CONFIG_FILE" "$resources/$CONFIG_FILE"
  log_info "config.json → $resources/"
}

# 同步配置文件到 release 目录（Linux/Windows zip 包需要）
sync_config_to_release() {
  local stage_dir="$1"
  [ -f "$CONFIGS_DIR/$CONFIG_FILE" ] || return 0
  mkdir -p "$stage_dir/configs"
  cp -f "$CONFIGS_DIR/$CONFIG_FILE" "$stage_dir/configs/$CONFIG_FILE"
}

# macOS only: 修复 Info.plist 中的 Bundle ID 和版本号
fix_app_info() {
  [ "$HOST_OS" = "darwin" ] || return 0
  local version
  version=$(get_version)
  local plist="$BUILD_DIR/bin/${APP_NAME}.app/Contents/Info.plist"
  [ -f "$plist" ] || return 0

  local changed=0
  local cur_bundle cur_ver
  cur_bundle=$(plutil -extract CFBundleIdentifier raw "$plist" 2>/dev/null || echo "")
  cur_ver=$(plutil -extract CFBundleShortVersionString raw "$plist" 2>/dev/null || echo "")

  if [ "$cur_bundle" != "$PROJECT_BUNDLE_ID" ]; then
    plutil -replace CFBundleIdentifier -string "$PROJECT_BUNDLE_ID" "$plist"
    changed=1
  fi
  if [ "$cur_ver" != "$version" ]; then
    plutil -replace CFBundleShortVersionString -string "$version" "$plist"
    plutil -replace CFBundleVersion -string "$version" "$plist"
    changed=1
  fi

  if [ "$changed" -eq 1 ]; then
    log_info "Info.plist 已更新: version=$version bundle=$PROJECT_BUNDLE_ID"
  fi
}

# ============================================================
# 构建步骤
# ============================================================
sync_app_icon() {
  mkdir -p "$BUILD_DIR"
  if [ -f "$ICON_FILE" ]; then
    cp -f "$ICON_FILE" "$BUILD_DIR/appicon.png"
    log_info "图标已同步: assets/appicon.png → build/appicon.png"
  else
    log_warn "未找到图标源文件 $ICON_FILE"
  fi
}

# 将构建产物复制到 release 目录
sync_release() {
  local release_dir="$PROJECT_DIR/release"
  mkdir -p "$release_dir"

  if [ "$HOST_OS" = "darwin" ]; then
    local app_path="$BUILD_DIR/bin/${APP_NAME}.app"
    local release_app="${release_dir}/${APP_NAME}.app"
    [ -d "$app_path" ] || return 0
    rm -rf "$release_app"
    cp -R "$app_path" "$release_app"
    log_success "Release App  → $release_app"
  else
    local bin_path
    bin_path=$(get_build_binary)
    [ -f "$bin_path" ] || return 0
    local release_bin="${release_dir}/${APP_NAME}${EXT}"
    cp -f "$bin_path" "$release_bin"
    chmod +x "$release_bin" 2>/dev/null || true
    log_success "Release Bin  → $release_bin"
  fi
}

# macOS only: 用 hdiutil 生成 DMG（文件名带版本号）
build_dmg() {
  [ "$HOST_OS" = "darwin" ] || return 0
  local version
  version=$(get_version)
  local release_dir="$PROJECT_DIR/release"
  local release_app="${release_dir}/${APP_NAME}.app"
  local release_dmg="${release_dir}/${APP_NAME}_v${version}.dmg"

  local source_app="$release_app"
  [ -d "$source_app" ] || source_app="$BUILD_DIR/bin/${APP_NAME}.app"
  [ -d "$source_app" ] || return 0

  mkdir -p "$release_dir"
  rm -f "$release_dmg"

  log_info "生成 DMG: ${APP_NAME}_v${version}.dmg ..."
  hdiutil create \
    -volname "$APP_NAME" \
    -srcfolder "$source_app" \
    -ov \
    -format UDZO \
    "$release_dmg" > /dev/null

  log_success "Release DMG  → $release_dmg"
}

# Windows / Linux: 生成 zip 包 (含可执行文件 + configs/)
build_zip() {
  [ "$HOST_OS" = "darwin" ] && return 0
  local version
  version=$(get_version)
  local release_dir="$PROJECT_DIR/release"
  local zip_name="${APP_NAME}_v${version}_${HOST_OS}_${HOST_ARCH}.zip"
  local zip_path="${release_dir}/${zip_name}"
  local stage_dir="${BUILD_DIR}/zip_stage"

  rm -rf "$stage_dir"
  mkdir -p "$stage_dir"

  local bin_path
  bin_path=$(get_build_binary)
  [ -f "$bin_path" ] || return 0
  cp -f "$bin_path" "$stage_dir/"

  sync_config_to_release "$stage_dir"

  if [ -f "$ICON_FILE" ]; then
    cp -f "$ICON_FILE" "$stage_dir/appicon.png"
  fi

  mkdir -p "$release_dir"
  rm -f "$zip_path"

  log_info "生成 ZIP: ${zip_name} ..."
  if command -v zip &>/dev/null; then
    (cd "$stage_dir" && zip -r "$zip_path" . > /dev/null)
  elif [ "$HOST_OS" = "windows" ] && command -v powershell &>/dev/null; then
    local stage_win zip_win
    stage_win=$(cygpath -w "$stage_dir" 2>/dev/null || echo "$stage_dir")
    zip_win=$(cygpath -w "$zip_path" 2>/dev/null || echo "$zip_path")
    powershell -NoProfile -Command "Compress-Archive -Path '${stage_win}\\*' -DestinationPath '${zip_win}' -Force"
  else
    rm -rf "$stage_dir"
    log_error "zip 命令不可用，请安装 zip 或使用 Windows PowerShell"
  fi
  rm -rf "$stage_dir"

  log_success "Release ZIP  → $zip_path"
}

# ============================================================
# 增量构建检查
# ============================================================

needs_build() {
  local binary
  binary=$(get_build_binary)

  # macOS 检查 .app 存在性
  if [ "$HOST_OS" = "darwin" ]; then
    [ -d "$BUILD_DIR/bin/${APP_NAME}.app" ] || return 0
  else
    [ -f "$binary" ] || return 0
  fi

  # 检查源文件是否比产物更新
  local newer
  newer=$(find "$PROJECT_DIR" \
    \( -name '*.go' -o -name 'VERSION' -o -name 'wails.json' -o -name 'config.json' \) \
    -not -path '*/build/*' \
    -not -path '*/release/*' \
    -not -path '*/.git/*' \
    -newer "$binary" \
    -print -quit 2>/dev/null)

  [ -n "$newer" ]
}

# ============================================================
# 命令实现
# ============================================================

kill_existing() {
  local pids
  pids=$(pgrep -f "${APP_NAME}" 2>/dev/null || true)
  if [ -n "$pids" ]; then
    log_info "停止已有 App 进程: $pids"
    echo "$pids" | xargs kill 2>/dev/null || true
    sleep 1
    pids=$(pgrep -f "${APP_NAME}" 2>/dev/null || true)
    if [ -n "$pids" ]; then
      echo "$pids" | xargs kill -9 2>/dev/null || true
    fi
    log_success "旧进程已停止"
  fi
}

cmd_build() {
  check_go; check_wails; ensure_build_deps
  local version
  version=$(get_version)
  sync_wails_version

  local platform
  platform=$(get_wails_platform)
  log_info "构建 $APP_NAME ($platform, version=$version)..."

  sync_app_icon

  # macOS 需要 GOTOOLCHAIN 指定 Go 版本，解决 LC_UUID 问题
  # Linux/Windows 使用系统 Go，ldflags 不需要 -B gobuildid
  local ldflags="$GO_BUILD_LDFLAGS -X main.Version=${version}"
  if [ "$HOST_OS" = "darwin" ]; then
    GOTOOLCHAIN=go1.22.12 "$WAILS_CMD" build -platform "$platform" -ldflags "$ldflags"
  else
    "$WAILS_CMD" build -platform "$platform" -ldflags "-s -w -X main.Version=${version}"
  fi
  log_success "构建完成"

  sync_config_macos
  fix_app_info
  sync_release

  if [ "$HOST_OS" = "darwin" ]; then
    build_dmg
  else
    build_zip
  fi

  local bin_path
  bin_path=$(get_build_binary)
  if [ -f "$bin_path" ]; then
    log_success "构建产物 → $bin_path"
  elif [ "$HOST_OS" = "darwin" ] && [ -d "$BUILD_DIR/bin/${APP_NAME}.app" ]; then
    log_success "构建产物 → $BUILD_DIR/bin/${APP_NAME}.app"
  fi
}

cmd_run() {
  if [ "$HOST_OS" = "windows" ]; then
    log_error "run 命令不支持 Windows，请直接运行 build/bin/${APP_NAME}.exe"
  fi

  local log_file="$BUILD_DIR/run.log"
  mkdir -p "$BUILD_DIR"

  kill_existing

  if needs_build; then
    cmd_build
  else
    log_success "源文件无变化，跳过构建"
  fi

  local binary
  if [ "$HOST_OS" = "darwin" ]; then
    binary="$BUILD_DIR/bin/${APP_NAME}.app/Contents/MacOS/${APP_NAME}"
  else
    binary=$(get_build_binary)
  fi

  log_info "启动 $APP_NAME (前台模式)..."
  log_info "按 Ctrl+C 关闭 App"
  echo ""

  : > "$log_file"
  "$binary" 2>&1 | tee -a "$log_file"
  local exit_code=${PIPESTATUS[0]}

  echo ""
  if [ "$exit_code" -eq 0 ]; then
    log_success "App 已正常退出"
  else
    log_warn "App 退出码: $exit_code"
  fi
}

# dev 模式：直接 go run 启动内嵌 HTTP 服务器
cmd_dev() {
  check_go
  mkdir -p "$BUILD_DIR"
  local log_file="$BUILD_DIR/dev.log"

  log_info "启动开发模式..."
  log_info "内嵌 HTTP 服务器 + Wails 窗口"
  log_info "按 Ctrl+C 停止"
  echo ""

  : > "$log_file"
  go run . 2>&1 | tee -a "$log_file"
  local exit_code=${PIPESTATUS[0]}

  echo ""
  if [ "$exit_code" -eq 0 ] || [ "$exit_code" -eq 130 ]; then
    log_success "开发模式已停止"
  else
    log_warn "退出码: $exit_code"
  fi
}

cmd_test() {
  check_go
  log_info "运行 Go 单元测试..."
  GOTOOLCHAIN=go1.22.12 go test -v ./tests/...
  log_success "全部测试通过"
}

cmd_clean() {
  log_info "清理构建产物..."
  kill_existing 2>/dev/null || true
  rm -rf build dist release
  find . -name '*.test' -delete 2>/dev/null || true
  log_success "清理完成"
}

# ============================================================
# 交互菜单
# ============================================================
commands=(
  "1|b|build    |构建桌面应用 (当前平台)"
  "2|r|run      |构建并前台运行 (Ctrl+C 关闭)"
  "3|d|dev      |开发模式: go run 启动内嵌 HTTP 服务"
  "4|t|test     |运行 Go 单元测试"
  "5|c|clean    |清理构建产物"
  "0|q|exit     |退出"
)

show_menu() {
  echo ""
  echo "=========================================="
  echo "    Class Manager - 构建工具 ($HOST_OS/$HOST_ARCH)"
  echo "=========================================="
  for c in "${commands[@]}"; do
    IFS='|' read -r num name desc <<< "$c"
    printf "  %s) %-10s %s\n" "$num" "$name" "$desc"
  done
  echo "=========================================="
  echo -n "请输入选择 [0-5]: "
}

show_help() {
  cat <<EOF
Class Manager 构建工具

用法: $0 [命令]

当前平台: $HOST_OS/$HOST_ARCH

命令:
  (无参数)    显示交互菜单（回车默认 run）
  1|build     构建桌面应用 ($(get_wails_platform))
  2|run       构建并前台运行 App (Ctrl+C 关闭)
  3|dev       开发模式: go run 启动内嵌 HTTP 服务
  4|test      运行 Go 单元测试
  5|clean     清理构建产物
  0|exit      退出

平台说明:
  macOS:   构建 .app + .dmg (darwin/universal)
  Linux:   构建 ELF 二进制 + .zip (linux/$HOST_ARCH)
  Windows: 构建 .exe + .zip (windows/$HOST_ARCH)

项目路径:
  构建产物:  $(get_build_binary)
  Release:   release/
  运行日志:  $BUILD_DIR/run.log
  配置文件:  $PROJECT_DIR/configs/config.json
  版本文件:  $PROJECT_DIR/VERSION
  Bundle ID: $PROJECT_BUNDLE_ID
EOF
}

main() {
  local cmd="${1:-}"

  if [ -z "$cmd" ]; then
    show_menu
    read -r choice
    cmd="${choice:-run}"
  fi

  case "$cmd" in
    1|b|build)      cmd_build ;;
    2|r|run|"")     cmd_run ;;
    3|d|dev)        cmd_dev ;;
    4|t|test)       cmd_test ;;
    5|c|clean)      cmd_clean ;;
    0|q|exit|quit)  echo "退出..."; exit 0 ;;
    help|-h|--help) show_help ;;
    *)              echo "默认 dev 模式"; cmd_dev ;;
  esac
}

main "$@"
