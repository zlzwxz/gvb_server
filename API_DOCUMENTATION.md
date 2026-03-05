# 后端接口文档

## 1. 认证相关接口

### 1.1 邮箱登录
- **接口地址**：`POST /api/email_login`
- **接口描述**：通过用户名/邮箱和密码进行登录，返回token
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | user_name | string | 是 | 用户名或邮箱 |
  | password | string | 是 | 密码 |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": "token"}`
  - 失败：`{"code": 400, "message": "用户名或密码错误", "data": null}`
- **示例请求**：
  ```json
  {
    "user_name": "zhangsan",
    "password": "1"
  }
  ```
- **示例响应**：
  ```json
  {
    "code": 200,
    "message": "success",
    "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```

### 1.2 登出
- **接口地址**：`POST /api/logout`
- **接口描述**：用户登出
- **请求参数**：无
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 2. 用户管理接口

### 2.1 获取用户列表
- **接口地址**：`GET /api/users`
- **接口描述**：获取用户列表
- **请求参数**：无
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": [用户列表]}`

### 2.2 更新用户角色
- **接口地址**：`PUT /api/user_role`
- **接口描述**：更新用户角色
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | user_id | int | 是 | 用户ID |
  | role | int | 是 | 角色ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 2.3 更新用户密码
- **接口地址**：`PUT /api/user_password`
- **接口描述**：更新用户密码（需要管理员权限）
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | user_id | int | 是 | 用户ID |
  | password | string | 是 | 新密码 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 2.4 删除用户
- **接口地址**：`DELETE /api/users`
- **接口描述**：删除用户（需要管理员权限）
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | user_ids | []int | 是 | 用户ID列表 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 2.5 绑定邮箱
- **接口地址**：`POST /api/user_bind_email`
- **接口描述**：绑定邮箱
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | email | string | 是 | 邮箱地址 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 2.6 创建用户
- **接口地址**：`POST /api/user_create`
- **接口描述**：创建用户
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | user_name | string | 是 | 用户名 |
  | password | string | 是 | 密码 |
  | email | string | 是 | 邮箱 |
  | nick_name | string | 否 | 昵称 |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 2.7 获取用户信息
- **接口地址**：`GET /api/user_info`
- **接口描述**：获取当前用户信息
- **请求参数**：无
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {用户信息}}`

### 2.8 更新用户昵称
- **接口地址**：`PUT /api/user_update_nick_name`
- **接口描述**：更新用户昵称、签名、链接
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | nick_name | string | 否 | 昵称 |
  | bio | string | 否 | 签名 |
  | link | string | 否 | 链接 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 2.9 获取QQ登录链接
- **接口地址**：`GET /api/qq_login_path`
- **接口描述**：获取QQ登录的跳转链接
- **请求参数**：无
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": "QQ登录链接"}`

## 3. 文章管理接口

### 3.1 获取文章列表
- **接口地址**：`GET /api/articles`
- **接口描述**：获取文章列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
  | category_id | int | 否 | 分类ID |
  | tag_id | int | 否 | 标签ID |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [文章列表], "total": 总数}}`

### 3.2 获取文章详情
- **接口地址**：`GET /api/article/{id}`
- **接口描述**：获取文章详情
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 文章ID |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {文章详情}}`

### 3.3 创建文章
- **接口地址**：`POST /api/article`
- **接口描述**：创建文章
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | title | string | 是 | 标题 |
  | content | string | 是 | 内容 |
  | category_id | int | 是 | 分类ID |
  | tag_ids | []int | 是 | 标签ID列表 |
  | cover | string | 否 | 封面图 |
  | summary | string | 否 | 摘要 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 3.4 更新文章
- **接口地址**：`PUT /api/article`
- **接口描述**：更新文章
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 文章ID |
  | title | string | 是 | 标题 |
  | content | string | 是 | 内容 |
  | category_id | int | 是 | 分类ID |
  | tag_ids | []int | 是 | 标签ID列表 |
  | cover | string | 否 | 封面图 |
  | summary | string | 否 | 摘要 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 3.5 删除文章
- **接口地址**：`DELETE /api/article`
- **接口描述**：删除文章
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 文章ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 3.6 获取文章分类列表
- **接口地址**：`GET /api/article_category_list`
- **接口描述**：获取文章分类列表
- **请求参数**：无
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": [分类列表]}`

### 3.7 获取文章标签列表
- **接口地址**：`GET /api/article_tag_list`
- **接口描述**：获取文章标签列表
- **请求参数**：无
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": [标签列表]}`

