# ES 与好友管理第二层细化说明

这份文档不是“快速上手版”，而是给你在已经知道项目入口之后，继续往下深挖时用的。

如果说 `BACKEND_BEGINNER_GUIDE.md` 解决的是“我该从哪里开始看”，
那这份文档解决的是：

1. ES 数据到底谁是主数据，谁是派生数据？
2. 现在为什么要补导入 / 导出？
3. 后台好友管理页的数据从哪里来？
4. “跳到聊天框”这件事，后端到底负责了什么，前端又负责了什么？

---

## 1. 先建立两个最重要的认识

### 1.1 `article_index` 才是文章主数据

项目里的文章并没有落在 MySQL 的文章表里，而是直接存在 Elasticsearch 里。

对应模型在：

- `models/article_model.go`

它保存的是文章完整数据，例如：

- 标题
- 正文
- 摘要
- 标签
- 封面
- 作者信息
- 附件
- 审核状态
- 是否私密

所以你可以把 `article_index` 理解成：

`文章业务真正依赖的主索引`

---

### 1.2 `full_text_index` 不是主数据，而是“全文搜索派生索引”

对应模型在：

- `models/full_text_model.go`

它保存的不是完整文章，而是为了全文搜索拆分出来的片段，例如：

- 标题分段
- 正文片段
- slug
- 关联文章 ID

这些数据来自：

- `service/es_ser/full_text_search.go`
- `GetSearchIndexDataByContent`

所以你要记住：

- `article_index` 丢了，文章主数据就丢了
- `full_text_index` 丢了，还可以根据 `article_index` 重新生成

这也是这次补 ES 命令时，我把“导出主数据 + 重建全文索引”做成闭环，而不是只做全文索引导出的原因。

---

## 2. 这次补了哪些 ES 命令

命令入口在：

- `flag/enter.go`
- `flag/es.go`
- `flag/es_io.go`

新增后的 `-es` 支持四种值：

### 2.1 `-es create`

作用：

- 重建 `article_index`
- 重建 `full_text_index`

注意：

- 这是“清空重建结构”
- 不会自动把历史文章补回去

适合场景：

- 你刚配好 ES，第一次初始化索引
- 你明确知道当前 ES 里的数据可以丢

---

### 2.2 `-es export`

作用：

- 把当前 `article_index` 所有文档导出成一个 JSON 备份文件

默认文件路径：

`backup/es/article_index_backup.json`

适合场景：

- 重建索引前先做备份
- 想把当前文章索引迁移到另一台机器
- 想保留一个可以回滚的快照

---

### 2.3 `-es import`

作用：

1. 读取本地备份文件
2. 重建 `article_index`
3. 批量恢复文章文档
4. 重建 `full_text_index`
5. 根据恢复出来的文章重新生成全文搜索索引

你要把它理解成：

`从备份文件完整恢复文章索引，并顺便重新补好全文搜索索引`

---

### 2.4 `-es sync-fulltext`

作用：

- 不动 `article_index`
- 直接根据当前文章索引，重新生成 `full_text_index`

适合场景：

- 文章数据还在
- 只是全文搜索索引坏了、丢了、或者你怀疑它不一致

---

### 2.5 自定义备份文件路径

增加了一个参数：

`-es_file`

例如：

```powershell
go run main.go -es export -es_file backup/es/my_backup.json
```

Windows 下如果你已经编译出 exe，也可以这样：

```powershell
.\gvb-server.exe -es import -es_file backup/es/my_backup.json
```

### 2.6 现在后台系统设置页也能做 ES 管理

除了命令行，现在后台管理员还能直接在系统设置页操作：

1. 读取当前 ES 索引列表
2. 选择某个索引导出 JSON
3. 上传 JSON 执行页面导入

对应后端接口在：

- `GET /api/settings/es/indices`
- `GET /api/settings/es/export?index=...`
- `POST /api/settings/es/import`

这里要注意：

- 列表和导出针对“可见索引”开放
- 页面导入目前只支持 `article_index`
- 页面导入不是“原样灌回 ES”，而是逐条按文章创建规则重新校验并创建

所以页面导入更像：

`后台业务导入工具`

而命令行 `-es import` 更像：

`离线恢复工具`

---

## 3. ES 导出的调用链路怎么走

你可以按这个顺序读代码：

`main.go -> flag.Parse -> flag.SwitchOption -> flag.EsExport -> service/es_ser.ExportArticleBackup`

### 3.1 `main.go`

`main.go` 会先初始化：

