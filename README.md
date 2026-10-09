# web-app

基于 Go 语言开发的 Web 服务，提供 HTTP 接口与定时任务（cron）能力，当前内置微信用户（`s_wechat_user`）查询接口。项目采用分层架构（Handler → Service → DAO → Model），通过 `urfave/cli` 将 HTTP 服务、定时任务、数据库迁移、代码生成统一为同一二进制的不同子命令。

## 技术栈

| 能力 | 依赖 | 说明 |
|------|------|------|
| 语言 | Go 1.23 | |
| Web 框架 | [gin](https://github.com/gin-gonic/gin) | HTTP 路由、中间件、优雅启停 |
| ORM | [GORM](https://gorm.io) + [gorm/gen](https://gorm.io/gen) | gen 生成类型安全的 DAO 查询代码（含 `Read()`/`Write()` 子句，基于 dbresolver）|
| 数据库 | MySQL | 驱动 `gorm.io/driver/mysql` |
| 配置 | [viper](https://github.com/spf13/viper) | YAML 配置 + 环境变量覆盖（`.` → `_`）|
| 日志 | [zap](https://go.uber.org/zap) + lumberjack | 控制台彩色输出 + 文件 JSON 轮转，Error 级别可上报 Sentry |
| 命令行 | [urfave/cli/v2](https://github.com/urfave/cli/tree/v2) | 子命令编排（serve / worker / db / generate）|
| 定时任务 | [robfig/cron/v3](https://github.com/robfig/cron) | 5 段式 cron 表达式，`DelayIfStillRunning` 防止任务重叠 |
| 请求追踪 | uuid | 每个请求注入 `x-request-id`，日志自动携带 `request_id` |

## 目录结构

```
web/
├── main.go                     # 入口，cli.App 装配（默认 action = serve）
├── config/app.yml              # 配置文件
├── commands/                   # 所有 CLI 子命令
│   ├── all.go                  #   注册所有子命令
│   ├── serve.go                #   启动 HTTP 服务（默认 action）
│   ├── worker.go               #   启动定时任务 (worker start)
│   ├── db.go                   #   数据表迁移 (db migrate)
│   ├── gen.go                  #   生成 DAO 代码 (generate)
│   └── demo/                   #   示例定时任务
│       ├── init.go             #     WorkerDemo + GracefulStart（含 switch 开关）
│       └── demo.go             #     实现 cron.Job 的任务体
├── server/http/
│   └── handlers/               # HTTP 处理器（按业务域分包）
│       ├── handler.go          #   根路由，挂载 /api 分组
│       └── wechat/             #   微信用户接口
│           ├── impl.go         #     WechatUserHandler + 路由注册
│           └── impl_user_list.go
├── internal/
│   ├── service/                # 业务逻辑层（按业务域分包）
│   │   └── wechat/
│   │       ├── interface.go    #     ServiceIFace 接口定义
│   │       ├── impl.go         #     Service 构造 + dao.SetDefault
│   │       └── impl_user_list.go
│   ├── dao/                    # gorm/gen 自动生成的查询代码（勿手改）
│   └── common/result.go        # ServiceResult、ServiceError 等公共类型
├── model/                      # GORM 模型定义
│   ├── base.go                 #   基础字段（id / 时间）
│   └── s_wechat_user.go
├── utils/                      # 基础设施工具
│   ├── viper.go                #   配置加载
│   ├── log.go                  #   zap 日志初始化
│   ├── db.go                   #   GORM 初始化
│   ├── gin.go                  #   HTTP Server、RequestID 中间件与优雅启停
│   ├── graceful/               #   定时任务生命周期管理（Start/Wait）
│   └── tool/
│       ├── httputil/           #   分页参数绑定、400/500 错误响应
│       ├── response/           #   统一响应结构（code/msg/data）
│       └── middleware/         #   SSE (text/event-stream) 响应头中间件
├── Dockerfile                  # 多阶段构建，时区 Asia/Shanghai
└── docker-compose.yml          # 单服务部署，挂载本地配置
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
  dsn: root:123456@tcp(127.0.0.1:3306)/wechat?charset=utf8mb4&parseTime=True&loc=Local
```

> 提前在 MySQL 中创建数据库 `wechat`（`CREATE DATABASE wechat DEFAULT CHARSET utf8mb4;`）。

### 2. 生成数据表

```bash
go run main.go db migrate
```

该命令对 `commands/db.go` 中注册的所有模型执行 GORM `AutoMigrate`，自动创建/更新数据表结构（当前为 `s_wechat_user`）。

### 3. 启动 HTTP 服务

```bash
go run main.go
# 或显式：go run main.go serve
```

默认监听 `:8080`。验证：

```bash
curl http://localhost:8080/health
# ok
curl http://localhost:8080/api/wechat-user/list
```

## 可用命令

| 命令 | 说明 |
|------|------|
| `go run main.go` | 启动 HTTP 服务（默认 action）|
| `go run main.go worker start` | 启动所有定时任务，收到 `Ctrl+C`/SIGTERM 优雅退出 |
| `go run main.go db migrate` | 自动创建/更新数据表 |
| `go run main.go generate` | 根据 model 生成 DAO 类型安全查询代码 |
| `go run main.go --version` | 查看版本（本地构建为 `local`，Docker 构建为 `docker`）|
| `go run main.go --help` | 查看所有命令 |

`--config <path>` 可指定其他配置文件（默认读取 `config/app.yml`）。

## API 接口

所有业务接口统一前缀 `/api`，响应结构：

```json
{
  "code": 0,
  "msg": "success",
  "data": { },
  "count": 1,
  "limit": 10,
  "offset": 0
}
```

`code` 为 0 表示成功，400 参数错误，500 内部错误；分页字段仅在列表接口返回。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/ping` | 存活探针，返回 `pong` |
| GET | `/health` | 健康检查，返回 `ok` |
| GET | `/api/wechat-user/list` | 微信用户列表：`score > 0` 且 `device_id` 非空，按 `score` 倒序取前 10 条 |

`/api/wechat-user/list` 返回的 `data` 结构：

```json
{
  "list": [
    { "user_name": "alice", "device_id": "abc123", "score": 100 }
  ]
}
```

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
    &model.SWechatUser{},
    &model.SXxx{},            // 新增
)
```

**3. 注册到代码生成** —— 在 `commands/gen.go` 的 `ApplyBasic` 中追加：

```go
g.ApplyBasic(
    model.SWechatUser{},
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

// 类型安全的链式查询（可参考 internal/service/wechat/impl_user_list.go）
xxxDao := dao.SXxx
list, total, err := xxxDao.
    Where(xxxDao.Name.Eq("foo")).
    Order(xxxDao.Id.Desc()).
    FindByPage(0, 10)
```

## 新增一个业务模块（HTTP 接口）

按现有 `wechat` 域的分层套路复制即可：

1. **Model**：`model/s_xxx.go`（见上节），并执行 migrate + generate。
2. **Service**：在 `internal/service/xxx/` 下建包
   - `interface.go` —— 定义 `XxxServiceIFace` 接口，方法签名统一为 `(ctx context.Context, ...) (common.ServiceResult, error)`
   - `impl.go` —— `Service` 结构体持有 `*gorm.DB`，`NewXxxService()` 构造函数内 `dao.SetDefault(utils.DB())`
   - `impl_<方法名>.go` —— 每个方法一个文件（如 `impl_user_list.go`）
3. **Handler**：在 `server/http/handlers/xxx/impl.go` 建 `XxxHandler`，`RegisterRoutes(routerGroup)` 注册路由；每个接口方法单独一个文件（如 `impl_user_list.go`）。响应统一走 `utils/tool/response`：成功 `response.Successful` / `SuccessfulWithPagination`，失败 `response.Failed` / `ParameterError` / `InternalError`；分页参数可用 `httputil.Pagination` 绑定（`limit` 默认 10、`offset` 默认 0）。
4. **挂载**：在 `server/http/handlers/handler.go` 的 `RegisterRoutes()` 中调用 `xxx.NewXxxHandler(h.router).RegisterRoutes(g)`。

## 新增一个定时任务

参考 `commands/demo/`：

1. 复制 `commands/demo/` 改名（例如 `commands/mytask/`），改包名。
2. `init.go`：`Worker*` 结构体 + `InitWorker*()` 构造（用 `cron.WithChain(cron.DelayIfStillRunning(...))` 防止任务重叠）+ `GracefulStart(ctx)`；cron 表达式从 viper 读取，建议像 demo 一样加 `switch` 配置项作为总开关（`false` 时跳过注册）。
3. 任务体：实现 `cron.Job` 的 `Run()` 方法，业务逻辑可直接调用 `internal/service/*`。
4. 在 `commands/worker.go` 增加一行 `graceful.Start(mytask.InitWorkerMytask())`。
5. 在 `config/app.yml` 增加对应 `switch` 与 cron 表达式配置。

```bash
go run main.go worker start    # 启动定时任务
```

> 定时任务进程必须只跑一个副本（cron 在进程内调度，多副本会导致任务重复执行）。

## 配置说明

`config/app.yml`：

```yaml
debug: true                     # 调试模式：gin debug 模式 + GORM 打印 SQL
app: web-app
server:
  addr: ":8080"                 # HTTP 监听地址
log:
  development: true
  level: debug                  # debug / info / warn / error
  disable_sentry: true          # false 时 Error 级别日志会上报 Sentry
  outputs:                      # 多路输出，stdout 为彩色控制台，其余为 JSON 文件
    - stdout
    - ./logs/web-app.log
tools:
  dialect: mysql
  dsn: root:123456@tcp(127.0.0.1:3306)/wechat?charset=utf8mb4&parseTime=True&loc=Local
demo:                           # 示例定时任务
  switch: false                 # 总开关，false 时不注册 cron
  cron: "*/1 * * * *"
```

任何配置项都可用环境变量覆盖（viper `AutomaticEnv`，`.` 替换为 `_`），例如：

```bash
tools_dsn="user:pass@tcp(host:3306)/db" go run main.go
```

## 部署

### Docker

镜像为多阶段构建：builder 阶段基于 `$BUILDPLATFORM` + `TARGETOS/TARGETARCH` 支持跨平台编译（如 `docker buildx build --platform linux/amd64`），并使用 BuildKit cache mount 加速重复构建；运行阶段基于 alpine、以非 root 用户运行，内置 `ca-certificates` 与 `Asia/Shanghai` 时区。`.dockerignore` 将 `.git/`、`logs/`、`tmp/` 及含 DSN 凭据的 `config/` 排除在构建上下文外，**镜像内不含配置文件**，运行时必须挂载配置或注入环境变量：

```bash
docker build -t web-app .
docker run -p 8080:8080 -v ./config/app.yml:/app/config/app.yml web-app
```

### docker-compose

`docker-compose.yml` 基于同一镜像定义两个服务（共享配置通过 YAML 锚点合并），均带 `restart: unless-stopped`、35s 停止宽限期（代码内 HTTP 优雅关闭最长 30s，需大于 Docker 默认的 10s）与日志轮转限制：

- `web` —— HTTP 服务，映射 8080 端口，健康检查走 `GET /health`
- `worker` —— 定时任务进程，归属 `worker` profile 默认不启动（cron 为进程内调度，只允许单副本）

DSN 默认指向 `host.docker.internal:3306`（宿主机 MySQL，Linux 上由 `extra_hosts` 保证该域名可用），可用 `TOOLS_DSN` 环境变量或同目录 `.env` 文件覆盖，优先级高于挂载的配置文件：

```bash
docker network create common-app-net   # 若外部网络不存在需先创建
docker compose up -d                   # 只启动 web
docker compose --profile worker up -d  # 启动 web + worker
docker compose logs -f web
docker compose down
```

## CI/CD 持续部署

项目内置 GitLab CI 流水线（[.gitlab-ci.yml](.gitlab-ci.yml)）与服务器部署脚本（[deploy/](deploy/)）：推送 `main` 分支自动完成 编译检查 → 构建镜像推送私有仓库 → SSH 到服务器更新容器（含自动 `db migrate`、健康检查失败自动回滚）。完整的从零搭建步骤见 [deploy/README.md](deploy/README.md)。

## 修改项目模块名

当前 Go module 名为 `web`（`go.mod` 第一行）。所有 .go 文件的 import 路径都以 `web/` 为前缀。如需改成自己的名字（例如 `github.com/yourname/agent`），按以下步骤。

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

# (2) 替换所有 .go 文件中的 import 路径前缀（web/ → new-name/）
#     macOS 的 sed 需要在 -i 后加空串：
find . -type f -name "*.go" -exec sed -i '' 's|web/|new-name/|g' {} +
#     Linux：
find . -type f -name "*.go" -exec sed -i 's|web/|new-name/|g' {} +

# (3) 整理依赖
go mod tidy
```

> `sed` 用 `|` 作分隔符，因此新模块名即便含 `/`（如 `github.com/yourname/agent`）也能正常替换。`web/` 前缀较短，替换后建议 `git diff` 检查一遍有无误伤。

### 3. 验证

```bash
go build ./...      # 编译通过即说明所有 import 路径都已替换
go list -m          # 确认已变为 new-name
```

> ⚠️ `go mod edit -module` 只改 `go.mod` 一处，**不会**自动改 .go 文件里的 import 路径，所以第 (2) 步全局替换不能省。建议替换前用 `git status` 确认变更范围，出问题可 `git checkout` 回退。
>
> 补充：`main.go` 里 `utils.InitViper("web-app", ...)` 的入参、`config/app.yml` 里的 `app:` 字段、日志文件名 `./logs/web-app.log`，以及 `Dockerfile`/`docker-compose.yml` 中的二进制名 `web-app`，与 Go module 名是**相互独立**的，按需另行修改即可。