### 3.8 文章收藏
- **接口地址**：`POST /api/article_collect`
- **接口描述**：收藏文章
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | article_id | int | 是 | 文章ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 3.9 获取收藏列表
- **接口地址**：`GET /api/article_collect_list`
- **接口描述**：获取收藏列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [收藏列表], "total": 总数}}`

### 3.10 批量取消收藏
- **接口地址**：`DELETE /api/article_collect_remove_batch`
- **接口描述**：批量取消收藏
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | ids | []int | 是 | 收藏ID列表 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 3.11 文章全文搜索
- **接口地址**：`GET /api/full_text_search`
- **接口描述**：文章全文搜索
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | keyword | string | 是 | 搜索关键词 |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [搜索结果], "total": 总数}}`

## 4. 评论管理接口

### 4.1 获取评论列表
- **接口地址**：`GET /api/comments`
- **接口描述**：获取评论列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | article_id | int | 是 | 文章ID |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [评论列表], "total": 总数}}`

### 4.2 创建评论
- **接口地址**：`POST /api/comment`
- **接口描述**：创建评论
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | article_id | int | 是 | 文章ID |
  | content | string | 是 | 评论内容 |
  | parent_id | int | 否 | 父评论ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 4.3 删除评论
- **接口地址**：`DELETE /api/comment`
- **接口描述**：删除评论
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 评论ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 4.4 评论点赞
- **接口地址**：`POST /api/comment_digg`
- **接口描述**：评论点赞
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | comment_id | int | 是 | 评论ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 5. 标签管理接口

### 5.1 获取标签列表
- **接口地址**：`GET /api/tags`
- **接口描述**：获取标签列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [标签列表], "total": 总数}}`

### 5.2 获取标签名称列表
- **接口地址**：`GET /api/tag_name_list`
- **接口描述**：获取标签名称列表
- **请求参数**：无
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": [标签名称列表]}`

### 5.3 创建标签
- **接口地址**：`POST /api/tag`
- **接口描述**：创建标签
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | name | string | 是 | 标签名称 |
  | color | string | 否 | 标签颜色 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 5.4 更新标签
- **接口地址**：`PUT /api/tag`
- **接口描述**：更新标签
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 标签ID |
  | name | string | 是 | 标签名称 |
  | color | string | 否 | 标签颜色 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 5.5 删除标签
- **接口地址**：`DELETE /api/tag`
- **接口描述**：删除标签
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 标签ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 6. 图片管理接口

### 6.1 上传图片
- **接口地址**：`POST /api/image_upload`
- **接口描述**：上传图片
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | file | file | 是 | 图片文件 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": "图片URL"}`

### 6.2 获取图片列表
- **接口地址**：`GET /api/images`
- **接口描述**：获取图片列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [图片列表], "total": 总数}}`

### 6.3 更新图片
- **接口地址**：`PUT /api/image`
- **接口描述**：更新图片信息
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 图片ID |
  | name | string | 是 | 图片名称 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 6.4 删除图片
- **接口地址**：`DELETE /api/image`
- **接口描述**：删除图片
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 图片ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 7. 广告管理接口

### 7.1 获取广告列表
- **接口地址**：`GET /api/adverts`
- **接口描述**：获取广告列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [广告列表], "total": 总数}}`

### 7.2 创建广告
- **接口地址**：`POST /api/advert`
- **接口描述**：创建广告
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | title | string | 是 | 广告标题 |
  | cover | string | 是 | 广告图片 |
  | link | string | 是 | 广告链接 |
  | sort | int | 否 | 排序 |
  | is_show | bool | 否 | 是否显示 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 7.3 更新广告
- **接口地址**：`PUT /api/advert`
- **接口描述**：更新广告
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 广告ID |
  | title | string | 是 | 广告标题 |
  | cover | string | 是 | 广告图片 |
  | link | string | 是 | 广告链接 |
  | sort | int | 否 | 排序 |
  | is_show | bool | 否 | 是否显示 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 7.4 删除广告
- **接口地址**：`DELETE /api/advert`
- **接口描述**：删除广告
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 广告ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 8. 菜单管理接口

### 8.1 获取菜单列表
- **接口地址**：`GET /api/menus`
- **接口描述**：获取菜单列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [菜单列表], "total": 总数}}`

### 8.2 获取菜单名称列表
- **接口地址**：`GET /api/menu_name_list`
- **接口描述**：获取菜单名称列表
- **请求参数**：无
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": [菜单名称列表]}`

### 8.3 创建菜单
- **接口地址**：`POST /api/menu`
- **接口描述**：创建菜单
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | title | string | 是 | 菜单标题 |
  | path | string | 是 | 菜单路径 |
  | icon | string | 否 | 菜单图标 |
  | sort | int | 否 | 排序 |
  | parent_id | int | 否 | 父菜单ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 8.4 更新菜单
- **接口地址**：`PUT /api/menu`
- **接口描述**：更新菜单
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 菜单ID |
  | title | string | 是 | 菜单标题 |
  | path | string | 是 | 菜单路径 |
  | icon | string | 否 | 菜单图标 |
  | sort | int | 否 | 排序 |
  | parent_id | int | 否 | 父菜单ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 8.5 删除菜单
