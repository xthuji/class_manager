# Class Manager 开发规范与架构指南

本规范基于主流 Go / Wails 桌面应用开发实践，结合本项目现有技术架构、测试架构与业务领域沉淀而成。**所有新增功能与维护改动都必须遵循此规范**，以保证代码一致性、可测试性与可维护性。

> **最高优先级约束｜macOS 10.15 (Catalina) 支持**：本项目必须兼容 macOS 10.15。这决定 Go 工具链锁定在 `go1.22.12`（Go 1.23+ 已放弃 10.15，**禁止升级**）；10.15 仅 amd64 可用；所有依赖须兼容 Go 1.22；禁用 macOS 11 Big Sur+ 专属 API。详见下文「macOS 10.15 支持约束」与第 10 章。

---

## 1. 项目概览

- **产品**：Class Manager —— 学生课时管理桌面应用（学生/课程/课时记录/充值/排课/通知/操作日志/仪表盘）。
- **形态**：跨平台桌面应用（macOS `.app` + `.dmg`、Linux ELF + `.zip`、Windows `.exe` + `.zip`），内嵌 HTTP 服务 + WebView 前端。
- **技术栈**：
  - **后端**：Go 1.22（macOS 强制 `GOTOOLCHAIN=go1.22.12`，Linux/Windows 使用系统 Go）
  - **桌面框架**：Wails v2.9.0（Go 后端 + WebView 前端，同一 HTTP Handler 提供 API 与静态资源）
  - **数据库**：SQLite，驱动 `modernc.org/sqlite`（纯 Go，无 CGO，跨平台编译）
  - **前端**：原生 HTML/CSS/JS（无构建步骤，通过 `go:embed` 内嵌），非 Node 前端框架
  - **HTTP**：标准库 `net/http`，使用 Go 1.22 的 `ServeMux` 方法+路径模式路由
- **构建/运行**：`scripts/run_tools.sh`（交互式菜单：build / run / dev / test / clean / exit，跨平台自适应）
- **模块路径**：`github.com/class_manager`（`go.mod` 中 `replace github.com/class_manager => .`）

### macOS 10.15 支持约束（最高优先级，适用于全部代码）

本项目必须支持 macOS 10.15 (Catalina)。这是首要平台约束，**所有技术选型、依赖升级、API 使用均不得破坏 10.15 兼容性**。

- **Go 工具链上限（关键）**：`Go 1.22.12` 是支持 macOS 10.15 的**最后一个** Go 版本——Go 1.23+ 已放弃 10.15、要求 macOS 11 Big Sur 或更高。因此：
  - `go.mod` 必须保持 `go 1.22` 与 `toolchain go1.22.12`，**禁止升级到 Go 1.23+**。
  - 所有构建/测试命令必须显式指定 `GOTOOLCHAIN=go1.22.12`（`scripts/run_tools.sh` 已强制），避免触发工具链自动下载更高版本。
  - 不得在 `go.mod` 写入会触发自动升级的更高 `go` 指令版本（如 `go 1.23`）。
- **CPU 架构**：macOS 10.15 **仅支持 amd64**（Apple Silicon arm64 需 macOS 11+）。`darwin/universal` 产物中的 amd64 切片在 10.15 运行，arm64 切片仅在 11+ 运行。**不要依赖仅 arm64 可用的行为/API**；新增构建目标时不得移除 amd64。
- **依赖管理**：所有 Go 依赖必须兼容 Go 1.22，且不得要求 macOS 11+ 或 Go 1.23+。升级依赖前须核对其 `go.mod` 最低 Go 版本；`modernc.org/sqlite` 升级尤需谨慎（新版本可能上调 Go 最低版本）。引入新依赖前先验证可在 Go 1.22 下 `go build`/`go test` 通过。
- **系统 API**：禁止使用 macOS 11 Big Sur 之后才有的原生 API（任何 CGO/系统调用/`syscall` 私有接口）。当前 `wails.json` 窗口配置（`titleBarStyle: hiddenInset`、`fullSizeContent`、`webviewIsTransparent`）均兼容 10.15，可继续使用；**新增 Wails 选项前先确认其 10.15 兼容性**。
- **CGO 禁用**：本项目不使用 CGO（采用纯 Go 的 `modernc.org/sqlite`）。**禁止引入需要 CGO 的依赖**，既为跨平台编译，也为规避链接器部署目标问题。如未来确需原生代码，必须设置 `MACOSX_DEPLOYMENT_TARGET=10.15` 并在 10.15 实测。
- **发布验证**：发布前必须在 macOS 10.15 (amd64) 环境实际运行验证（启动、核心功能、构建产物 `.app`/`.dmg` 安装运行）。`wails doctor` 通过仅代表开发机环境正常，**不等于** 10.15 兼容。

