# gvb_server

## 项目定位

`gvb_server` 是 GVB 社区系统的后端服务，基于 Go + Gin 构建，承担以下职责：

- 提供前台社区和后台运营的 HTTP API。
- 管理用户、文章、评论、收藏、菜单、广告、板块、公告等业务数据。
- 提供好友系统、群组、私信、社区悬赏、在线状态、WebSocket 实时通知、语音通话信令。
- 提供 Elasticsearch 全文检索、图片与附件访问控制、操作审计日志、定时任务。

建议配合下面几份文档一起看：

- [前后端总览](./docs/FULLSTACK_OVERVIEW.md)
- [后端入门说明](./BACKEND_BEGINNER_GUIDE.md)
- [接口文档摘要](./API_DOCUMENTATION.md)
- [Swagger 导出文档](./docs/swagger.yaml)

## 技术栈与依赖

| 类别 | 说明 |
| --- | --- |
| 语言 | Go 1.25.5 |
| Web 框架 | Gin |
| ORM | Gorm |
| 数据库 | MySQL |
| 缓存/黑名单 | Redis |
| 搜索 | Elasticsearch 7 |
| 日志 | Logrus + 自定义审计插件 |
| 文档 | Swagger / Postman |
| 实时通信 | WebSocket |

## 启动流程

程序入口在 [`main.go`](./main.go)。正常启动链路如下：

1. `core.InitConf()` 读取 `settings.yaml`。
2. `core.InitLogger()` 初始化日志。
3. `core.InitGorm()` 连接 MySQL。
4. 自动迁移一批运行期模型，并补齐默认板块。
5. 初始化 Redis、Elasticsearch、地址库。
6. 解析命令行参数，如果是一次性命令模式则执行后退出。
7. 初始化定时任务。
8. `routers.InitRouter()` 注册 Gin 路由。
9. 启动 HTTP 服务，并暴露 Swagger 页面。

## 目录结构

| 路径 | 作用 |
| --- | --- |
| [`main.go`](./main.go) | 服务启动入口。 |
| [`config`](./config) | 配置结构定义，对应 `settings.yaml` 的各个 section。 |
| [`core`](./core) | 配置读取、日志、Gorm、Redis、ES、地址库等底层初始化。 |
| [`routers`](./routers) | Gin 路由注册层，定义 URL 和鉴权方式。 |
| [`api`](./api) | Handler 层，解析参数并组织响应。 |
| [`service`](./service) | 业务服务层和定时任务、ES 工具。 |
| [`models`](./models) | 数据模型，既包含 MySQL 模型，也包含 ES 文档结构。 |
| [`middleware`](./middleware) | JWT 鉴权、安全头、操作审计。 |
| [`plugins`](./plugins) | 日志等插件扩展。 |
| [`utils`](./utils) | JWT、密码、文件、响应等工具。 |
| [`docs`](./docs) | Swagger 生成物、Postman 集合、补充文档。 |
| [`flag`](./flag) | 命令行任务入口，如建表、创建用户、ES 导入导出。 |

## 配置说明

默认配置文件是 [`settings.yaml`](./settings.yaml)，主要 section 包括：

- `mysql`
- `logger`
- `system`
- `email`
- `jwt`
- `site_info`
- `qq`
- `qiniu`
- `news`
- `upload`
- `redis`
- `es`

对应的 Go 结构体定义在 [`config/enter.go`](./config/enter.go) 和 `config/conf_*.go`。

### 环境变量覆盖

[`core/conf_core.go`](./core/conf_core.go) 支持用环境变量覆盖敏感配置：

- `GVB_JWT_SECRET`
- `GVB_EMAIL_PASSWORD`
- `GVB_MYSQL_PASSWORD`

也会对 `jwt.secret` 做基础校验，避免明显不安全的配置启动。

## 路由与权限模型

### 路由入口

- Swagger：`/swagger/*any`
- 静态资源：`/uploads/*filepath`
- 业务接口统一挂在：`/api`

路由注册总入口在 [`routers/enter.go`](./routers/enter.go)。

### 鉴权方式

[`middleware/jwt_auth.go`](./middleware/jwt_auth.go) 提供两类权限中间件：

- `JwtAuth()`：普通登录用户即可访问。
- `JwtAdmin()`：仅管理员可访问。

同时兼容两种 token 传法：

- `Authorization: Bearer <token>`
- `token: <token>`

### 主要路由模块

