# GVB Server Beginner Guide

这份文档是给第一次接触 `gvb_server` 的新手准备的。

后端文件数量很多，所以不要一开始就试图“按文件名从上往下看完”。
更好的读法是先建立大图，再按链路进入。

这份文档会先回答：

1. 后端从哪里启动？
2. 一次 HTTP 请求在后端里怎么流动？
3. 图片上传、附件下载、好友消息这些功能分别走哪些模块？
4. 每个顶层目录负责什么？
5. `api / routers / service / models / utils` 之间是什么关系？

## 1. 先看什么

建议按这个顺序读：

1. `main.go`
2. `core/conf_core.go`
3. `global/global.go`
4. `routers/enter.go`
5. `middleware/jwt_auth.go`
6. `api/enter.go`
7. 一个具体业务模块，例如：
   - 图片：`api/images_api/image_upload.go`
   - 附件：`api/file_api/file_upload.go`、`api/file_api/file_download.go`
   - 社交：`api/social_api/*.go`

## 2. 后端启动链路

程序启动后大致按下面顺序执行：

1. `main.go`
2. `core.InitConf()`
   - 读取 `settings.yaml`
   - 覆盖敏感环境变量
   - 做安全校验
3. `core.InitLogger()`
4. `core.InitGorm()`
5. `AutoMigrate(...)`
   - 自动迁移数据库表
6. `core.ConnectRedis()`
7. `core.EsConnect()`
8. `core.InitAddrDB()`
9. `flag.Parse()`
   - 如果是命令行任务模式，执行任务后退出
10. `cron_ser.CronInit()`
11. `routers.InitRouter()`
12. `router.Run(addr)`

## 3. 一次 HTTP 请求如何流动

以一个普通接口为例，请求大致会走下面这条链路：

1. 浏览器请求某个 `/api/...` 地址
2. `routers/*.go` 里注册的路由匹配到目标 handler
3. 先经过中间件，例如：
   - `middleware.JwtAuth()`
   - `middleware.JwtAdmin()`
   - `middleware.OperationAudit()`
4. 进入 `api/<module>_api/*.go` 的具体处理函数
5. API 层做：
   - 参数解析
   - 基础校验
   - 权限判断
   - 调用 service / model / utils
6. 数据通常会落到：
   - MySQL（Gorm）
   - Redis
   - Elasticsearch
   - 本地磁盘上传目录
7. 最后通过 `models/res` 里的统一响应结构返回给前端

## 4. 顶层目录地图

### `api`

这里可以理解为“控制器层 / Handler 层”。
每个子目录一般对应一个业务模块，例如文章、用户、图片、社交等。

常见文件命名规则：

- `enter.go`: 模块聚合入口
- `xxx_create.go`: 创建
- `xxx_list.go`: 列表
- `xxx_update.go`: 更新
- `xxx_remove.go`: 删除
- `xxx_detail.go`: 详情
- `xxx_helper.go`: 仅给当前模块复用的辅助函数

主要模块：

- `api/advert_api`: 广告接口
- `api/announcement_api`: 公告接口
- `api/article_api`: 文章、收藏、全文搜索、审核、举报等
- `api/board_api`: 板块接口
- `api/chat_api`: 公共聊天室接口
- `api/comment_api`: 评论接口
- `api/data_api`: 数据统计接口
- `api/digg_api`: 点赞接口
- `api/file_api`: 文章附件上传/下载
- `api/images_api`: 图片上传、图片列表、图片删除、图片重命名
- `api/log_api`: 日志接口
- `api/menu_api`: 菜单接口
- `api/message_api`: 消息接口
- `api/new_api`: 资讯接口
- `api/settings_api`: 系统配置接口
- `api/social_api`: 好友、黑名单、私信、群组、在线状态、WebSocket、语音信令
- `api/tag_api`: 标签接口
- `api/user_api`: 登录、注册、资料、空间、签到、权限相关接口

### `config`

定义配置结构体，对应 `settings.yaml` 的字段结构。

例如：

- `conf_mysql.go`: MySQL 配置
- `conf_logger.go`: 日志配置
- `conf_system.go`: 服务监听地址等系统配置
- `conf_upload.go`: 上传目录和大小限制
- `conf_site_info.go`: 站点信息
- `enter.go`: 聚合总配置结构

### `core`

这里可以理解为“框架初始化层”。
负责初始化程序运行所需的底层依赖。