---

## 2. 分层架构（必须严格遵守）

请求流向（自上而下，禁止跨层调用、禁止反向依赖）：

```
main.go (启动: 日志 -> 配置 -> DB -> HTTP Server -> Wails 窗口)
   │
   ▼
pkg/server   HTTP 层：路由注册、请求解码、响应编码、panic recovery
   │          (handlers.go / server.go，调用 pkg/api)
   ▼
pkg/api      API 层：薄封装，仅做 service 调用与 DTO 传递，不含业务逻辑
   │          (每个领域一个 *_api.go，构造函数 New*API())
   ▼
pkg/services 业务层：所有业务逻辑、SQL、事务、操作日志记录
   │          (直接使用 db.DB 全局变量，构造函数 New*Service())
   ▼
pkg/db       数据层：DB 连接、Schema、Migrations、导出/导入
   │
pkg/models   模型层：实体结构 + 请求/响应 DTO（纯数据，无逻辑）
pkg/utils    工具层：日志（InitLogger / LogInfo / LogInfof / LogErrorf）
pkg/config   配置层：从 config.json 加载（端口/窗口尺寸/端口回退）
```

### 各层职责与约束

| 层 | 职责 | 禁止 |
|----|------|------|
| `server` | 路由、HTTP 请求/响应处理、panic 恢复 | 写业务逻辑、直接写 SQL（仪表盘等聚合查询除外） |
| `api` | 转发调用到 service | 写业务逻辑、直接访问 db |
| `services` | 业务逻辑、SQL、事务、操作日志 | 直接处理 HTTP、返回 HTTP 状态码 |
| `models` | 数据结构定义（含 JSON tag） | 含任何方法逻辑 |
| `db` | 连接管理、Schema、迁移 | 含业务语义 |

> **例外**：`pkg/server/handlers.go` 中的仪表盘统计（`handleDashboardStats` 等）含聚合 SQL。新增聚合查询类只读接口可沿用此模式，但写操作必须走 service 层。

---

## 3. 业务领域模型

| 领域 | 表 | 模型文件 | 说明 |
|------|----|----------|------|
| 学生 | `students` | `models/student.go` | 姓名/学号/联系方式/总课时/已完成课时/退学状态 |
| 课程 | `courses` | `models/course.go` | 课程名(唯一)/描述/阈值(默认2) |
| 选课 | `student_course` | - | 学生↔课程 多对多 |
| 课时记录 | `hour_records` | `models/hour_record.go` | 扣减课时（学生+课程+课时+日期+描述） |
| 课时充值 | `hour_recharges` | `models/hour_recharge.go` | 增加课时（学生+课时+日期+描述） |
| 排课 | `schedules` | `models/schedule.go` | 周几/节次/起止时间/消耗课时 |
| 通知 | `notifications` | `models/notification.go` | 课时阈值告警（unread/read） |
| 操作日志 | `operation_logs` | `models/operation_log.go` | 审计日志（操作类型/实体类型/实体ID/描述） |

**核心关系**：学生总课时 = 充值之和；已完成课时 = 课时记录之和；剩余课时 = 总课时 - 已完成课时。低于课程阈值时生成通知。

---

## 4. 编码规范

### 4.1 命名与结构