| 路由模块 | 文件 | 说明 |
| --- | --- | --- |
| 设置 | [`routers/settings_router.go`](./routers/settings_router.go) | 站点信息、ES 工具、同步文章/图片。 |
| 图片 | [`routers/images_router.go`](./routers/images_router.go) | 图片上传、列表、权限与元数据。 |
| 文件 | [`routers/file_router.go`](./routers/file_router.go) | 文章附件上传和受控下载。 |
| 广告 | [`routers/advert_router.go`](./routers/advert_router.go) | 广告位管理。 |
| 公告 | [`routers/announcement_router.go`](./routers/announcement_router.go) | 前台公告与后台公告管理。 |
| 菜单 | [`routers/menu_router.go`](./routers/menu_router.go) | 前台导航菜单。 |
| 板块 | [`routers/board_router.go`](./routers/board_router.go) | 板块列表与后台维护。 |
| 用户 | [`routers/user_router.go`](./routers/user_router.go) | 登录、注册、用户资料、用户空间。 |
| 社交 | [`routers/social_router.go`](./routers/social_router.go) | 好友、群组、私信、社区广场、赏金大厅、WebSocket。 |
| 标签 | [`routers/tag_router.go`](./routers/tag_router.go) | 标签维护。 |
| 消息 | [`routers/message_router.go`](./routers/message_router.go) | 历史私信接口。 |
| 文章 | [`routers/article_router.go`](./routers/article_router.go) | 文章 CRUD、收藏、审核、举报、搜索。 |
| 点赞 | [`routers/digg_router.go`](./routers/digg_router.go) | 文章点赞。 |
| 评论 | [`routers/comment_router.go`](./routers/comment_router.go) | 评论和评论点赞。 |
| 资讯 | [`routers/new_router.go`](./routers/new_router.go) | 新闻源聚合。 |
| 聊天室 | [`routers/chat_router.go`](./routers/chat_router.go) | 公共聊天室数据。 |
| 日志 | [`routers/log_router.go`](./routers/log_router.go) | 审计日志列表和删除。 |
| 统计 | [`routers/data_router.go`](./routers/data_router.go) | 首页与后台看板统计。 |

## 核心数据模型

### 文章与搜索

- [`models/article_model.go`](./models/article_model.go) 定义文章 ES 文档结构。
- 全文搜索和文章正文主要依赖 Elasticsearch 索引 `article_index`。
- 评论、举报、收藏、附件等周边能力通过其他 MySQL 模型补充。

### 用户与空间

- [`models/user_model.go`](./models/user_model.go)：用户基础资料、等级、积分、签到信息。
- [`models/user_space_model.go`](./models/user_space_model.go)：用户空间动态和留言。

### 社交与实时通信

- [`models/social_model.go`](./models/social_model.go)：关注、拉黑、在线状态、群组、消息、已读游标、通话记录。
- [`models/community_model.go`](./models/community_model.go)：闲聊广场和赏金大厅的帖子与回复。

### 站点内容

- `board_model.go`：板块
- `menu_model.go`：导航菜单
- `announcement_model.go`：公告
- `banner_model.go`：图片/封面
- `advert_model.go`：广告
- `comment_model.go`：评论
- `message_model.go`：历史私信

## 中间件与运行期机制

### 操作审计

[`middleware/operation_audit.go`](./middleware/operation_audit.go) 会自动记录非只读请求的操作日志，包含：

- 请求方法和路径
- 请求体快照
- 响应业务码和 HTTP 状态码
- 耗时

### 静态资源保护

[`routers/enter.go`](./routers/enter.go) 中的 `/uploads/*filepath` 不是简单目录映射，而是做了：

- 路径清洗，防止目录穿越。
- 受保护目录拦截。
- 私有图片所有权判断。
- 附件必须走下载接口而不是直接裸链访问。

### 定时任务

[`service/cron_ser`](./service/cron_ser) 负责文章统计、评论统计等后台同步任务。

## 命令行模式

[`flag/enter.go`](./flag/enter.go) 定义了几类一次性命令：

```bash
go run . -db
go run . -user admin
go run . -user user
go run . -es create
go run . -es export
go run . -es import
go run . -es sync-fulltext
```

如果命中这些参数，程序会执行完任务后直接退出，不再启动 Web 服务。

## 本地运行

### 1. 准备依赖

建议至少准备：

- MySQL
- Redis
- Elasticsearch

### 2. 检查配置

按你的本地环境修改 [`settings.yaml`](./settings.yaml) 中的数据库、Redis、ES、邮件和 JWT 配置。

### 3. 启动服务

```bash
go run .
```

默认监听：`http://127.0.0.1:8080`

Swagger 页面：`http://127.0.0.1:8080/swagger/index.html#/`

## 如何阅读这个项目

推荐顺序：

1. [`main.go`](./main.go)
2. [`routers/enter.go`](./routers/enter.go)
3. 某个具体 `routers/*.go`
4. 对应 `api/<module>_api`
5. 需要时继续看 `service` 和 `models`

如果你是从前端某个页面反查后端，直接对照：

`src/api/*.js -> /api 路由 -> routers/*.go -> api/* -> service/* -> models/*`

## 现有文档怎么配合看

- [前后端总览](./docs/FULLSTACK_OVERVIEW.md)：适合理清两端模块和调用链。
- [后端入门说明](./BACKEND_BEGINNER_GUIDE.md)：适合从后端视角快速入门。
- [Swagger YAML](./docs/swagger.yaml)：适合核对接口字段。
- [Postman 集合](./docs/GVB_Server.postman_collection.json)：适合直接联调。

## 一句话总结

这个后端的主线非常明确：

`main -> core 初始化 -> routers 注册 -> middleware 鉴权/审计 -> api 处理参数 -> service 组织业务 -> models 落到 MySQL/Redis/ES`
