# GVB Server Detailed API Documentation

更新日期：`2026-03-08`

这份文档重点补充旧 Markdown 和旧 Swagger 里不够清晰的部分，主要覆盖：

- 配置扩展与 ES 导入导出
- 社交系统
- 公告与板块
- 文章举报、审核与附件
- 签到与个人空间

完整路由清单见 `docs/API_DOCUMENTATION_LATEST.md`，完整结构化定义见 `docs/swagger.json` / `docs/swagger.yaml`。

## 文件清单

- Swagger：`docs/swagger.json`
- Swagger YAML：`docs/swagger.yaml`
- 最新路由总表：`docs/API_DOCUMENTATION_LATEST.md`
- 详细联调文档：`docs/API_DOCUMENTATION_DETAILED.md`
- Postman Collection：`docs/GVB_Server_Detailed.postman_collection.json`
- Postman Environment：`docs/GVB_Server.postman_environment.json`

## 基础约定

- 基础地址：`http://127.0.0.1:8080/api`
- 鉴权头：
  - 历史写法：`token: <token>`
  - 标准写法：`Authorization: Bearer <token>`
- 通用成功响应格式：

```json
{
  "code": 0,
  "data": {},
  "msg": "成功"
}
```

- 文件下载、ES 导出接口返回的是文件流，不是 JSON。
- 管理员接口统一建议使用 `adminToken`。
- 普通用户联调接口统一建议使用 `token`。

## 测试变量

| 变量 | 示例值 | 说明 |
| --- | --- | --- |
| `baseUrl` | `http://127.0.0.1:8080` | 本地服务地址 |
| `token` | `登录后返回` | 普通用户 token |
| `adminToken` | `管理员登录后返回` | 管理员 token |
| `registerEmail` | `tester@example.com` | 测试注册邮箱 |
| `registerCode` | `123456` | 注册验证码占位值 |
| `userName` | `tester_demo` | 普通用户账号 |
| `userPassword` | `Admin@123456` | 普通用户密码示例 |
| `adminUserName` | `admin` | 管理员账号示例，按实际数据替换 |
| `adminPassword` | `Admin@123456` | 管理员密码示例，按实际数据替换 |
| `targetUserId` | `2` | 社交目标用户 ID |
| `groupId` | `1` | 群组 ID |
| `groupNo` | `G000001` | 群组号 |
| `boardId` | `1` | 板块 ID |
| `articleId` | `article-demo-001` | 文章 ID / ES 文档 ID |
| `fileId` | `1` | 文章附件 ID |
| `socialFileId` | `1` | 社交文件 ID |
| `spaceUserId` | `2` | 空间主人用户 ID |
| `uploadFilePath` | `C:\Users\Public\Documents\demo.pdf` | 文章附件上传路径 |
| `socialUploadFilePath` | `C:\Users\Public\Documents\demo-chat.pdf` | 社交文件上传路径 |
| `importJsonPath` | `C:\Users\Public\Documents\article_index_export.json` | ES 导入文件路径 |

## 推荐调用顺序

1. 普通用户注册：`/api/user_register_email_code` -> `/api/user_create` -> `/api/email_login`
2. 管理员联调后台：先用管理员账号调用 `/api/email_login` 拿到 `adminToken`
3. 测试配置扩展：先跑预览接口，再执行同步或导入
4. 测试社交：先准备两个用户，再测关注、私聊、群聊、文件
5. 测试空间：先登录，再测签到、发动态、发留言

## 鉴权与初始化

### 1. 发送注册验证码

```bash
curl --request POST "{{baseUrl}}/api/user_register_email_code" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "email": "{{registerEmail}}"
}'
```

注意事项：

- 需要服务端已经配置邮件发送能力。
- 这个验证码只用于普通注册流程。

### 2. 创建普通用户

```bash
curl --request POST "{{baseUrl}}/api/user_create" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "nick_name": "测试用户",
  "user_name": "{{userName}}",
  "email": "{{registerEmail}}",
  "code": "{{registerCode}}",
  "password": "{{userPassword}}",
  "role": 2
}'
```

### 3. 登录并获取 token

```bash
curl --request POST "{{baseUrl}}/api/email_login" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "user_name": "{{userName}}",
  "password": "{{userPassword}}"
}'
```

示例响应：

```json
{
  "code": 0,
  "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.demo.user",
  "msg": "成功"
}
```