- 配置
- 日志
- MySQL
- Redis
- Elasticsearch

然后再解析命令行参数。

如果发现这次是命令模式，例如：

- `-db`
- `-user`
- `-es`

那么程序会执行完命令直接退出，不会继续启动 Web 服务。

这一步的意义是：

`导出 / 导入 / 重建索引属于一次性维护任务，不应该顺便启动整个网站`

---

### 3.2 `flag.Parse`

这里把命令行参数解析成结构体：

- `DB`
- `User`
- `ES`
- `ESFile`

其中：

- `ES` 决定你到底想执行哪种 ES 维护命令
- `ESFile` 决定备份文件读写路径

---

### 3.3 `flag.SwitchOption`

这里是命令分发中心。

如果你传的是：

- `-es export`

它就会走：

- `EsExport(option.ESFile)`

---

### 3.4 `service/es_ser.ExportArticleBackup`

这是这次新增的核心函数之一，做了 4 件事：

1. 检查 `global.ESClient` 是否已初始化
2. 从 `article_index` 滚动读取所有文档
3. 把文档 ID 和原始 `source` 组织成 JSON
4. 写入本地备份文件

这里为什么保存的是：

- `文档 ID`
- `原始 source`

而不是只保存一个反序列化后的 Go 结构体？

因为这样更接近 ES 里的真实存储内容，恢复时更稳，不容易因为 Go 结构体的零值覆盖掉原本不存在的字段。

---

## 4. ES 导入的调用链路怎么走

导入链路：

`main.go -> flag.Parse -> flag.SwitchOption -> flag.EsImport -> service/es_ser.ImportArticleBackup`

### 4.1 第一步：读本地备份文件

函数：

- `readArticleBackup`

它负责：

- 读取 JSON 文件
- 反序列化为 `ArticleBackupPayload`
- 校验版本号和文档数组

---

### 4.2 第二步：重建文章索引

函数：

- `models.ArticleModel{}.CreateIndex()`

这里会先删旧索引，再建新索引。

所以你要清楚一件事：

`-es import` 是恢复动作，不是“往现有索引里追加几条数据”。

它的意图是：

`让当前 article_index 回到备份文件对应的状态`

---

### 4.3 第三步：批量恢复文章文档

函数：

- `bulkRestoreArticleDocuments`

这里使用的是 ES 的 Bulk API。

这么做有两个目的：

1. 一次导入很多文章时效率更高
2. 不需要每条文档都单独发一次请求

代码里还做了批次控制：

- 每批 `200` 条

这样不会一次塞太多请求，避免单批负载过大。

---

### 4.4 第四步：重建全文索引

函数：

- `rebuildFullTextFromBackup`

这一步不是简单把备份里的全文索引原样导入，而是：

1. 重新建 `full_text_index`
2. 逐条解析文章文档
3. 只保留“允许公开搜索”的文章
4. 用 `GetSearchIndexDataByContent` 重新切片生成全文索引文档

这样做的好处是：

- 全文索引一定和当前恢复后的文章数据一致
- 不依赖旧的全文索引快照
- 即使全文索引以前坏过，也能重建成干净状态

---

## 5. 为什么不是所有文章都能进全文索引

过滤逻辑在新增函数：

- `allowFullTextRebuild`

只有满足下面条件才会进入全文索引：

1. 有文章 ID
2. 有标题
3. 有正文
4. 不是私密文章
5. 审核状态是：
   - `ArticleReviewApproved`
   - `ArticleReviewLegacy`

这和现有业务逻辑保持一致。

你可以对照这些位置一起看：

- `api/article_api/article_create.go`
- `api/article_api/article_update.go`
- `api/article_api/article_review.go`
- `api/article_api/article_permission_helper.go`

它们都在表达同一个业务规则：

`只有公开可见的文章，才允许进入公开搜索`

---

## 6. 什么时候该用哪条 ES 命令

### 场景 A：第一次初始化

命令：

```powershell
go run main.go -es create
```

结果：

- 只建空索引

---

### 场景 B：准备升级、迁移、重建前先备份

命令：

```powershell
go run main.go -es export
```

结果：

- 把当前文章主数据保存到 `backup/es/article_index_backup.json`

---

### 场景 C：你刚把索引清空了，现在要恢复数据

命令：

```powershell
go run main.go -es import
```

结果：

1. 恢复 `article_index`
2. 自动重建 `full_text_index`

---

### 场景 D：文章主索引还在，但全文搜索结果不对

命令：

```powershell
go run main.go -es sync-fulltext
```

结果：

- 只重建全文索引