- **接口地址**：`DELETE /api/menu`
- **接口描述**：删除菜单
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 菜单ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 9. 消息管理接口

### 9.1 获取消息列表
- **接口地址**：`GET /api/messages`
- **接口描述**：获取消息列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [消息列表], "total": 总数}}`

### 9.2 获取所有消息
- **接口地址**：`GET /api/messages_all`
- **接口描述**：获取所有消息（管理员）
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [消息列表], "total": 总数}}`

### 9.3 创建消息
- **接口地址**：`POST /api/message`
- **接口描述**：创建消息
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | content | string | 是 | 消息内容 |
  | user_id | int | 否 | 接收用户ID（管理员发送消息时使用） |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

### 9.4 获取消息记录
- **接口地址**：`GET /api/message_record`
- **接口描述**：获取消息记录
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [消息记录], "total": 总数}}`

## 10. 点赞管理接口

### 10.1 文章点赞
- **接口地址**：`POST /api/digg_article`
- **接口描述**：文章点赞
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | article_id | int | 是 | 文章ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 11. 聊天管理接口

### 11.1 获取聊天列表
- **接口地址**：`GET /api/chat_list`
- **接口描述**：获取聊天列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [聊天列表], "total": 总数}}`

### 11.2 创建聊天群组
- **接口地址**：`POST /api/chat_group`
- **接口描述**：创建聊天群组
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | name | string | 是 | 群组名称 |
  | avatar | string | 否 | 群组头像 |
  | user_ids | []int | 是 | 成员ID列表 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 12. 日志管理接口

### 12.1 获取日志列表
- **接口地址**：`GET /api/logs`
- **接口描述**：获取日志列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [日志列表], "total": 总数}}`

### 12.2 删除日志
- **接口地址**：`DELETE /api/log`
- **接口描述**：删除日志
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | id | int | 是 | 日志ID |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 13. 数据统计接口

### 13.1 获取数据汇总
- **接口地址**：`GET /api/data_sum`
- **接口描述**：获取数据汇总
- **请求参数**：无
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {数据汇总}}`

### 13.2 获取7天登录数据
- **接口地址**：`GET /api/data_seven_login`
- **接口描述**：获取7天登录数据
- **请求参数**：无
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": [7天登录数据]}`

## 14. 系统设置接口

### 14.1 获取系统设置
- **接口地址**：`GET /api/settings`
- **接口描述**：获取系统设置
- **请求参数**：无
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {系统设置}}`

### 14.2 更新系统设置
- **接口地址**：`PUT /api/setting`
- **接口描述**：更新系统设置
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | site_name | string | 否 | 网站名称 |
  | site_desc | string | 否 | 网站描述 |
  | site_keywords | string | 否 | 网站关键词 |
  | site_icp | string | 否 | 网站ICP |
  | site_copyright | string | 否 | 网站版权 |
  | site_analytics | string | 否 | 网站统计代码 |
- **请求头**：`Authorization: Bearer {token}`
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": null}`

## 15. 新闻管理接口

### 15.1 获取新闻列表
- **接口地址**：`GET /api/news`
- **接口描述**：获取新闻列表
- **请求参数**：
  | 参数名 | 类型 | 必选 | 描述 |
  | --- | --- | --- | --- |
  | page | int | 否 | 页码 |
  | size | int | 否 | 每页数量 |
- **响应格式**：
  - 成功：`{"code": 200, "message": "success", "data": {"list": [新闻列表], "total": 总数}}`

## 16. 认证方式

### 16.1 JWT认证
- **认证方式**：Bearer Token
- **请求头**：`Authorization: Bearer {token}`
- **token获取**：通过登录接口获取
- **token有效期**：根据系统配置

## 17. 测试账号

### 17.1 邮箱登录测试账号
- **用户名**：zhangsan
- **密码**：1
- **登录接口**：`POST /api/email_login`

## 18. 响应格式

所有接口的响应格式统一为：

```json
{
  "code": 200, // 状态码
  "message": "success", // 消息
  "data": null // 数据
}
```

### 18.1 状态码说明
- **200**：成功
- **400**：请求错误
- **401**：未授权
- **403**：禁止访问
- **404**：资源不存在
- **500**：服务器错误

## 19. 注意事项

1. 所有需要认证的接口都需要在请求头中添加`Authorization: Bearer {token}`
2. 测试环境数据库为测试数据库，可以进行数据的增删改查操作
3. 部分接口需要管理员权限，需要使用管理员账号登录
4. 上传文件接口需要使用`multipart/form-data`格式提交
5. 接口参数中的必填项必须填写，否则会返回400错误
6. 接口返回的数据结构可能会根据实际情况有所调整，请以实际返回为准