调用注意事项：

- `user_name` 支持用户名或邮箱。
- 管理员联调时同样走这个接口，只是账号密码换成管理员。

## 配置扩展与 ES 管理

### 1. 获取公开站点配置

```bash
curl "{{baseUrl}}/api/settings/public/site_info"
```

适用场景：

- 前台首页初始化
- 网站标题、公告栏、备案信息展示

### 2. 预览枫枫文章同步

```bash
curl "{{baseUrl}}/api/settings/site_info/sync_fengfeng_preview?limit=20" \
  --header "token: {{adminToken}}"
```

示例响应：

```json
{
  "code": 0,
  "data": {
    "source_total": 320,
    "latest_scanned": 20,
    "sync_limit": 20,
    "new_candidate": 6,
    "duplicate_count": 12,
    "invalid_count": 2,
    "candidates": [
      {
        "article_id": "ff_demo_001",
        "title": "Go 注释补全实践"
      }
    ]
  },
  "msg": "成功"
}
```

### 3. 执行枫枫文章同步

```bash
curl --request POST "{{baseUrl}}/api/settings/site_info/sync_fengfeng" \
  --header "token: {{adminToken}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "article_ids": ["ff_demo_001", "ff_demo_002"],
  "sync_all": false,
  "include_update": true,
  "limit": 20
}'
```

调用注意事项：

- 先预览，再同步。
- `sync_all=true` 适合批量抓取；联调阶段建议先指定 `article_ids`。
- `include_update=true` 会尝试更新系统内已存在的同源文章。

### 4. 查看 ES 索引列表

```bash
curl "{{baseUrl}}/api/settings/es/indices" \
  --header "token: {{adminToken}}"
```

### 5. 导入 ES 文章索引

```bash
curl --request POST "{{baseUrl}}/api/settings/es/import" \
  --header "token: {{adminToken}}" \
  --form "index=article_index" \
  --form "file=@{{importJsonPath}}"
```

调用注意事项：

- 当前页面导入只支持 `article_index`。
- 导入不是直接回写原始文档，而是按文章创建规则重新校验。

## 社交系统

### 1. 获取好友面板摘要

```bash
curl "{{baseUrl}}/api/social/summary" \
  --header "token: {{token}}"
```

### 2. 搜索用户和群组

```bash
curl "{{baseUrl}}/api/social/discovery?key=2" \
  --header "token: {{token}}"
```

适用场景：

- 加好友前先查用户
- 入群前先查群号

### 3. 发送单聊消息

```bash
curl --request POST "{{baseUrl}}/api/social/direct/messages" \
  --header "token: {{token}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "rev_user_id": 2,
  "content": "你好，这是一条来自 Postman 的测试私信。",
  "msg_type": "text",
  "file_id": 0
}'
```

调用注意事项：

- 发文件时先调用 `/api/social/files` 上传，再把 `msg_type` 改成 `file`。
- 若双方存在拉黑关系，会直接失败。

### 4. 创建群组

```bash
curl --request POST "{{baseUrl}}/api/social/groups" \
  --header "token: {{token}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "name": "产品讨论群",
  "notice": "仅讨论产品需求和 Bug",
  "member_ids": [2, 3]
}'
```

调用注意事项：

- 只能邀请好友入群。
- 群成员总数上限是 30。

### 5. 更新在线状态

```bash
curl --request PUT "{{baseUrl}}/api/social/presence" \
  --header "token: {{token}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "mode": "busy",
  "status_text": "正在处理接口文档和 Postman",
  "is_invisible": false
}'
```

### 6. 上传社交文件

```bash
curl --request POST "{{baseUrl}}/api/social/files" \
  --header "token: {{token}}" \
  --form "file=@{{socialUploadFilePath}}"
```

### 7. 社交 WebSocket

```javascript
const ws = new WebSocket('ws://127.0.0.1:8080/api/social/ws?token=' + token)
ws.onmessage = (event) => console.log(JSON.parse(event.data))
ws.onopen = () => {
  ws.send(JSON.stringify({
    action: 'update_presence',
    mode: 'busy',
    status_text: '正在处理接口文档',
    is_invisible: false
  }))
}
```

调用注意事项：

- token 可以放 query、`token` 头，或 `Authorization: Bearer`。
- 常见 `action`：`ping`、`update_presence`、`call_invite`、`call_accept`、`call_reject`、`call_end`。

