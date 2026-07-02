# Tool-Agent

基于 Go 语言开发的工具服务，提供 HTTP 接口与定时任务（cron）能力。项目采用分层架构（Handler → Service → DAO → Model），通过 `urfave/cli` 将 HTTP 服务、定时任务、数据库迁移、代码生成统一为同一二进制的不同子命令。

## 技术栈

| 能力 | 依赖 | 说明 |
|------|------|------|
| 语言 | Go 1.23 | |
| Web 框架 | [gin](https://github.com/gin-gonic/gin) | HTTP 路由、中间件 |
| ORM | [GORM](https://gorm.io) + [gorm/gen](https://gorm.io/gen) | gen 生成类型安全的 DAO 查询代码 |
| 数据库 | MySQL | 驱动 `gorm.io/driver/mysql`，支持 `dbresolver` 读写分离 |
| 配置 | [viper](https://github.com/spf13/viper) | YAML 配置 + 环境变量覆盖（`.` → `_`）|
| 日志 | [zap](https://go.uber.org/zap) + lumberjack | 结构化日志、文件轮转，支持 Sentry 上报 |
| 命令行 | [urfave/cli/v2](https://github.com/urfave/cli/tree/v2) | 子命令编排（serve / worker / db / generate）|
| 定时任务 | [robfig/cron/v3](https://github.com/robfig/cron) | 5 段式 cron 表达式 |
| 请求追踪 | uuid | 每个请求注入 `x-request-id` |

## 目录结构

```
tool-agent/
├── main.go                     # 入口，cli.App 装配
├── config/app.yml              # 配置文件
├── commands/                   # 所有 CLI 子命令
│   ├── all.go                  #   注册所有子命令
│   ├── serve.go                #   启动 HTTP 服务（默认 action）
│   ├── worker.go               #   启动定时任务 (worker start)
│   ├── db.go                   #   数据表迁移 (db migrate)
│   ├── gen.go                  #   生成 DAO 代码 (generate)
│   └── demo/                   #   示例定时任务
│       ├── init.go             #     Worker + GracefulStart
│       └── demo.go             #     实现 cron.Job 的任务体
├── server/http/
│   ├── handlers/               # HTTP 处理器（按业务域分包）
│   │   ├── handler.go          #   根路由，挂载 /api 分组
│   │   └── picture_book/       #   绘本接口
│   ├── httputil/               #   分页等工具
│   └── response/               #   统一响应结构
├── internal/
│   ├── service/                # 业务逻辑层（按业务域分包）
│   │   └── picture_book/
│   ├── dao/                    # gorm/gen 自动生成的查询代码（勿手改）
│   └── common/                 #   ServiceResult、ServiceError 等公共类型
├── model/                      # GORM 模型定义
│   ├── base.go                 #   基础字段（id / 时间）
│   └── s_picture_book.go
├── utils/                      # 基础设施工具
│   ├── viper.go                #   配置加载
│   ├── log.go                  #   zap 日志初始化
│   ├── db.go                   #   GORM 初始化
│   ├── gin.go                  #   HTTP Server 与优雅启停
│   └── graceful/               #   定时任务生命周期管理（Start/Wait）
├── Dockerfile                  # 多阶段构建
└── docker-compose.yml          # web + worker 两个服务
```

## 环境要求

- Go >= 1.23
- MySQL 5.7+ / 8.0

## 快速开始

### 1. 准备配置

编辑 `config/app.yml`，把 `tools.dsn` 指向一个可连接的 MySQL：

```yaml
tools:
  dialect: mysql
  dsn: root:123456@tcp(127.0.0.1:3307)/tool-agent?charset=utf8mb4&parseTime=True&loc=Local
```

> 提前在 MySQL 中创建数据库 `tool-agent`（`CREATE DATABASE \`tool-agent\` DEFAULT CHARSET utf8mb4;`）。

### 2. 生成数据表

```bash
go run main.go db migrate
```

该命令对 `commands/db.go` 中注册的所有模型执行 GORM `AutoMigrate`，自动创建/更新数据表结构。

### 3. 启动 HTTP 服务

```bash
go run main.go
# 或显式：go run main.go serve
```

默认监听 `:8080`。健康检查：`curl http://localhost:8080/health` 返回 `ok`。

## 修改项目模块名

当前 Go module 名为 `tool-agent`（`go.mod` 第一行）。所有 .go 文件的 import 路径都以 `tool-agent/` 为前缀。如需改成自己的名字（例如 `github.com/yourname/agent`），按以下步骤。

### 1. 查看当前模块名

```bash
go list -m          # 输出当前 module 路径
# 或直接看 go.mod 第一行
head -1 go.mod
```

### 2. 修改模块名

假设新模块名为 `new-name`：

```bash
# (1) 更新 go.mod 的 module 声明
go mod edit -module new-name

# (2) 替换所有 .go 文件中的 import 路径前缀（tool-agent/ → new-name/）
#     macOS 的 sed 需要在 -i 后加空串：
find . -type f -name "*.go" -exec sed -i '' 's|tool-agent/|new-name/|g' {} +
#     Linux：
find . -type f -name "*.go" -exec sed -i 's|tool-agent/|new-name/|g' {} +

# (3) 整理依赖
go mod tidy
```

> `sed` 用 `|` 作分隔符，因此新模块名即便含 `/`（如 `github.com/yourname/agent`）也能正常替换。

### 3. 验证

```bash
go build ./...      # 编译通过即说明所有 import 路径都已替换
go list -m          # 确认已变为 new-name
```

> ⚠️ `go mod edit -module` 只改 `go.mod` 一处，**不会**自动改 .go 文件里的 import 路径，所以第 (2) 步全局替换不能省。建议替换前用 `git status` 确认变更范围，出问题可 `git checkout` 回退。
>
> 补充：`main.go` 里 `utils.InitViper("tool-agent", ...)` 的入参和 `config/app.yml` 里的 `app:` 字段与 Go module 名是**相互独立**的，按需另行修改即可。

## 可用命令

| 命令 | 说明 |
|------|------|
| `go run main.go` | 启动 HTTP 服务（默认 action）|
| `go run main.go worker start` | 启动所有定时任务，收到 `Ctrl+C`/SIGTERM 优雅退出 |
| `go run main.go db migrate` | 自动创建/更新数据表 |
| `go run main.go generate` | 根据 model 生成 DAO 类型安全查询代码 |
| `go run main.go --help` | 查看所有命令 |

`--config <path>` 可指定其他配置文件。

## 数据表与 DAO 的生成流程

项目用 [gorm/gen](https://gorm.io/gen) 从 model 结构体反向生成类型安全的查询代码（产物在 `internal/dao/`，**不要手改**）。新增一张表的完整流程：

**1. 定义模型** —— 在 `model/` 下新建文件，例如 `model/s_xxx.go`：

```go
package model

type SXxx struct {
    BaseModelFieldId                          // 主键 id
    Name string `gorm:"column:name;type:varchar(255);comment:名称;NOT NULL" json:"name"`
    BaseModelFieldTime                        // created_at / updated_at / deleted_at
}

func (m *SXxx) TableName() string {
    return "s_xxx"
}
```

`BaseModelFieldId` 与 `BaseModelFieldTime` 定义在 `model/base.go`，提供主键与时间字段。

**2. 注册到迁移命令** —— 在 `commands/db.go` 的 `AutoMigrate` 中追加：

```go
db.AutoMigrate(
    &model.SPictureBook{},
    &model.SXxx{},            // 新增
)
```

**3. 注册到代码生成** —— 在 `commands/gen.go` 的 `ApplyBasic` 中追加：

```go
g.ApplyBasic(
    model.SPictureBook{},
    model.SXxx{},             // 新增
)
```

**4. 执行**：

```bash
go run main.go db migrate    # 在数据库中创建/更新表
go run main.go generate      # 生成 internal/dao/s_xxx.gen.go
```

**5. 在 service 中使用生成的 DAO**：

```go
func NewXxxService() *Service {
    s := &Service{db: utils.DB()}
    dao.SetDefault(utils.DB())   // 全局注入生成的查询对象
    return s
}

// 类型安全的链式查询
q := dao.Q
info, err := q.SXxx.WithContext(ctx).Where(q.SXxx.Name.Eq("foo")).First()
```

## 新增一个业务模块（HTTP 接口）

按现有 `picture_book` 域的分层套路复制即可：

1. **Model**：`model/s_xxx.go`（见上节），并执行 migrate + generate。
2. **Service**：在 `internal/service/xxx/` 下建包
   - `interface.go` —— 定义 `ServiceIFace` 接口
   - `impl.go` —— `Service` 结构体持有 `*gorm.DB`，`NewXxxService()` 构造函数内 `dao.SetDefault(utils.DB())`
   - `impl_create.go` / `impl_update.go` / `impl_delete.go` / `impl_list.go` —— 每个方法一个文件
3. **Handler**：在 `server/http/handlers/xxx/impl.go` 建 `XxxHandler`，`RegisterRoutes(routerGroup)` 注册路由。
4. **挂载**：在 `server/http/handlers/handler.go` 的 `RegisterRoutes()` 中调用 `xxx.NewXxxHandler(...).RegisterRoutes(g)`。

## 新增一个定时任务

参考 `commands/demo/`：

1. 复制 `commands/demo/` 改名（例如 `commands/mytask/`），改包名。
2. `init.go`：`Worker*` 结构体 + `InitWorker*()` 构造（用 `cron.WithChain(cron.DelayIfStillRunning(...))` 防止任务重叠）+ `GracefulStart(ctx)`，cron 表达式从 viper 读取。
3. 任务体：实现 `cron.Job` 的 `Run()` 方法，业务逻辑可直接调用 `internal/service/*`。
4. 在 `commands/worker.go` 增加一行 `graceful.Start(mytask.InitWorkerMytask())`。
5. 在 `config/app.yml` 增加对应 cron 表达式配置。

```bash
go run main.go worker start    # 启动定时任务
```

> 定时任务进程必须只跑一个副本（cron 在进程内调度，多副本会导致任务重复执行）。

## 配置说明

`config/app.yml`：

```yaml
debug: true
app: tool-agent
server:
  addr: ":8080"
log:
  development: true
  level: debug
  disable_sentry: true
  outputs:
    - stdout
    - ./logs/tool-agent.log
tools:
  dialect: mysql
  dsn: root:123456@tcp(127.0.0.1:3307)/tool-agent?charset=utf8mb4&parseTime=True&loc=Local
demo:                         # 示例定时任务
  cron: "*/1 * * * *"
```

任何配置项都可用环境变量覆盖（viper `AutomaticEnv`，`.` 替换为 `_`），例如：

```bash
tools_dsn="user:pass@tcp(host:3306)/db" go run main.go
```

## 部署

### Docker

镜像构建：

```bash
docker build -t tool-agent .
```

### docker-compose

`docker-compose.yml` 用同一镜像跑两个服务：`web`（HTTP）和 `worker`（定时任务）。

```bash
docker compose up -d          # 同时启动 web + worker
docker compose logs -f worker
docker compose down
```

> 容器内 `127.0.0.1` 指容器自身，若 MySQL 在宿主机上，需用环境变量 `tools_dsn=...` 指向可达地址。

## API 接口

所有业务接口统一前缀 `/api`，响应结构：

```json
{ "code": 0, "msg": "操作成功", "data": {} }
```

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/ping` | 存活探针，返回 `pong` |
| GET | `/health` | 健康检查，返回 `ok` |
| POST | `/api/picture_book/create` | 创建绘本 |
| POST | `/api/picture_book/update` | 更新绘本 |
| POST | `/api/picture_book/delete` | 删除绘本 |
| GET | `/api/picture_book/list` | 绘本列表 |
```