- **Service**：无状态结构体 `type XxxService struct{}`，构造 `func NewXxxService() *XxxService`。
- **API**：持有 service `type XxxAPI struct{ service *services.XxxService }`，构造 `func NewXxxAPI() *XxxAPI`，方法仅转发。
- **单例**：跨请求共享状态（如 `HolidayService` 缓存）使用 `sync.Once` 单例 `GetHolidayService()`。
- **包名**：全小写单数（`services`、`models`、`utils`）。

### 4.2 Model 定义规范

每个实体定义：实体结构 + Create/Update/List/Search 请求 DTO。

```go
type Student struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    // 计算字段（不落库）
    RemainingHours float64 `json:"remaining_hours"`
}

type StudentCreateRequest struct { /* 仅创建所需字段 */ }
type StudentUpdateRequest struct { ID int64; /* ... */ }
type StudentListRequest   struct { Page, PageSize int }
type StudentSearchRequest struct { Page, PageSize int; /* 过滤字段 */ }
```

- 所有 JSON 字段使用 `snake_case` tag。
- 计算字段（如 `RemainingHours`）在 service 的 `scan*` 方法中计算，不写入 DB。
- 时间字段：DB 存 RFC3339 字符串，Model 用 `time.Time`，在 scan 时 `time.Parse(time.RFC3339, str)`。

### 4.3 分页响应（统一格式）

所有列表/搜索接口返回 `models.PaginatedResponse`：

```go
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Total      int         `json:"total"`
    Page       int         `json:"page"`
    PageSize   int         `json:"page_size"`
    TotalPages int         `json:"total_pages"`
}
```

分页约定：`page` 默认 1，`pageSize` 默认 10；`offset = (page-1)*pageSize`；`totalPages = (total + pageSize - 1) / pageSize`。

### 4.4 数据库规范

- **时间**：DB 列用 `TEXT`，写入 `db.GetTimestamp()`（RFC3339）；读取时 `time.Parse`。
- **布尔**：DB 列用 `INTEGER`（0/1），Model 用 `bool`，scan 时 `isDropped == 1` 转换。
- **主键**：`INTEGER PRIMARY KEY AUTOINCREMENT`。
- **外键**：声明 `FOREIGN KEY`，但**不开启级联删除**，删除实体时在 service 层事务中手动清理关联数据（见 `DeleteStudent`）。
- **连接池**：`SetMaxOpenConns(10)` / `SetMaxIdleConns(5)`，支持嵌套查询与并发，避免单连接死锁。
- **SQL**：一律使用参数化查询 `?` 占位符，**禁止字符串拼接用户输入**（防注入）。动态 `IN (...)` 用占位符切片拼接。

### 4.5 Schema 与迁移

- 新表/新库结构写在 `pkg/db/schema.go` 的 `InitSchema()`（仅首次建库执行）。
- **结构变更**（加列/删列/重命名/加索引）追加到 `pkg/db/migrations.go` 的 `migrations` 切片，**幂等执行**：
  - `ALTER TABLE ADD COLUMN` → 重复执行报 `duplicate column`，已做错误白名单跳过。
  - `ALTER TABLE DROP/RENAME` → 报 `no such column`，已跳过。
  - `DROP TABLE` → 报 `no such table`，已跳过。
- **不要修改已下发的 migration**，只能追加新的。

### 4.6 事务

涉及多表写入/删除时必须用事务：`db.DB.Begin()` → `tx.Exec` → 成功 `tx.Commit()`，失败 `tx.Rollback()`。参考 `DeleteStudent`、`ImportAllData`。

### 4.7 操作日志（审计）

所有写操作（create/update/delete）必须记录操作日志，通过 `OperationLogService.LogChange`：

```go
logService := NewOperationLogService()
// create: 传入实体显示名，changes 忽略
logService.LogChange("create", "student", id, fmt.Sprintf("%s (%s)", name, studentID), nil)
// update: 仅传入变化字段；changes 为空则不记录
var changes []FieldChange
if old.Name != req.Name {
    changes = append(changes, FieldChange{Field: "姓名", Old: old.Name, New: req.Name})
}
logService.LogChange("update", "student", req.ID, req.Name, changes)
// delete: 传入实体显示名
logService.LogChange("delete", "student", id, old.Name, nil)
```

