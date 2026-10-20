# slb-api

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
  dsn: root:slb123456@tcp(127.0.0.1:33306)/slb_reservoir?charset=utf8mb4&parseTime=True&loc=Local
```

本地数据由 Docker 容器 `slb-mysql` 提供（水库结构化数据，库 `slb_reservoir`，含
`reservoir_feature_attr` / `reservoir_capacity_curve` / `reservoir_discharge_curve` 三张表，
由 `/Users/jiang/slb/scripts/extract_reservoir_to_mysql.py` 入库，参见该目录下 README）。

## 接口

### POST /api/reservoir/query —— 水库数据查询

按站点编号数组 + 数据类型数组查询已入库的结构化数据。

请求体：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| stcd | string[] | 是 | 站点编号数组，如 `["7I505060","81811813"]` |
| slb-type | string[] | 否 | 数据类型数组，可选值 `attr`（特征属性）/ `capacity`（库容曲线）/ `discharge`（泄流曲线）；未传或为空数组时默认查询全部三类 |

```bash
curl -s -X POST http://127.0.0.1:8083/api/reservoir/query \
  -H 'Content-Type: application/json' \
  -d '{"stcd":["7I505060","81811813"],"slb-type":["attr","capacity","discharge"]}'
```

响应（`data.list` 与请求 `stcd` 顺序一致；每个站点的 `attr`/`capacity`/`discharge`
三个字段固定返回，未请求的类型或查无数据时为空数组）：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "stcd": "7I505060",
        "attr": [
          {
            "region_code": "331123000197",
            "file_name": "2023年度成屏二级水库控制运用计划报告",
            "source_file": "aerial",
            "table_index": 0,
            "category": "工程位置",
            "attr_name": "工程地点",
            "attr_value": "…",
            "unit": "",
            "remark": ""
          }
        ],
        "capacity": [
          {
            "region_code": "331123000197",
            "file_name": "2023年度成屏二级水库控制运用计划报告",
            "source_file": "capacity",
            "table_index": 1,
            "seq": 0,
            "water_level": 230,
            "capacity": 0,
            "capacity_unit": "万m³"
          }
        ],
        "discharge": [
          {
            "water_level": 263,
            "discharge": 0,
            "discharge_unit": "m³/s",
            "seq": 0
          }
        ]
      }
    ]
  }
}
```

错误：参数缺失/为空/类型非法返回 `{"code":400,"msg":"…"}`。

### GET /api/wechat-user/list —— 微信用户列表（示例接口）

原有示例接口，风格参考。