- `conf_core.go`: 读配置
- `gorm.go`: 初始化数据库
- `redis.go`: 初始化 Redis
- `es.go`: 初始化 Elasticsearch
- `logrus.go`: 初始化日志
- `addr_db.go`: 初始化 IP 地址库

### `flag`

命令行工具模式。
如果不是正常启动 Web 服务，而是执行一次性任务（例如建表、创建用户），通常会从这里走。

### `global`

全局共享变量。

例如：

- `global.Config`
- `global.DB`
- `global.Redis`
- `global.ESClient`
- `global.Log`

### `middleware`

Gin 中间件层。

- `jwt_auth.go`: 登录鉴权 / 管理员鉴权
- `operation_audit.go`: 操作审计
- `security_headers.go`: 安全响应头

### `models`

数据模型层。

这里放两大类东西：

1. 数据表模型（Gorm Model）
2. 通用类型与响应结构

主要文件：

- `article_model.go`: 文章模型
- `banner_model.go`: 图片模型
- `board_model.go`: 板块模型
- `comment_model.go`: 评论模型
- `message_model.go`: 消息模型
- `social_model.go`: 社交模型
- `user_model.go`: 用户模型
- `user_space_model.go`: 用户空间模型
- `article_file_model.go`: 附件模型
- `announcement_model.go`: 公告模型
- `article_report_model.go`: 举报模型
- `models/ctype/*`: 各种枚举/自定义类型
- `models/res/*`: 统一响应结构与错误码

### `plugins`

第三方能力接入层。

- `plugins/email`: 邮件发送
- `plugins/log_stash`: 操作日志模型与辅助
- `plugins/qiniu`: 七牛云
- `plugins/qq`: QQ 登录相关

### `routers`

路由注册层。

每个文件负责把某个业务模块的 API handler 注册到 Gin 路由上。

- `routers/enter.go`: 总入口
- `advert_router.go`: 广告路由
- `announcement_router.go`: 公告路由
- `article_router.go`: 文章路由
- `board_router.go`: 板块路由
- `chat_router.go`: 公共聊天路由
- `comment_router.go`: 评论路由
- `data_router.go`: 统计路由
- `digg_router.go`: 点赞路由
- `file_router.go`: 附件路由
- `images_router.go`: 图片路由
- `log_router.go`: 日志路由
- `menu_router.go`: 菜单路由
- `message_router.go`: 消息路由
- `new_router.go`: 资讯路由
- `settings_router.go`: 配置路由
- `social_router.go`: 社交路由
- `tag_router.go`: 标签路由
- `user_router.go`: 用户路由

### `service`

服务层 / 业务逻辑层。

不是所有 API 都一定会经过 service，但一旦逻辑开始变复杂，通常就会抽到这里。

主要目录：

- `service/board_ser`: 板块初始化等
- `service/common`: 通用列表等服务
- `service/crawl_ser`: 内容/图片抓取与同步
- `service/cron_ser`: 定时任务
- `service/es_ser`: ES 搜索相关
- `service/image_ser`: 图片上传规则、白名单等
- `service/redis_ser`: Redis 相关逻辑，例如 token 退出、计数、点赞同步
- `service/user_ser`: 用户默认头像、默认资料等逻辑

### `utils`

通用工具函数层。

主要能力：

- `utils/jwts`: JWT 生成与解析
- `utils/pwd`: 密码加密
- `utils/sanitize`: 内容与 URL 清洗
- `utils/requests`: HTTP 请求封装
- `utils/random`: 随机码 / 随机字符串
- `utils/md5.go`: 哈希
- `utils/get_addr_by_ip.go`: IP 地址处理

## 5. 路由、API、Service、Model 的关系

这四层是新手最容易搞混的。

你可以这样记：

1. `routers`
   只负责“这个 URL 交给谁处理”
2. `api`
   负责“接住请求、解析参数、做基础校验、返回响应”
3. `service`
   负责“复杂业务逻辑”
4. `models`
   负责“数据结构和数据持久化映射”

一个简化例子：

1. `routers/images_router.go` 注册 `/api/images`
2. `api/images_api/image_upload.go` 接住请求
3. 里面调用白名单、哈希、路径检查等逻辑
4. 最后写入 `models.BannerModel`

## 6. 图片上传链路

建议把这条链路单独看懂，因为它同时涉及前端、路由、静态文件、数据库。

### 6.1 上传时

1. 前端把文件以 `multipart/form-data` 发送到 `/api/images`
2. `routers/images_router.go` 把路由注册到图片模块
3. `api/images_api/image_upload.go` 做：
   - 读取文件
   - 校验登录身份
   - 校验后缀
   - 校验大小
   - 检查 MIME
   - 检查图片头
   - 计算 MD5 去重
   - 生成保存目录和文件名
   - 写磁盘
   - 写数据库
