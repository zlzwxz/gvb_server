# GVB Server API 接口文档

> 版本: 1.0  
> 基础URL: `http://127.0.0.1:8080/api`  
> 测试账号: zhangsan / 1

---

## 目录

1. [认证相关](#认证相关)
2. [用户管理](#用户管理)
3. [文章管理](#文章管理)
4. [标签管理](#标签管理)
5. [评论管理](#评论管理)
6. [消息管理](#消息管理)
7. [广告管理](#广告管理)
8. [菜单管理](#菜单管理)
9. [图片管理](#图片管理)
10. [数据统计](#数据统计)
11. [日志管理](#日志管理)
12. [其他接口](#其他接口)

---

## 认证相关

### 1. 邮箱登录

**接口**: `POST /api/email_login`

**描述**: 通过用户名/邮箱和密码进行登录

**请求参数**:
```json
{
  "user_name": "zhangsan",
  "password": "1"
}
```

**响应示例**:
```json
{
  "code": 0,
  "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "msg": "登录成功"
}
```

**说明**: 返回的 data 字段是 JWT token，后续请求需要在 Header 中携带 `token: <jwt_token>`

---

### 2. 用户注销

**接口**: `POST /api/logout`

**描述**: 用户退出登录

**请求头**:
```
token: <jwt_token>
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "注销成功"
}
```

---

### 3. QQ登录获取链接

**接口**: `GET /api/qq_login_path`

**描述**: 获取QQ登录跳转链接

**响应示例**:
```json
{
  "code": 0,
  "data": "https://graph.qq.com/oauth2.0/authorize?...",
  "msg": "获取成功"
}
```

---

### 4. QQ登录回调

**接口**: `POST /api/qq_login?code={code}`

**描述**: QQ登录回调接口

**请求参数**:
- Query: `code` - QQ授权码

**响应示例**:
```json
{
  "code": 0,
  "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "msg": "登录成功"
}
```

---

## 用户管理

### 1. 获取用户信息

**接口**: `GET /api/user_info`

**描述**: 根据token获取当前用户信息

**请求头**:
```
token: <jwt_token>
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "nick_name": "张三",
    "user_name": "zhangsan",
    "avatar": "https://example.com/avatar.jpg",
    "email": "zhangsan@example.com",
    "role": 1,
    "sign_status": 3
  },
  "msg": "获取成功"
}
```

---

### 2. 获取用户列表

**接口**: `GET /api/users?page=1&limit=10`

**描述**: 获取所有用户列表（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
- `page` (可选): 页码，默认1
- `limit` (可选): 每页数量，默认10

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 100,
    "list": [
      {
        "id": 1,
        "nick_name": "张三",
        "user_name": "zhangsan",
        "role": 1
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 3. 创建用户

**接口**: `POST /api/user_create`

**描述**: 创建新用户（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "user_name": "lisi",
  "password": "123456",
  "role": 2,
  "nick_name": "李四",
  "email": "lisi@example.com"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "创建成功"
}
```

---

### 4. 批量删除用户

**接口**: `DELETE /api/users`

**描述**: 批量删除用户（管理员权限）

**请求头**:
```
token: <jwt_token>
Content-Type: application/json
```

**请求参数**:
```json
{
  "id_list": [1, 2, 3]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

### 5. 修改密码

**接口**: `PUT /api/user_password`

**描述**: 修改当前用户密码

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "old_password": "1",
  "new_password": "newpassword123"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "密码修改成功"
}
```

---

### 6. 修改用户角色

**接口**: `PUT /api/user_role`

**描述**: 修改用户角色（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "user_id": 2,
  "role": 3
}
```

**角色说明**:
- 1: 管理员
- 2: 普通用户
- 3: 游客
- 4: 被禁用的用户

**响应示例**:
```json
{
  "code": 0,
  "msg": "角色修改成功"
}
```

---

### 7. 绑定邮箱

**接口**: `POST /api/user_bind_email`

**描述**: 用户绑定邮箱

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "绑定成功"
}
```

---

### 8. 修改昵称

**接口**: `PUT /api/user_update_nick_name`

**描述**: 修改用户昵称

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "nick_name": "新昵称"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "修改成功"
}
```

---

## 文章管理

### 1. 获取文章列表

**接口**: `GET /api/articles?page=1&limit=10&tag=Go`

**描述**: 获取文章列表，支持分页和标签筛选

**请求参数**:
- `page` (可选): 页码，默认1
- `limit` (可选): 每页数量，默认10
- `sort` (可选): 排序方式
- `tag` (可选): 标签筛选
- `is_user` (可选): 是否只显示当前用户的文章

**请求头** (当is_user为true时必填):
```
token: <jwt_token>
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 50,
    "list": [
      {
        "id": "article_001",
        "title": "Go语言入门教程",
        "abstract": "这是一篇关于Go语言的入门教程...",
        "content": "# Go语言入门...",
        "category": "后端开发",
        "tags": ["Go", "后端"],
        "banner_url": "https://example.com/banner.jpg",
        "user_id": 1,
        "user_nick_name": "张三",
        "look_count": 100,
        "digg_count": 20,
        "comment_count": 5,
        "collects_count": 10,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 2. 获取文章详情

**接口**: `GET /api/articles/{id}`

**描述**: 根据ID获取文章详情

**请求参数**:
- Path: `id` - 文章ID

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "id": "article_001",
    "title": "Go语言入门教程",
    "content": "# Go语言入门...",
    "tags": ["Go", "后端"],
    "look_count": 100,
    "digg_count": 20
  },
  "msg": "获取成功"
}
```

---

### 3. 根据标题获取文章

**接口**: `GET /api/articles/detail?title=Go语言入门教程`

**描述**: 根据文章标题获取详情

**请求参数**:
- Query: `title` - 文章标题

**响应示例**: 同上

---

### 4. 创建文章

**接口**: `POST /api/articles`

**描述**: 创建新文章

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "title": "我的第一篇博客",
  "abstract": "这是文章简介",
  "content": "# 文章内容\n这是正文内容...",
  "category": "生活随笔",
  "source": "原创",
  "link": "",
  "banner_id": 1,
  "tags": ["生活", "随笔"]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "创建成功"
}
```

---

### 5. 更新文章

**接口**: `PUT /api/articles`

**描述**: 更新文章

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id": "article_001",
  "title": "更新的标题",
  "abstract": "更新的简介",
  "content": "更新的内容",
  "category": "技术",
  "source": "原创",
  "link": "",
  "banner_id": 2,
  "tags": ["Go", "后端"]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "更新成功"
}
```

---

### 6. 批量删除文章

**接口**: `DELETE /api/articles`

**描述**: 批量删除文章

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id_list": ["article_001", "article_002"]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

### 7. 全文搜索

**接口**: `GET /api/articles/search?key=Go&page=1&limit=10`

**描述**: 全文搜索文章

**请求参数**:
- `key` (必填): 搜索关键词
- `page` (可选): 页码
- `limit` (可选): 每页数量

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 10,
    "list": [
      {
        "id": "article_001",
        "title": "Go语言入门教程",
        "abstract": "..."
      }
    ]
  },
  "msg": "搜索成功"
}
```

---

### 8. 获取文章标签列表

**接口**: `GET /api/articles/tags`

**描述**: 获取所有文章标签

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 20,
    "list": [
      {
        "tag": "Go",
        "count": 15,
        "article_id_list": ["article_001", "article_002"]
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 9. 获取文章分类列表

**接口**: `GET /api/articles/categorys`

**描述**: 获取文章分类列表

**响应示例**:
```json
{
  "code": 0,
  "data": [
    {
      "label": "后端开发",
      "value": "backend"
    },
    {
      "label": "前端开发",
      "value": "frontend"
    }
  ],
  "msg": "获取成功"
}
```

---

### 10. 获取文章日历数据

**接口**: `GET /api/articles/calendar`

**描述**: 获取过去一年每天的文章发布数量

**响应示例**:
```json
{
  "code": 0,
  "data": [
    {
      "date": "2024-01-01",
      "count": 5
    }
  ],
  "msg": "获取成功"
}
```

---

### 11. 收藏文章

**接口**: `POST /api/articles/collects`

**描述**: 收藏文章

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id": "article_001"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "收藏成功"
}
```

---

### 12. 获取收藏列表

**接口**: `GET /api/articles/collects?page=1&limit=10`

**描述**: 获取当前用户的文章收藏列表

**请求头**:
```
token: <jwt_token>
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 10,
    "list": [
      {
        "id": "article_001",
        "title": "Go语言入门",
        "abstract": "..."
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 13. 批量取消收藏

**接口**: `DELETE /api/articles/collects/batch`

**描述**: 批量取消收藏

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id_list": ["article_001", "article_002"]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "取消收藏成功"
}
```

---

### 14. 根据ID获取文章内容

**接口**: `GET /api/articles/content/{id}`

**描述**: 根据ID获取文章内容

**请求参数**:
- Path: `id` - 文章ID

**响应示例**:
```json
{
  "code": 0,
  "data": "# Go语言入门...",
  "msg": "获取成功"
}
```

---

## 标签管理

### 1. 获取标签列表

**接口**: `GET /api/tags?page=1&limit=10`

**描述**: 获取标签列表

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 50,
    "list": [
      {
        "id": 1,
        "tag": "Go",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 2. 获取标签名称列表

**接口**: `GET /api/tags/names`

**描述**: 获取所有标签名称

**响应示例**:
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "tag": "Go"
    },
    {
      "id": 2,
      "tag": "Python"
    }
  ],
  "msg": "获取成功"
}
```

---

### 3. 创建标签

**接口**: `POST /api/tags`

**描述**: 创建新标签（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "tag": "JavaScript"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "创建成功"
}
```

---

### 4. 更新标签

**接口**: `PUT /api/tags/{id}`

**描述**: 更新标签（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
- Path: `id` - 标签ID
- Body:
```json
{
  "tag": "TypeScript"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "更新成功"
}
```

---

### 5. 删除标签

**接口**: `DELETE /api/tags`

**描述**: 批量删除标签（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id_list": [1, 2, 3]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

## 评论管理

### 1. 获取评论列表

**接口**: `GET /api/comments?article_id=article_001`

**描述**: 获取文章评论列表

**请求参数**:
- `article_id` (必填): 文章ID

**响应示例**:
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "content": "写得真好！",
      "user_id": 2,
      "user_nick_name": "李四",
      "article_id": "article_001",
      "parent_comment_id": 0,
      "digg_count": 5,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "msg": "获取成功"
}
```

---

### 2. 创建评论

**接口**: `POST /api/comments`

**描述**: 发表评论

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "article_id": "article_001",
  "content": "这篇文章很有帮助！",
  "parent_comment_id": 0
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "评论成功"
}
```

---

### 3. 删除评论

**接口**: `DELETE /api/comments/{id}`

**描述**: 删除评论

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
- Path: `id` - 评论ID

**响应示例**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

### 4. 点赞评论

**接口**: `POST /api/comments/digg/{id}`

**描述**: 点赞/取消点赞评论

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
- Path: `id` - 评论ID

**响应示例**:
```json
{
  "code": 0,
  "msg": "点赞成功"
}
```

---

## 消息管理

### 1. 获取消息列表

**接口**: `GET /api/messages`

**描述**: 获取当前用户的消息列表

**请求头**:
```
token: <jwt_token>
```

**响应示例**:
```json
{
  "code": 0,
  "data": [
    {
      "send_user_id": 2,
      "send_user_nick_name": "李四",
      "send_user_avatar": "https://example.com/avatar.jpg",
      "rev_user_id": 1,
      "content": "你好！",
      "message_count": 5,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "msg": "获取成功"
}
```

---

### 2. 获取所有消息（管理员）

**接口**: `GET /api/messages/all?page=1&limit=10`

**描述**: 获取所有消息列表（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**响应示例**: 同上

---

### 3. 发送消息

**接口**: `POST /api/messages`

**描述**: 发送消息给其他用户

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "rev_user_id": 2,
  "content": "你好，这是测试消息"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "发送成功"
}
```

---

### 4. 获取消息记录

**接口**: `POST /api/messages/record`

**描述**: 获取与指定用户的消息记录

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "user_id": 2
}
```

**响应示例**:
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "content": "你好",
      "send_user_id": 1,
      "rev_user_id": 2,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "msg": "获取成功"
}
```

---

## 广告管理

### 1. 获取广告列表

**接口**: `GET /api/adverts?page=1&limit=10`

**描述**: 获取广告列表

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 10,
    "list": [
      {
        "id": 1,
        "title": "测试广告",
        "href": "https://example.com",
        "images": "https://example.com/ad.jpg",
        "is_show": true,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 2. 创建广告

**接口**: `POST /api/adverts`

**描述**: 创建广告（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "title": "新广告",
  "href": "https://example.com",
  "images": "https://example.com/ad.jpg",
  "is_show": true
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "创建成功"
}
```

---

### 3. 更新广告

**接口**: `PUT /api/adverts/{id}`

**描述**: 更新广告（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
- Path: `id` - 广告ID
- Body: 同创建广告

**响应示例**:
```json
{
  "code": 0,
  "msg": "更新成功"
}
```

---

### 4. 删除广告

**接口**: `DELETE /api/adverts`

**描述**: 批量删除广告（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id_list": [1, 2, 3]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

## 菜单管理

### 1. 获取菜单列表

**接口**: `GET /api/menus`

**描述**: 获取菜单列表

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 5,
    "list": [
      {
        "id": 1,
        "title": "首页",
        "path": "/",
        "slogan": "欢迎来到首页",
        "abstract": ["简介1", "简介2"],
        "abstract_time": 3,
        "banner_time": 5,
        "sort": 1,
        "banners": [
          {
            "id": 1,
            "path": "https://example.com/banner1.jpg"
          }
        ],
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 2. 获取菜单名称列表

**接口**: `GET /api/menus/names`

**描述**: 获取菜单名称列表

**响应示例**:
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "title": "首页",
      "path": "/"
    }
  ],
  "msg": "获取成功"
}
```

---

### 3. 获取菜单详情

**接口**: `GET /api/menus/{id}`

**描述**: 获取菜单详情

**请求参数**:
- Path: `id` - 菜单ID

**响应示例**: 同菜单列表项

---

### 4. 创建菜单

**接口**: `POST /api/menus`

**描述**: 创建菜单（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "title": "关于我们",
  "path": "/about",
  "slogan": "关于我们的介绍",
  "abstract": ["简介1", "简介2", "简介3"],
  "abstract_time": 3,
  "banner_time": 5,
  "sort": 2,
  "image_sort_list": [
    {
      "image_id": 1,
      "sort": 1
    }
  ]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "创建成功"
}
```

---

### 5. 更新菜单

**接口**: `PUT /api/menus/{id}`

**描述**: 更新菜单（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
- Path: `id` - 菜单ID
- Body: 同创建菜单

**响应示例**:
```json
{
  "code": 0,
  "msg": "更新成功"
}
```

---

### 6. 删除菜单

**接口**: `DELETE /api/menus`

**描述**: 批量删除菜单（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id_list": [1, 2, 3]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

## 图片管理

### 1. 获取图片列表

**接口**: `GET /api/images?page=1&limit=10`

**描述**: 获取图片列表

**请求头**:
```
token: <jwt_token>
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 50,
    "list": [
      {
        "id": 1,
        "path": "uploads/2024/01/01/image.jpg",
        "name": "image.jpg",
        "image_type": 1,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 2. 上传图片

**接口**: `POST /api/images`

**描述**: 上传图片

**请求头**:
```
token: <jwt_token>
Content-Type: multipart/form-data
```

**请求参数**:
- FormData: `image` - 图片文件

**响应示例**:
```json
{
  "code": 0,
  "data": "uploads/2024/01/01/image.jpg",
  "msg": "上传成功"
}
```

---

### 3. 更新图片名称

**接口**: `PUT /api/images`

**描述**: 更新图片名称

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id": 1,
  "name": "新名称.jpg"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "更新成功"
}
```

---

### 4. 删除图片

**接口**: `DELETE /api/images`

**描述**: 批量删除图片

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id_list": [1, 2, 3]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

## 数据统计

### 1. 获取数据统计

**接口**: `GET /api/data_sum`

**描述**: 获取系统数据统计

**请求头**:
```
token: <jwt_token>
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "user_count": 100,
    "article_count": 500,
    "message_count": 1000,
    "chat_group_count": 50,
    "now_login_count": 10,
    "now_sign_count": 5
  },
  "msg": "获取成功"
}
```

---

### 2. 获取近七日登录数据

**接口**: `GET /api/data/seven_login`

**描述**: 获取近七日登录和注册数据

**请求头**:
```
token: <jwt_token>
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "date_list": ["2024-01-01", "2024-01-02", "2024-01-03", "2024-01-04", "2024-01-05", "2024-01-06", "2024-01-07"],
    "login_data": [10, 15, 8, 20, 12, 18, 25],
    "sign_data": [2, 3, 1, 5, 2, 4, 6]
  },
  "msg": "获取成功"
}
```

---

## 日志管理

### 1. 获取日志列表

**接口**: `GET /api/logs?page=1&limit=10&type=1`

**描述**: 获取系统日志列表

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
- `page` (可选): 页码
- `limit` (可选): 每页数量
- `type` (可选): 日志类型
- `level` (可选): 日志级别

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 100,
    "list": [
      {
        "id": 1,
        "ip": "127.0.0.1",
        "addr": "本地",
        "level": 1,
        "content": "用户登录",
        "user_id": 1,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 2. 删除日志

**接口**: `DELETE /api/logs`

**描述**: 批量删除日志

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id_list": [1, 2, 3]
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

---

## 其他接口

### 1. 文章点赞

**接口**: `POST /api/digg`

**描述**: 点赞/取消点赞文章

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
```json
{
  "id": "article_001"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "点赞成功"
}
```

---

### 2. 获取聊天列表

**接口**: `GET /api/chats?page=1&limit=10`

**描述**: 获取群聊消息列表

**请求头**:
```
token: <jwt_token>
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "count": 100,
    "list": [
      {
        "id": 1,
        "content": "大家好",
        "user_id": 1,
        "user_nick_name": "张三",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "msg": "获取成功"
}
```

---

### 3. 获取新闻列表

**接口**: `GET /api/news`

**描述**: 获取新闻列表（从Redis获取）

**响应示例**:
```json
{
  "code": 0,
  "data": [
    {
      "title": "新闻标题",
      "url": "https://example.com/news",
      "hot": "100万",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "msg": "获取成功"
}
```

---

### 4. 获取系统配置

**接口**: `GET /api/settings/{name}`

**描述**: 获取系统配置信息

**请求参数**:
- Path: `name` - 配置名称，如 `site`, `email`, `qq` 等

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "site_title": "我的博客",
    "site_logo": "https://example.com/logo.png"
  },
  "msg": "获取成功"
}
```

---

### 5. 更新系统配置

**接口**: `PUT /api/settings/{name}`

**描述**: 更新系统配置（管理员权限）

**请求头**:
```
token: <jwt_token>
```

**请求参数**:
- Path: `name` - 配置名称
- Body: 配置内容（根据配置类型不同而不同）

**响应示例**:
```json
{
  "code": 0,
  "msg": "更新成功"
}
```

---

## 通用响应格式

### 成功响应
```json
{
  "code": 0,
  "data": {},
  "msg": "操作成功"
}
```

### 错误响应
```json
{
  "code": 1001,
  "data": null,
  "msg": "错误信息"
}
```

### 常见错误码
- `0`: 成功
- `1001`: 参数错误
- `1002`: 未授权
- `1003`: 资源不存在
- `1004`: 服务器内部错误
- `1005`: 权限不足

---

## 认证说明

### JWT Token 使用

1. 登录成功后获取 token
2. 在需要认证的接口请求头中添加: `token: <your_jwt_token>`
3. token 过期后需要重新登录获取

### 权限说明

- **普通用户**: 可以访问大部分接口
- **管理员**: 可以访问所有接口，包括用户管理、系统配置等
- **游客**: 只能访问公开接口

---

*文档生成时间: 2026-03-01*
