# GVB Server API 测试指南

> 测试账号: zhangsan / 1  
> 基础URL: `http://127.0.0.1:8080/api`

---

## 快速开始

### 1. 获取 Token

```bash
curl -X POST http://127.0.0.1:8080/api/email_login \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "zhangsan",
    "password": "1"
  }'
```

**预期响应**:
```json
{
  "code": 0,
  "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "msg": "登录成功"
}
```

将返回的 `data` 字段保存为 `TOKEN` 变量，后续请求使用。

---

## 接口测试用例

### 一、用户管理测试

#### 1. 获取用户信息
```bash
curl -X GET http://127.0.0.1:8080/api/user_info \
  -H "token: $TOKEN"
```

**预期响应**:
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "nick_name": "张三",
    "user_name": "zhangsan",
    "role": 1
  },
  "msg": "获取成功"
}
```

---

#### 2. 修改昵称
```bash
curl -X PUT http://127.0.0.1:8080/api/user_update_nick_name \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "nick_name": "张三丰"
  }'
```

---

#### 3. 修改密码
```bash
curl -X PUT http://127.0.0.1:8080/api/user_password \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "old_password": "1",
    "new_password": "newpass123"
  }'
```

---

#### 4. 用户注销
```bash
curl -X POST http://127.0.0.1:8080/api/logout \
  -H "token: $TOKEN"
```

---

### 二、文章管理测试

#### 1. 获取文章列表
```bash
curl -X GET "http://127.0.0.1:8080/api/articles?page=1&limit=5"
```

**预期响应**:
```json
{
  "code": 0,
  "data": {
    "count": 50,
    "list": [
      {
        "id": "article_001",
        "title": "测试文章标题",
        "abstract": "这是文章简介...",
        "look_count": 100,
        "digg_count": 20
      }
    ]
  },
  "msg": "获取成功"
}
```

---

#### 2. 获取文章详情
```bash
curl -X GET http://127.0.0.1:8080/api/articles/article_001
```

---

#### 3. 创建文章
```bash
curl -X POST http://127.0.0.1:8080/api/articles \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "title": "我的第一篇测试博客",
    "abstract": "这是测试文章的简介",
    "content": "# 欢迎使用GVB博客系统\n\n这是一篇测试文章。\n\n## 功能特点\n- 文章管理\n- 标签管理\n- 评论系统",
    "category": "技术分享",
    "source": "原创",
    "link": "",
    "banner_id": 1,
    "tags": ["Go", "后端", "测试"]
  }'
```

---

#### 4. 更新文章
```bash
curl -X PUT http://127.0.0.1:8080/api/articles \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id": "article_001",
    "title": "更新后的标题",
    "abstract": "更新后的简介",
    "content": "更新后的内容",
    "category": "技术分享",
    "source": "原创",
    "link": "",
    "banner_id": 1,
    "tags": ["Go", "后端"]
  }'
```

---

#### 5. 删除文章
```bash
curl -X DELETE http://127.0.0.1:8080/api/articles \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id_list": ["article_001", "article_002"]
  }'
```

---

#### 6. 搜索文章
```bash
curl -X GET "http://127.0.0.1:8080/api/articles/search?key=Go&page=1&limit=10"
```

---

#### 7. 获取文章标签
```bash
curl -X GET http://127.0.0.1:8080/api/articles/tags
```

---

#### 8. 获取文章分类
```bash
curl -X GET http://127.0.0.1:8080/api/articles/categorys
```

---

#### 9. 收藏文章
```bash
curl -X POST http://127.0.0.1:8080/api/articles/collects \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id": "article_001"
  }'
```

---

#### 10. 获取收藏列表
```bash
curl -X GET "http://127.0.0.1:8080/api/articles/collects?page=1&limit=10" \
  -H "token: $TOKEN"
```

---

#### 11. 取消收藏
```bash
curl -X DELETE http://127.0.0.1:8080/api/articles/collects/batch \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id_list": ["article_001"]
  }'
```

---

#### 12. 文章点赞
```bash
curl -X POST http://127.0.0.1:8080/api/digg \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id": "article_001"
  }'
```

---

### 三、标签管理测试

#### 1. 获取标签列表
```bash
curl -X GET "http://127.0.0.1:8080/api/tags?page=1&limit=10"
```

---

#### 2. 获取标签名称列表
```bash
curl -X GET http://127.0.0.1:8080/api/tags/names
```

---

#### 3. 创建标签
```bash
curl -X POST http://127.0.0.1:8080/api/tags \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "tag": "Vue3"
  }'
