# Class Manager

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

学生课时管理桌面应用(macOS / Linux / Windows)，用于管理学生、课程、课时记录、充值、排课与到期提醒。基于 [Wails v2](https://wails.io) 构建：Go 后端内嵌 HTTP 服务，WebView 加载原生 HTML/CSS/JS 前端，数据存储在本地 SQLite，无需任何外部服务。

> 本项目使用 AI 辅助编程工具开发完成。

## 功能特性

- **学生管理**：学生信息维护、剩余课时跟踪、搜索
- **课程管理**：课程定义与课时配置
- **课时记录**：上课消课自动扣减、历史记录查询
- **课时充值**：充值流水管理
- **排课**：课程安排与日历视图
- **到期/阈值通知**：课时不足等阈值提醒
- **操作日志**：全量变更审计记录
- **数据导出/导入**：全量数据备份与恢复

## 技术栈

| 组件 | 选型 |
|------|------|
| 桌面框架 | Wails v2.9.0 (Go 后端 + WebView) |
| 开发语言 | Go 1.22 (工具链锁定 `go1.22.12`) |
| 数据库 | SQLite (`modernc.org/sqlite`，纯 Go 无 CGO) |
| 前端 | 原生 HTML/CSS/JS，经 `go:embed` 内嵌，无 Node 构建链 |
| HTTP | 标准库 `net/http` (Go 1.22 `ServeMux` 路由)，同一 Handler 提供 API 与静态资源 |

## 环境要求

**Go & Wails CLI (三平台通用)**

- Go 1.22 (须保持 `go.mod` 中 `toolchain go1.22.12`，禁止升级到 Go 1.23+，否则破坏 macOS 10.15 兼容)
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation) (`scripts/run_tools.sh` 缺失时会自动安装 v2.9.0)

**平台开发依赖**

| 平台 | WebView 运行时 | 说明 |
|------|----------------|------|
| macOS | 系统自带 | 10.15 (Catalina) 及以上 |
| Linux | WebKit2GTK | `sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev build-essential pkg-config libglib2.0-dev libsoup-3.0-dev javascriptcoregtk-4.0-dev zip` |
| Windows | WebView2 | Windows 10 及以上，已内置 Edge WebView2 运行时 |

## 快速开始

项目提供统一入口脚本 `scripts/run_tools.sh`（交互式菜单，不带参数运行显示菜单）：

```bash
# 构建应用（自动适配当前平台）
./scripts/run_tools.sh build
# macOS 产物: build/bin/ClassManager.app + release/ClassManager_v{版本}.dmg
# Linux  产物: build/bin/ClassManager + release/ClassManager_v{版本}_linux_amd64.zip
# Windows 产物: build/bin/ClassManager.exe + release/ClassManager_v{版本}_windows_amd64.zip

# 运行应用（自动按需构建后启动）
./scripts/run_tools.sh run

# 运行测试
./scripts/run_tools.sh test

# 清理构建目录
./scripts/run_tools.sh clean
```

也可以直接手动执行（必须显式指定工具链）：

```bash
GOTOOLCHAIN=go1.22.12 go test -v ./tests/...
# macOS
GOTOOLCHAIN=go1.22.12 wails build -platform darwin/universal
# Linux
wails build -platform linux/amd64
# Windows (PowerShell)
wails build -platform windows/amd64
```

## 项目结构

```
├── main.go                  # 应用入口（日志→配置→DB→HTTP→Wails 窗口）
├── configs/config.json      # 应用配置（端口/窗口尺寸）
├── docs/                    # 项目文档
│   └── ARCHITECTURE.md      # 架构设计与开发规范
├── pkg/
│   ├── api/                 # HTTP 接口层（参数解析与响应封装）
│   ├── services/            # 业务逻辑层（含审计日志、事务）
│   ├── models/              # 数据模型
│   ├── db/                  # SQLite 连接、Schema、迁移、导入导出
│   ├── config/              # 配置加载（多路径优先级）
│   ├── server/              # 路由注册 + 静态资源（go:embed）
│   │   └── static/          # 原生前端（HTML/CSS/JS）
│   └── utils/               # 日志等工具
├── scripts/
│   ├── run_tools.sh         # 构建/运行/测试/清理统一入口（跨平台）
│   └── release.sh           # 发布工具（打 tag 触发 CI）
├── tests/                   # Service 级 + HTTP 级集成测试
├── build/                   # 构建产物（已忽略）
└── .github/workflows/       # GitHub Actions 发布工作流
```

## 配置说明

`configs/config.json`（构建时自动同步到 macOS `.app/Contents/Resources/`、Linux/Windows zip 同目录）：

```json
{
  "port": 3100,
  "allow_port_fallback": true,
  "width": 1000,
  "height": 800
}
```

配置加载优先级（`config.Load`）：macOS `.app/Contents/Resources/` → 可执行文件同目录 → `cwd/configs/` → `cwd/` → 内置默认值。

## 数据存储

- **macOS**：`~/Library/Application Support/ClassManager/class_manager.sqlite`
- **Linux / Windows**：可执行文件同目录下 `class_manager.sqlite`
- 日志位于同目录 `logs/`
- 应用内置全量数据导出/导入功能，可用于备份与迁移

## 版本发布

版本号以根目录 `VERSION` 文件为唯一数据源，构建时自动同步到 `wails.json` 与 macOS `Info.plist`。

发布流程（需配置 GitHub 远程仓库，默认远程名为 `origin`，可用 `--remote` 覆盖）：

```bash
# 1. 修改 VERSION（如 1.1.0）并提交代码

# 2. 预览发布动作
./scripts/release.sh --dry-run

# 3. 打 tag v{version} 并推送，触发 GitHub Actions
./scripts/release.sh
```

推送 tag 后，[.github/workflows/release.yml](.github/workflows/release.yml) 会在 **macOS 15 / Ubuntu 22.04 / Windows latest** 三平台并行构建：

| 平台 | Runner | Wails 目标 | 产物 |
|------|--------|-----------|------|
| macOS | `macos-15` | `darwin/universal` | `.dmg` |
| Linux | `ubuntu-22.04` | `linux/amd64` | `.zip` |
| Windows | `windows-latest` | `windows/amd64` | `.zip` |

所有产物自动上传到同一个 GitHub Release。

## 兼容性说明

**macOS（最高优先级）**

本项目必须兼容 macOS 10.15 (Catalina)：
- Go 锁定 `1.22.12` —— Go 1.23+ 已放弃 10.15 支持，**禁止升级**
- 10.15 仅支持 amd64；`darwin/universal` 产物中的 amd64 切片可在 10.15 运行
- 全部依赖须兼容 Go 1.22，禁止引入需要 CGO 或 macOS 11+ 专属 API 的依赖
- 发布前须在 macOS 10.15 (amd64) 环境实测构建产物可正常安装运行

**Linux**

- 已在 Ubuntu 22.04 runner 上构建验证，`libgtk-3-dev` / `libwebkit2gtk-4.0-dev` 等系统依赖已在 workflow 中安装
- 纯 Go 无 CGO，无额外编译依赖

**Windows**

- 已在 `windows-latest` runner 上构建验证，使用内置 WebView2
- 纯 Go 无 CGO，无额外编译依赖

## 开发规范

详细的架构设计、分层约定、测试规范与新增功能 Recipe 见 [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)。

## 开源许可

本项目基于 [MIT License](LICENSE) 开源。

## AI 辅助开发声明

本项目在开发过程中使用了 AI 辅助编程工具，在架构设计、代码编写、测试生成等方面获得了 AI 的支持。所有代码均经过人工审查与验证。