- `FieldChange.Field` 用中文显示名（姓名/联系方式/总课时…）。
- 实体名解析用 `resolveStudentName(id)` / `resolveCourseName(id)` 辅助函数。

### 4.8 日志

- 启动时 `utils.InitLogger()`，日志同时写文件与 stdout。
- 使用 `utils.LogInfo` / `utils.LogInfof` / `utils.LogErrorf`，**不要用裸 `log` / `fmt.Println`**（仅在 db 包初始化阶段可用）。
- 日志消息用中文，业务关键操作记录 ID 与关键字段。

---

## 5. HTTP / API 规范

### 5.1 路由注册（`pkg/server/server.go` 的 `registerRoutes`）

使用 Go 1.22 方法+路径模式：

```go
s.mux.HandleFunc("GET    /api/students",           s.handleListStudents)
s.mux.HandleFunc("POST   /api/students",           s.handleCreateStudent)
s.mux.HandleFunc("GET    /api/students/{id}",      s.handleGetStudent)
s.mux.HandleFunc("PUT    /api/students/{id}",      s.handleUpdateStudent)
s.mux.HandleFunc("POST   /api/students/batch-delete", s.handleBatchDeleteStudents)
```

- 路径风格：`/api/{资源复数}`，路径参数 `{id}`。
- 批量操作用 `/batch-xxx` 子路径（如 `/batch-delete`、`/batch-import`、`/batch-read`）。
- 路由顺序注意冲突：固定路径（如 `/generate-id`、`/search`）须在 `{id}` 之前注册，或用更具体的子路径（如 `/{sid}/courses`）避免被 `{id}` 吞掉。

### 5.2 Handler 规范

```go
func (s *Server) handleCreateStudent(w http.ResponseWriter, r *http.Request) {
    var req models.StudentCreateRequest
    if err := decodeJSON(r, &req); err != nil {
        writeError(w, http.StatusBadRequest, "无效的请求数据")
        return
    }
    student, err := s.studentAPI.CreateStudent(req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, student)
}
```

- 用辅助函数：`decodeJSON(r, &v)`、`writeJSON(w, status, data)`、`writeError(w, status, msg)`、`parseID(r, "id")`。
- 成功：创建 `201`，查询/更新 `200`，删除无 body 时 `200`。
- 错误响应统一 `{"error": "message"}`。
- 路径参数取值：`r.PathValue("id")`（Go 1.22）。

### 5.3 静态资源与 SPA

- 前端文件位于 `pkg/server/static/`，通过 `//go:embed static/*` 内嵌。
- `/` 走静态文件服务器；非 `/api/` 且文件不存在时回退到 `index.html`（SPA fallback）。
- `Server.Handler()` 返回带 panic recovery 的包装，**防止 handler panic 导致 Wails 应用崩溃**。新增 handler 无需自行 recover。

---

## 6. 测试架构

### 6.1 测试组织

- 所有测试在 `tests/` 包（与 `pkg` 平级），统一 `package tests`。
- 两种测试粒度：
  1. **Service 级**：直接调用 `services.NewXxxService()` 方法（如 `student_test.go`）。
  2. **HTTP/API 级**：用 `httptest` 调用 `server.NewServer()`（如 `server_test.go`）。

### 6.2 测试隔离（关键）

- `tests/main_test.go` 的 `TestMain` 在所有测试前创建**临时隔离 SQLite 库**，覆盖 `db.DB`，执行 `InitSchema()`。**绝不污染用户正式数据库**（正式库在 `~/Library/Application Support/ClassManager/`）。
- 每个测试用例开头调用 `resetTables(t)` 清空所有表，保证从干净状态开始。
- `resetTables` 清表顺序遵循外键依赖（先子后父）：notifications → schedules → hour_recharges → hour_records → student_course → courses → students → operation_logs。

### 6.3 测试辅助函数

```go
doRequest(t, srv, "POST", "/api/students", body)   // 发请求，返回 *httptest.ResponseRecorder
decodeJSONBody(t, w, &result)                       // 解码响应体
```

### 6.4 编写测试规范