```

---

#### 4. 更新标签
```bash
curl -X PUT http://127.0.0.1:8080/api/tags/1 \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "tag": "Vue3+TypeScript"
  }'
```

---

#### 5. 删除标签
```bash
curl -X DELETE http://127.0.0.1:8080/api/tags \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id_list": [1, 2]
  }'
```

---

### 四、评论管理测试

#### 1. 获取评论列表
```bash
curl -X GET "http://127.0.0.1:8080/api/comments?article_id=article_001"
```

---

#### 2. 发表评论
```bash
curl -X POST http://127.0.0.1:8080/api/comments \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "article_id": "article_001",
    "content": "这篇文章写得真好！学到了很多。",
    "parent_comment_id": 0
  }'
```

---

#### 3. 回复评论
```bash
curl -X POST http://127.0.0.1:8080/api/comments \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "article_id": "article_001",
    "content": "感谢你的支持！",
    "parent_comment_id": 1
  }'
```

---

#### 4. 点赞评论
```bash
curl -X POST http://127.0.0.1:8080/api/comments/digg/1 \
  -H "token: $TOKEN"
```

---

#### 5. 删除评论
```bash
curl -X DELETE http://127.0.0.1:8080/api/comments/1 \
  -H "token: $TOKEN"
```

---

### 五、消息管理测试

#### 1. 获取消息列表
```bash
curl -X GET "http://127.0.0.1:8080/api/messages" \
  -H "token: $TOKEN"
```

---

#### 2. 发送消息
```bash
curl -X POST http://127.0.0.1:8080/api/messages \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "rev_user_id": 2,
    "content": "你好，这是测试私信"
  }'
```

---

#### 3. 获取消息记录
```bash
curl -X POST http://127.0.0.1:8080/api/messages/record \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "user_id": 2
  }'
```

---

### 六、广告管理测试

#### 1. 获取广告列表
```bash
curl -X GET "http://127.0.0.1:8080/api/adverts?page=1&limit=10"
```

---

#### 2. 创建广告
```bash
curl -X POST http://127.0.0.1:8080/api/adverts \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "title": "测试广告",
    "href": "https://example.com",
    "images": "https://example.com/ad.jpg",
    "is_show": true
  }'
```

---

#### 3. 更新广告
```bash
curl -X PUT http://127.0.0.1:8080/api/adverts/1 \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "title": "更新后的广告",
    "href": "https://example.com",
    "images": "https://example.com/ad.jpg",
    "is_show": true
  }'
```

---

#### 4. 删除广告
```bash
curl -X DELETE http://127.0.0.1:8080/api/adverts \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id_list": [1, 2]
  }'
```

---

### 七、菜单管理测试

#### 1. 获取菜单列表
```bash
curl -X GET http://127.0.0.1:8080/api/menus
```

---

#### 2. 创建菜单
```bash
curl -X POST http://127.0.0.1:8080/api/menus \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "title": "测试菜单",
    "path": "/test",
    "slogan": "测试标语",
    "abstract": ["简介1", "简介2"],
    "abstract_time": 3,
    "banner_time": 5,
    "sort": 1
  }'
```

---

#### 3. 更新菜单
```bash
curl -X PUT http://127.0.0.1:8080/api/menus/1 \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "title": "更新后的菜单",
    "path": "/test",
    "slogan": "更新后的标语",
    "abstract": ["简介1", "简介2", "简介3"],
    "abstract_time": 3,
    "banner_time": 5,
    "sort": 1
  }'
```

---

#### 4. 删除菜单
```bash
curl -X DELETE http://127.0.0.1:8080/api/menus \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id_list": [1]
  }'
```

---

### 八、图片管理测试

#### 1. 获取图片列表
```bash
curl -X GET "http://127.0.0.1:8080/api/images?page=1&limit=10" \
  -H "token: $TOKEN"
```

---

#### 2. 上传图片
```bash
curl -X POST http://127.0.0.1:8080/api/images \
  -H "token: $TOKEN" \
  -F "image=@/path/to/your/image.jpg"
```

---

#### 3. 更新图片名称
```bash
curl -X PUT http://127.0.0.1:8080/api/images \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id": 1,
    "name": "新图片名称.jpg"
  }'
```

---

#### 4. 删除图片
```bash
curl -X DELETE http://127.0.0.1:8080/api/images \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id_list": [1, 2]
  }'
```

---

### 九、数据统计测试

#### 1. 获取数据统计
```bash
curl -X GET http://127.0.0.1:8080/api/data_sum \
  -H "token: $TOKEN"
