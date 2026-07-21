#!/bin/bash

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$PROJECT_DIR/build"
FRONTEND_DIR="$PROJECT_DIR/frontend"
APP_PATH="$BUILD_DIR/bin/ClassManager.app"

BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

check_dependencies() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go first."
        exit 1
    fi
    
    if ! command -v npm &> /dev/null; then
        log_error "Node.js/NPM is not installed. Please install Node.js first."
        exit 1
    fi
    
    if [ -z "$WAILS_CMD" ]; then
        if command -v wails &> /dev/null; then
            WAILS_CMD="wails"
        elif [ -x "$HOME/go/bin/wails" ]; then
            WAILS_CMD="$HOME/go/bin/wails"
        else
            log_warn "Wails CLI is not installed. Installing..."
            go install github.com/wailsapp/wails/v2/cmd/wails@latest
            if [ -x "$HOME/go/bin/wails" ]; then
                WAILS_CMD="$HOME/go/bin/wails"
            else
                log_error "Failed to install Wails CLI"
                exit 1
            fi
        fi
    fi
}

needs_frontend_build() {
    local dist_dir="$FRONTEND_DIR/dist"
    local src_dir="$FRONTEND_DIR/src"
    
    if [ ! -d "$dist_dir" ]; then
        return 0
    fi
    
    local dist_mtime=$(find "$dist_dir" -type f -exec stat -f "%m" {} \; | sort -n | tail -1)
    local src_mtime=$(find "$src_dir" -type f -exec stat -f "%m" {} \; | sort -n | tail -1)
    
    if [ -z "$dist_mtime" ]; then
        return 0
    fi
    
    if [ -z "$src_mtime" ]; then
        return 1
    fi
    
    if [ "$src_mtime" -gt "$dist_mtime" ]; then
        return 0
    fi
    
    return 1
}

needs_backend_build() {
    if [ ! -d "$APP_PATH" ]; then
        return 0
    fi
    
    local app_mtime=$(stat -f "%m" "$APP_PATH")
    local go_mtime=$(find "$PROJECT_DIR" -name "*.go" -type f -exec stat -f "%m" {} \; | sort -n | tail -1)
    
    if [ -z "$go_mtime" ]; then
        return 1
    fi
    
    if [ "$go_mtime" -gt "$app_mtime" ]; then
        return 0
    fi
    
    return 1
}

clean_build() {
    log_info "Cleaning build directory..."
    rm -rf "$BUILD_DIR"
    rm -f "$PROJECT_DIR/class_manager.sqlite"
    log_success "Build directory cleaned"
}

install_frontend_deps() {
    if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
        log_info "Installing frontend dependencies..."
        cd "$FRONTEND_DIR"
        npm install
        log_success "Frontend dependencies installed"
    fi
}

build_frontend() {
    install_frontend_deps
    
    log_info "Building frontend..."
    cd "$FRONTEND_DIR"
    npm run build
    log_success "Frontend built successfully"
}

build_app() {
    check_dependencies
    
    if needs_frontend_build; then
        build_frontend
    else
        log_info "Frontend is up to date, skipping build"
    fi
    
    log_info "Building Wails application..."
    cd "$PROJECT_DIR"
    "$WAILS_CMD" build -platform darwin/universal
    
    if [ -d "$APP_PATH" ]; then
        APP_SIZE=$(du -sh "$APP_PATH" | cut -f1)
        log_success "Application built successfully: $APP_PATH ($APP_SIZE)"
    else
        log_error "Application build failed"
        exit 1
    fi
}

run_tests() {
    check_dependencies
    
    log_info "Running tests..."
    cd "$PROJECT_DIR"
    go test ./...
    log_success "All tests passed"
}

run_app() {
    check_dependencies
    
    if needs_frontend_build; then
        build_frontend
    else
        log_info "Frontend is up to date, skipping build"
    fi
    
    if needs_backend_build; then
        log_info "Building Wails application..."
        cd "$PROJECT_DIR"
        "$WAILS_CMD" build -platform darwin/universal
        
        if [ ! -d "$APP_PATH" ]; then
            log_error "Application build failed"
            exit 1
        fi
    else
        log_info "Backend is up to date, skipping build"
    fi
    
    log_info "Starting application..."
    open "$APP_PATH"
    log_success "Application started"
}

show_menu() {
    echo ""
    echo "=========================================="
    echo "          Class Manager - 操作菜单"
    echo "=========================================="
    echo "  1) 清理构建 (Clean)"
    echo "  2) 构建应用 (Build)"
    echo "  3) 运行测试 (Test)"
    echo "  4) 启动应用 (Run)"
    echo "  5) 退出 (Exit)"
    echo "=========================================="
    echo -n "请输入选择 [1-5]: "
}

handle_menu() {
    show_menu
    read -r choice
    
    case "$choice" in
        1)
            clean_build
            ;;
        2)
            build_app
            ;;
        3)
            run_tests
            ;;
        4)
            run_app
            ;;
        5)
            echo "退出..."
            exit 0
            ;;
        *)
            log_error "无效选择，请输入 1-5"
            handle_menu
            ;;
    esac
}

main() {
    if [ $# -eq 0 ]; then
        handle_menu
        return
    fi
    
    case "$1" in
        clean)
            clean_build
            ;;
        build)
            build_app
            ;;
        test)
            run_tests
            ;;
        run)
            run_app
            ;;
        help)
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  clean   Clean build directory"
            echo "  build   Build the application (auto-detects changes)"
            echo "  test    Run tests"
            echo "  run     Start the application (auto-builds if needed)"
            echo "  help    Show this help message"
            echo ""
            echo "不带参数运行时将显示交互式菜单"
            echo ""
            ;;
        *)
            log_error "Unknown command: $1"
            echo "Run '$0 help' for available commands"
            exit 1
            ;;
    esac
}

main "$@"