---

## 7. 后台好友管理接口的后端链路

你这次提到“好友管理里面功能不全，比如跳转到聊天框”，
这里要先区分：

1. 后端负责“把管理页数据查出来”
2. 前端负责“拿到数据后跳到哪个页面”

后端好友管理路由在：

- `routers/social_router.go`

后台管理相关接口有三个：

1. `GET /social/manage/summary`
2. `GET /social/manage/follows`
3. `GET /social/manage/blocks`
4. `GET /social/manage/groups`

它们对应实现都在：

- `api/social_api/social_admin.go`

这里再补一个容易忽略的点：

后台页现在虽然支持“切换筛选自动刷新、输入关键字自动刷新”，
但后端并没有因此新增一套专门的刷新接口。

原因很简单：

- 这几个接口本来就是标准列表查询接口
- 前端只是把原来“用户手动点刷新”的时机
- 改成了“tab / mode / key 变化时自动重新请求”

所以如果你后面还想继续给后台社交管理页加筛选，
优先复用现有参数：

- `page`
- `limit`
- `key`
- `mode`

而不是先去加一个新的“筛选刷新接口”。

---

### 7.1 `AdminSummaryView`

作用：

- 返回关注关系数
- 返回双向好友数
- 返回黑名单关系数
- 返回群组数
- 返回群成员总量
- 返回在线人数

它的特点是：

`这是概览接口，不返回具体列表，只负责页面顶部统计卡片`

---

### 7.2 `AdminFollowListView`

作用：

- 返回好友 / 单向关注 / 全部关注关系

关键点：

- `mode=friend` 只看互相关注
- `mode=follow` 只看单向关注
- `mode=all` 看全部关注关系

它通过 SQL 自连接判断是不是双向好友：

- `uf`
- `uf2`

逻辑上就是：

`A 关注 B，并且 B 也关注 A -> 互为好友`

---

### 7.3 `AdminBlockListView`

作用：

- 返回黑名单关系

它查的是：

- 谁拉黑了谁
- 原因是什么
- 创建时间是什么时候

---

### 7.4 `AdminGroupListView`

作用：

- 返回好友群组

会带上：

- 群号
- 群名
- 群头像
- 群主
- 成员数
- 群公告
- 最近更新时间

---

## 8. “跳到聊天框”为什么不是后端接口完成的

这是很多新手第一次做前后端联动时容易混淆的地方。

“跳到聊天框”本质上不是数据接口动作，而是：

`前端路由跳转动作`

也就是说，后端不需要额外提供一个“帮我跳页面”的接口。

真正分工是：

### 后端负责

1. 提供管理列表数据
2. 提供私信 / 群聊消息接口
3. 提供会话列表接口

### 前端负责

1. 从管理表格取到用户 ID 或群组 ID
2. 调用路由跳到 `/messages`
3. 把目标 ID 放到 query 里
4. 私信页根据 query 自动选中会话并拉消息

所以你以后看到“跳转聊天框”这类需求时，第一反应要是：

`先找前端 router 和目标页面的 query/watch 逻辑`

而不是先去找后端接口。

---

## 9. 阅读这块功能时，推荐按这条顺序看

### ES 导入导出线

`main.go -> flag/enter.go -> flag/es_io.go -> service/es_ser/backup.go -> models/article_model.go -> models/full_text_model.go`

### 后台好友管理后端线

`routers/social_router.go -> api/social_api/social_admin.go`

### 私信全文链路

`api/social_api/social_message.go -> api/social_api/social_group.go -> api/social_api/social_ws.go`

---

## 10. 你最容易踩的坑

### 10.1 误以为 `-es create` 会恢复历史数据

不会。

它只会重建空索引结构。

如果你要恢复数据，应该先：

1. `-es export`
2. 需要时再 `-es import`

---

### 10.2 误以为 `full_text_index` 是主数据

不是。

它只是为了全文搜索加速的派生索引。

---

### 10.3 误以为后台好友管理页可以“管理员代替别人聊天”

不是。

管理页里的“去聊天”只是把当前登录者带到聊天页，
然后由当前登录者和目标用户 / 群组建立会话。

它不会模拟表格里的两位用户互相发送消息。

---

### 10.4 误以为后端要写一个“跳页面接口”

不需要。

页面跳转是前端路由层的职责。

---

## 11. 最后给你的一个记忆口诀

你可以把这块功能记成一句话：

`文章主数据在 article_index，全文搜索在 full_text_index；后台管理负责查数据，跳聊天由前端路由负责。`