```

---

#### 2. 获取近七日登录数据
```bash
curl -X GET http://127.0.0.1:8080/api/data/seven_login \
  -H "token: $TOKEN"
```

---

### 十、日志管理测试

#### 1. 获取日志列表
```bash
curl -X GET "http://127.0.0.1:8080/api/logs?page=1&limit=10" \
  -H "token: $TOKEN"
```

---

#### 2. 删除日志
```bash
curl -X DELETE http://127.0.0.1:8080/api/logs \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "id_list": [1, 2, 3]
  }'
```

---

### 十一、其他接口测试

#### 1. 获取新闻列表
```bash
curl -X GET http://127.0.0.1:8080/api/news
```

---

#### 2. 获取聊天列表
```bash
curl -X GET "http://127.0.0.1:8080/api/chats?page=1&limit=10" \
  -H "token: $TOKEN"
```

---

#### 3. 获取系统配置
```bash
curl -X GET http://127.0.0.1:8080/api/settings/site
```

---

#### 4. 更新系统配置
```bash
curl -X PUT http://127.0.0.1:8080/api/settings/site \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -d '{
    "site_title": "我的博客",
    "site_logo": "https://example.com/logo.png"
  }'
```

---

## 测试脚本（Bash）

保存以下内容为 `test_api.sh`:

```bash
#!/bin/bash

BASE_URL="http://127.0.0.1:8080/api"
TOKEN=""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "=== GVB Server API 测试脚本 ==="
echo ""

# 1. 登录
echo "1. 测试登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/email_login" \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "zhangsan",
    "password": "1"
  }')

echo "响应: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"data":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}登录失败${NC}"
  exit 1
fi

echo -e "${GREEN}登录成功，Token: ${TOKEN:0:30}...${NC}"
echo ""

# 2. 获取用户信息
echo "2. 测试获取用户信息..."
curl -s -X GET "$BASE_URL/user_info" \
  -H "token: $TOKEN" | jq .
echo ""

# 3. 获取文章列表
echo "3. 测试获取文章列表..."
curl -s -X GET "$BASE_URL/articles?page=1&limit=5" | jq '.data.list[0:2]'
echo ""

# 4. 获取标签列表
echo "4. 测试获取标签列表..."
curl -s -X GET "$BASE_URL/tags?page=1&limit=5" | jq '.data.list[0:3]'
echo ""

# 5. 获取评论列表
echo "5. 测试获取评论列表..."
curl -s -X GET "$BASE_URL/comments?article_id=article_001" | jq '.data[0:3]'
echo ""

# 6. 获取广告列表
echo "6. 测试获取广告列表..."
curl -s -X GET "$BASE_URL/adverts?page=1&limit=5" | jq '.data.list[0:3]'
echo ""

# 7. 获取菜单列表
echo "7. 测试获取菜单列表..."
curl -s -X GET "$BASE_URL/menus" | jq '.data.list[0:3]'
echo ""

# 8. 获取数据统计
echo "8. 测试获取数据统计..."
curl -s -X GET "$BASE_URL/data_sum" \
  -H "token: $TOKEN" | jq .
echo ""

echo -e "${GREEN}=== 测试完成 ===${NC}"
```

**运行方式**:
```bash
chmod +x test_api.sh
./test_api.sh
```

---

## Postman 集合

### 环境变量
```json
{
  "base_url": "http://127.0.0.1:8080/api",
  "token": ""
}
```

### 预请求脚本（登录接口）
```javascript
// 登录成功后自动设置token
if (pm.response.code === 200) {
  const jsonData = pm.response.json();
  if (jsonData.code === 0) {
    pm.environment.set("token", jsonData.data);
  }
}
```

### 认证接口的 Header
```javascript
// 在需要认证的接口的 Pre-request Script 中添加
pm.request.headers.add({
  key: 'token',
  value: pm.environment.get("token")
});
```

---

## 常见问题

### 1. Token 失效
**现象**: 返回 `{"code": 1002, "msg": "未授权"}`  
**解决**: 重新调用登录接口获取新 token

### 2. 参数错误
**现象**: 返回 `{"code": 1001, "msg": "参数错误"}`  
**解决**: 检查请求参数是否符合接口文档要求

### 3. 权限不足
**现象**: 返回 `{"code": 1005, "msg": "权限不足"}`  
**解决**: 确认当前用户角色是否有权限调用该接口

---

*测试文档生成时间: 2026-03-01*