## 板块与公告

### 1. 创建板块

```bash
curl --request POST "{{baseUrl}}/api/boards" \
  --header "token: {{adminToken}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "name": "Go",
  "slug": "go",
  "description": "Go 后端开发",
  "notice": "先看版规再发帖",
  "rules": "禁止广告和灌水",
  "sort": 1,
  "is_enabled": true,
  "moderator_user_ids": [1],
  "deputy_moderator_user_ids": [2]
}'
```

### 2. 创建公告

```bash
curl --request POST "{{baseUrl}}/api/announcements" \
  --header "token: {{adminToken}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "title": "系统维护通知",
  "content": "今晚 23:00 开始维护，预计 30 分钟。",
  "level": "warning",
  "jump_link": "/notice/maintenance",
  "board_id": 0,
  "sort": 1,
  "is_show": true,
  "starts_at": "2026-03-08 20:00:00",
  "ends_at": "2026-03-09 08:00:00"
}'
```

调用注意事项：

- `board_id=0` 表示全站公告。
- `jump_link` 仅支持 `http(s)` 或站内相对路径。

## 文章举报、审核与附件

### 1. 提交文章举报

```bash
curl --request POST "{{baseUrl}}/api/articles/reports" \
  --header "token: {{token}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "article_id": "{{articleId}}",
  "reason": "内容涉嫌广告或引流",
  "content": "文中多处包含导流二维码和外链。"
}'
```

### 2. 审核文章

```bash
curl --request PUT "{{baseUrl}}/api/articles/review" \
  --header "token: {{adminToken}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "id": "{{articleId}}",
  "review_status": 2,
  "review_reason": ""
}'
```

调用注意事项：

- `review_status=2` 表示通过。
- `review_status=3` 表示驳回。
- 管理员可全局审核，版主只能审核自己负责板块下的文章。

### 3. 上传文章附件

```bash
curl --request POST "{{baseUrl}}/api/files" \
  --header "token: {{token}}" \
  --form "file=@{{uploadFilePath}}"
```

### 4. 获取后台运营统计

```bash
curl "{{baseUrl}}/api/data_sum/admin" \
  --header "token: {{adminToken}}"
```

## 签到与个人空间

### 1. 执行签到

```bash
curl --request POST "{{baseUrl}}/api/user_check_in" \
  --header "token: {{token}}"
```

### 2. 获取空间资料

```bash
curl "{{baseUrl}}/api/users/{{spaceUserId}}/profile"
```

### 3. 发布空间动态

```bash
curl --request POST "{{baseUrl}}/api/space/posts" \
  --header "token: {{token}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "content": "今天把接口文档、Swagger 和 Postman 都补齐了。",
  "attachments": [
    {
      "file_id": 1,
      "name": "接口清单.pdf",
      "url": "/api/files/1/download",
      "size": 204800
    }
  ],
  "is_private": false
}'
```

### 4. 发表空间留言

```bash
curl --request POST "{{baseUrl}}/api/space/messages" \
  --header "token: {{token}}" \
  --header "Content-Type: application/json" \
  --data-raw '{
  "space_user_id": 2,
  "content": "留言测试：接口文档和 Postman 都已经补好了。",
  "is_private": false
}'
```

调用注意事项：

- 私密动态默认仅本人和管理员可见。
- 私密留言默认仅空间主人、留言人和管理员可见。

## 常见排错

- 返回“未携带token”：确认请求头至少带了 `token` 或 `Authorization: Bearer`。
- 返回“权限错误”：通常是普通用户拿 `token` 去调管理员接口，应改用 `adminToken`。
- 文件上传失败：先检查 `settings.yaml` 里的 `upload.max_size`，再确认扩展名是否被允许。
- 社交发文件失败：必须先调用 `/api/social/files` 上传，再把返回的 `id` 填到消息接口的 `file_id`。
- ES 导入失败：先调用 `/api/settings/es/indices` 确认索引名；页面导入只支持 `article_index`。

## 备注

- 这份文档重点写了新增和最近补齐的接口，完整路由清单请看 `docs/API_DOCUMENTATION_LATEST.md`。
- 若要直接联调，优先导入 `docs/GVB_Server_Detailed.postman_collection.json` 和 `docs/GVB_Server.postman_environment.json`。