- 命名：`TestXxxService_Method`（service 级）、`TestAPI_XxxCRUD`（HTTP 级）。
- 每个用例 `resetTables(t)` 开头。
- 用 `t.Fatalf` 给出明确失败信息（包含期望值与实际值）。
- 覆盖：正常路径 + 边界（空值/自定义ID/批量）+ 错误分支。

### 6.5 运行测试

```bash
# 推荐：通过 scripts/run_tools.sh
./scripts/run_tools.sh test
# 或直接（须指定工具链）
GOTOOLCHAIN=go1.22.12 go test -v ./...
```

---

## 7. 构建与运行（`scripts/run_tools.sh`）

- 无参数运行显示**交互式菜单**；直接回车默认 `run`。
- 命令：`build` / `run` / `dev` / `test` / `clean` / `exit`（数字或单词均可）。
- **跨平台自适应**：脚本自动检测当前 OS，选择对应的 Wails 目标平台与打包格式。
- **增量构建**：`needs_build` 检测 Go 源码、`VERSION`、`wails.json`、`configs/config.json` 是否新于二进制。
- **版本同步**：`sync_wails_version` 自动将 `VERSION` 文件同步到 `wails.json` 的 `productVersion`。
- 构建：
  - **macOS**：`GOTOOLCHAIN=go1.22.12 wails build -platform darwin/universal -ldflags "-B gobuildid"` —— `GOTOOLCHAIN` 必须显式指定以防止 Go 工具链自动升级到 1.23+；`-B gobuildid` 确保新版 macOS (26+) 的 dyld 能正确加载二进制。
  - **Linux**：`wails build -platform linux/amd64`。需系统先装 `libgtk-3-dev` / `libwebkit2gtk-4.0-dev` / `libsoup-3.0-dev` / `javascriptcoregtk-4.0-dev`。
  - **Windows**：`wails build -platform windows/amd64`。依赖内置 WebView2。
- 产物：

| 平台 | 构建产物 | 打包产物 |
|------|---------|---------|
| macOS | `build/bin/ClassManager.app` | `release/ClassManager_v{version}.dmg`（universal：amd64 + arm64） |
| Linux | `build/bin/ClassManager` | `release/ClassManager_v{version}_linux_amd64.zip`（含可执行文件 + `configs/config.json` + `appicon.png`） |
| Windows | `build/bin/ClassManager.exe` | `release/ClassManager_v{version}_windows_amd64.zip`（同 Linux 结构） |

- macOS 额外处理：构建后用 `plutil` 修正 `.app/Info.plist` 的 Bundle ID 和版本号；`configs/config.json` 同步到 `.app/Contents/Resources/`。
- Linux/Windows 额外处理：zip 打包时 `configs/config.json` 放在可执行文件同目录下，便于就地修改。
- **CI 发布**：推送 `v*` tag 触发 `.github/workflows/release.yml`，在 `macos-15` / `ubuntu-22.04` / `windows-latest` 三 runner 并行构建，全部产物上传到同一个 GitHub Release。

---

## 8. 配置规范

- 配置文件 `configs/config.json`：`port` / `allow_port_fallback` / `width` / `height`。
- 加载优先级（`config.Load`）：`.app/Contents/Resources/` → 可执行文件同目录 → `cwd/configs/` → `cwd/` → 硬编码默认值。
- `allow_port_fallback`：开发设 `false`（端口占用即失败便于调试），交付设 `true`（自动递增找可用端口）。
- **配置文件随 App 内嵌分发**（放在可执行文件同目录 / Resources），用户可就地修改。

---

## 9. 新增功能完整 Recipe（端到端）

以「新增一个 XXX 资源的 CRUD」为例，按顺序修改：

1. **Model** (`pkg/models/xxx.go`)
   - 定义 `Xxx` 实体 + `XxxCreateRequest` / `XxxUpdateRequest` / `XxxListRequest`。

2. **Schema / Migration** (`pkg/db/schema.go` + `pkg/db/migrations.go`)
   - 新表写入 `InitSchema()`；若需兼容旧库，追加幂等 migration。