4. 返回给前端图片路径

### 6.2 显示时

1. 前端拿到数据库里的路径
2. 前端通过 `$resolveImg` 处理
3. 浏览器访问 `/uploads/...`
4. `routers/enter.go` 的静态资源路由处理请求
5. 如果路径不属于受保护目录，就把文件返回给浏览器

## 7. 附件上传/下载链路

附件和图片的一个关键区别是：

- 图片通常允许公开显示
- 附件通常需要权限控制

所以附件下载不直接走静态目录，而是走 API。

### 7.1 上传

1. 前端上传到 `/api/files`
2. `api/file_api/file_upload.go`：
   - 检查扩展名白名单
   - 检查大小
   - 计算哈希
   - 用户内去重
   - 保存到 `attachments` 子目录
   - 记录数据库
3. 返回的是下载接口地址 `/api/files/:id/download`

### 7.2 下载

1. 前端访问 `/api/files/:id/download`
2. `api/file_api/file_download.go` 检查：
   - 是否登录
   - 是否有权限下载
   - 文件是否存在
   - 路径是否仍在上传目录内
3. 最后通过 `c.FileAttachment(...)` 返回

## 8. 鉴权链路

最常用的鉴权中间件是：

- `middleware.JwtAuth()`
- `middleware.JwtAdmin()`

它们都会做：

1. 提取 token
2. 解析 JWT
3. 检查 Redis 黑名单
4. 成功后把 claims 放进 Gin 上下文

区别是：

- `JwtAuth()` 只要求已登录
- `JwtAdmin()` 额外要求管理员角色

## 9. 社交实时链路

社交模块是当前后端里最复杂的一块之一。

相关文件主要在：

- `api/social_api/social_ws.go`
- `api/social_api/social_message.go`
- `api/social_api/social_message_call.go`
- `api/social_api/social_presence.go`
- `api/social_api/social_group.go`
- `api/social_api/social_follow.go`
- `api/social_api/social_block.go`
- `api/social_api/social_file.go`

你可以把它拆成三部分理解：

### 9.1 静态数据接口

比如：

- 好友列表
- 黑名单
- 群组列表
- 历史消息

这些一般是普通 HTTP 接口。

### 9.2 实时事件

通过 WebSocket 推送：

- 新消息
- 状态变化
- 群组变化
- 来电邀请

### 9.3 语音通话

后端本身不传音频流，只做“信令中转”：

1. 谁邀请谁
2. 谁接受/拒绝
3. offer / answer / candidate 怎么转发

真正音频数据走的是前端之间的 WebRTC。

## 10. 新手最容易卡住的 5 个地方

### 10.1 为什么 handler 里能直接拿到 `claims`

因为前面的 `JwtAuth` / `JwtAdmin` 已经把解析后的用户信息塞进 `gin.Context` 了。

### 10.2 为什么很多目录里都有 `enter.go`

`enter.go` 通常是“当前模块的聚合入口”，方便统一导出结构体或服务对象。

### 10.3 为什么有些逻辑在 API 层，有些在 Service 层

这个项目并不是所有功能都严格三层分离。
一般是：

- 简单功能：直接在 API 层完成
- 复杂功能：抽到 Service

### 10.4 为什么附件不直接暴露 `/uploads/...`

因为附件下载通常需要权限控制，不能像公开图片一样随便裸链访问。

### 10.5 为什么有 Redis 黑名单

JWT 是无状态 token，默认签发后只要没过期就一直可用。
如果用户退出登录，单靠 JWT 本身无法立刻失效，所以需要 Redis 记录“已退出 token”。

## 11. 新手最推荐的三条阅读线

### 登录线

`main.go -> routers/user_router.go -> middleware/jwt_auth.go -> api/user_api/* -> utils/jwts/*`

### 图片线

`routers/images_router.go -> api/images_api/image_upload.go -> models/banner_model.go -> routers/enter.go`

### 社交线

`routers/social_router.go -> api/social_api/* -> models/social_model.go -> 前端 social store`

## 12. 最后给新手的建议

读这个后端时，不要按“功能多不多”来选入口，而要按“链路是否完整”来读。

最推荐你先完整读通这三条链路：

1. 登录鉴权链路
2. 图片上传显示链路
3. 私信 / 好友实时链路

只要你把这三条链路走通，这个项目的大多数结构你就能自己顺着看下去了。