3. **Service** (`pkg/services/xxx_service.go`)
   - `type XxxService struct{}` + `NewXxxService()`。
   - 实现 `Create` / `Update` / `Delete` / `BatchDelete` / `GetByID` / `List` / `Search`。
   - 每个 `Create/Update/Delete` 调用 `OperationLogService.LogChange` 记录审计。
   - `scan*` 辅助方法处理时间字符串与布尔转换。
   - 写操作多表时用事务。

4. **API** (`pkg/api/xxx_api.go`)
   - `type XxxAPI struct{ service *services.XxxService }` + `NewXxxAPI()`，方法转发到 service。

5. **Server** (`pkg/server/server.go` + `handlers.go`)
   - `Server` 结构体加字段 `xxxAPI *api.XxxAPI`，`NewServer` 中初始化。
   - `registerRoutes` 注册路由（注意路径冲突顺序）。
   - `handlers.go` 实现 `handleXxx` 方法，用 `decodeJSON`/`writeJSON`/`writeError`/`parseID`。

6. **前端** (`pkg/server/static/`) — 按需
   - 修改 `app.js` / `components.js` 调用新 API。

7. **测试** (`tests/xxx_test.go`)
   - `resetTables(t)` 开头。
   - Service 级 + HTTP 级测试，覆盖正常/边界/错误。

8. **验证**
   - `./scripts/run_tools.sh test` 全量测试通过（强制 `GOTOOLCHAIN=go1.22.12`）。
   - `./scripts/run_tools.sh build` 构建成功。
   - 涉及依赖/原生 API 变更时，须在 macOS 10.15 (amd64) 环境实测确认兼容。

---

## 10. 关键约束与注意事项

### macOS 10.15 兼容（最高优先级，macOS 专属）

- Go 锁定 `1.22.12`（Go 1.23+ 已弃 10.15，**禁止升级**）
- 10.15 仅 amd64；`darwin/universal` 产物中的 amd64 切片可在 10.15 运行
- 依赖须兼容 Go 1.22；禁用 macOS 11 Big Sur+ 专属 API
- 发布前须在 macOS 10.15 (amd64) 环境实测构建产物可正常安装运行
- **CI Runner** 固定用 `macos-15`：`macos-latest` 已切换到 macOS 26，其 dyld 严格校验 LC_UUID 且与 wails v2.9.0 存在 SDK 兼容性问题

### 三平台通用

- **Go 版本**：`go.mod` 锁定 `go 1.22` / `toolchain go1.22.12`。新增依赖须兼容此版本，且不得要求 Go 1.23+。
- **无 CGO**：使用 `modernc.org/sqlite` 纯 Go 驱动，**禁止引入需要 CGO 的依赖**（否则跨平台编译会失败）。
- **前端无构建**：直接编辑 `static/` 下 HTML/CSS/JS，不引入 Node 构建链。
- **删除操作**：本项目未启用外键级联，删除实体时**必须**在事务中手动清理所有关联表数据。
- **启动流程**：`main.go` 顺序为 日志 → 配置 → DB → HTTP Server(goroutine) → 阈值检查 → Wails 窗口。新增启动任务须注意此顺序与 goroutine 边界。
- **错误处理**：service 返回 `error`，handler 转为 HTTP 状态码；`Get*ByID` 未找到时返回 `(nil, nil)`（非 error），handler 据此返回 `404`。
- **导出/导入**：全量数据导出/导入在 `pkg/db/connection.go`（`ExportAllData` / `ImportAllData`），新增表须同步更新 `ExportData` 结构与导入逻辑（导入须在事务中、按外键依赖顺序清表与插入）。

### Linux / Windows 注意事项

- CI 中 Linux runner 需先装系统依赖（GTK/WebKit2GTK/Soup/JavaScriptCore/zip），详见 `.github/workflows/release.yml`。
- CI 中 Linux/Windows 不指定 `GOTOOLCHAIN`，直接使用 runner 的系统 Go。
- Windows 构建产物依赖 WebView2（Windows 10+ 已内置 Edge WebView2 运行时）。

### `.gitignore`

`.trae/`（IDE 本地元数据）、`build/`、`release/`、`*.sqlite`、`*.log`、`.env*` 等已忽略；**不跟踪** `.trae/` 目录。
