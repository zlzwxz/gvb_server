# 前端代码生成指南

> 基于 GVB Server API 的前端开发指南  
> 支持生成 TypeScript、JavaScript、Axios、Fetch 等代码

---

## 目录

1. [Swagger/OpenAPI 代码生成](#swaggeropenapi-代码生成)
2. [TypeScript 类型定义](#typescript-类型定义)
3. [API 请求封装](#api-请求封装)
4. [React/Vue 集成示例](#reactvue-集成示例)
5. [自动生成工具推荐](#自动生成工具推荐)

---

## Swagger/OpenAPI 代码生成

### 方法一：使用 Swagger Codegen

#### 1. 安装 Swagger Codegen
```bash
# macOS
brew install swagger-codegen

# 或使用 Docker
docker pull swaggerapi/swagger-codegen-cli
```

#### 2. 生成 TypeScript-Angular 客户端
```bash
swagger-codegen generate \
  -i docs/swagger.json \
  -l typescript-angular \
  -o ./frontend/src/api
```

#### 3. 生成 TypeScript-Fetch 客户端
```bash
swagger-codegen generate \
  -i docs/swagger.json \
  -l typescript-fetch \
  -o ./frontend/src/api
```

#### 4. 生成 JavaScript 客户端
```bash
swagger-codegen generate \
  -i docs/swagger.json \
  -l javascript \
  -o ./frontend/src/api
```

---

### 方法二：使用 OpenAPI Generator（推荐）

#### 1. 安装 OpenAPI Generator
```bash
# npm
npm install @openapitools/openapi-generator-cli -g

# 或使用 npx
npx @openapitools/openapi-generator-cli version
```

#### 2. 生成 TypeScript-Axios 客户端
```bash
openapi-generator-cli generate \
  -i docs/swagger.json \
  -g typescript-axios \
  -o ./frontend/src/api \
  --additional-properties=supportsES6=true,npmName=gvb-api-client
```

#### 3. 生成 TypeScript-Fetch 客户端
```bash
openapi-generator-cli generate \
  -i docs/swagger.json \
  -g typescript-fetch \
  -o ./frontend/src/api \
  --additional-properties=supportsES6=true
```

#### 4. 生成 JavaScript 客户端
```bash
openapi-generator-cli generate \
  -i docs/swagger.json \
  -g javascript \
  -o ./frontend/src/api \
  --additional-properties=useES6=true,moduleName=gvbApi
```

---

## TypeScript 类型定义

### 手动定义核心类型

```typescript
// types/api.ts

// 通用响应格式
export interface ApiResponse<T = any> {
  code: number;
  data: T;
  msg: string;
}

// 分页请求参数
export interface PageParams {
  page?: number;
  limit?: number;
  sort?: string;
}

// 分页响应数据
export interface PageData<T> {
  count: number;
  list: T[];
}

// 用户类型
export interface User {
  id: number;
  created_at: string;
  nick_name: string;
  user_name: string;
  avatar?: string;
  email?: string;
  role: Role;
  sign_status: SignStatus;
}

export enum Role {
  Admin = 1,
  User = 2,
  Visitor = 3,
  Disabled = 4
}

export enum SignStatus {
  QQ = 1,
  Gitee = 2,
  Email = 3
}

// 文章类型
export interface Article {
  id: string;
  title: string;
  abstract: string;
  content: string;
  category: string;
  tags: string[];
  banner_id?: number;
  banner_url?: string;
  user_id: number;
  user_nick_name: string;
  user_avatar?: string;
  look_count: number;
  digg_count: number;
  comment_count: number;
  collects_count: number;
  created_at: string;
  updated_at?: string;
}

// 标签类型
export interface Tag {
  id: number;
  tag: string;
  created_at: string;
}

// 评论类型
export interface Comment {
  id: number;
  content: string;
  user_id: number;
  user_nick_name: string;
  user_avatar?: string;
  article_id: string;
  parent_comment_id: number;
  digg_count: number;
  created_at: string;
}

// 广告类型
export interface Advert {
  id: number;
  title: string;
  href: string;
  images: string;
  is_show: boolean;
  created_at: string;
}

// 菜单类型
export interface Menu {
  id: number;
  title: string;
  path: string;
  slogan?: string;
  abstract?: string[];
  abstract_time?: number;
  banner_time?: number;
  sort: number;
  banners?: Banner[];
  created_at: string;
}

export interface Banner {
  id: number;
  path: string;
}

// 消息类型
export interface Message {
  send_user_id: number;
  send_user_nick_name: string;
  send_user_avatar?: string;
  rev_user_id: number;
  rev_user_nick_name: string;
  rev_user_avatar?: string;
  content: string;
  message_count: number;
  created_at: string;
}

// 图片类型
export interface Image {
  id: number;
  path: string;
  name: string;
  image_type: ImageType;
  created_at: string;
}

export enum ImageType {
  Local = 1,
  QiNiu = 2
}

// 日志类型
export interface Log {
  id: number;
  ip: string;
  addr?: string;
  level: number;
  content: string;
  user_id?: number;
  created_at: string;
}

// 统计数据类型
export interface DataSum {
  user_count: number;
  article_count: number;
  message_count: number;
  chat_group_count: number;
  now_login_count: number;
  now_sign_count: number;
}

export interface DateCountData {
  date_list: string[];
  login_data: number[];
  sign_data: number[];
}
```

---

## API 请求封装

### Axios 封装示例

```typescript
// utils/request.ts
import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { ApiResponse } from '@/types/api';

const request: AxiosInstance = axios.create({
  baseURL: 'http://127.0.0.1:8080/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.token = token;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { code, msg } = response.data;
    
    if (code !== 0) {
      // 处理业务错误
      console.error(`API Error: ${msg}`);
      return Promise.reject(new Error(msg));
    }
    
    return response.data.data;
  },
  (error) => {
    // 处理 HTTP 错误
    if (error.response) {
      const { status } = error.response;
      switch (status) {
        case 401:
          // Token 过期，清除 token 并跳转登录
          localStorage.removeItem('token');
          window.location.href = '/login';
          break;
        case 403:
          console.error('权限不足');
          break;
        case 404:
          console.error('资源不存在');
          break;
        default:
          console.error('服务器错误');
      }
    }
    return Promise.reject(error);
  }
);

export default request;
```

---

### API 模块封装

```typescript
// api/user.ts
import request from '@/utils/request';
import { User, PageData, PageParams, ApiResponse } from '@/types/api';

interface LoginParams {
  user_name: string;
  password: string;
}

interface UpdatePasswordParams {
  old_password: string;
  new_password: string;
}

interface UpdateNicknameParams {
  nick_name: string;
}

export const userApi = {
  // 登录
  login: (params: LoginParams) => 
    request.post<string>('/email_login', params),
  
  // 获取用户信息
  getUserInfo: () => 
    request.get<User>('/user_info'),
  
  // 获取用户列表
  getUserList: (params: PageParams) => 
    request.get<PageData<User>>('/users', { params }),
  
  // 修改密码
  updatePassword: (params: UpdatePasswordParams) => 
    request.put('/user_password', params),
  
  // 修改昵称
  updateNickname: (params: UpdateNicknameParams) => 
    request.put('/user_update_nick_name', params),
  
  // 注销登录
  logout: () => 
    request.post('/logout'),
  
  // 创建用户（管理员）
  createUser: (params: Partial<User> & { password: string }) => 
    request.post('/user_create', params),
  
  // 删除用户（管理员）
  removeUsers: (id_list: number[]) => 
    request.delete('/users', { data: { id_list } })
};
```

```typescript
// api/article.ts
import request from '@/utils/request';
import { Article, PageData, PageParams, ApiResponse } from '@/types/api';

interface ArticleSearchParams extends PageParams {
  tag?: string;
  is_user?: boolean;
}

interface CreateArticleParams {
  title: string;
  abstract?: string;
  content: string;
  category?: string;
  source?: string;
  link?: string;
  banner_id?: number;
  tags?: string[];
}

interface UpdateArticleParams extends CreateArticleParams {
  id: string;
}

export const articleApi = {
  // 获取文章列表
  getArticleList: (params: ArticleSearchParams) => 
    request.get<PageData<Article>>('/articles', { params }),
  
  // 获取文章详情
  getArticleDetail: (id: string) => 
    request.get<Article>(`/articles/${id}`),
  
  // 根据标题获取文章
  getArticleByTitle: (title: string) => 
    request.get<Article>('/articles/detail', { params: { title } }),
  
  // 创建文章
  createArticle: (params: CreateArticleParams) => 
    request.post('/articles', params),
  
  // 更新文章
  updateArticle: (params: UpdateArticleParams) => 
    request.put('/articles', params),
  
  // 删除文章
  removeArticles: (id_list: string[]) => 
    request.delete('/articles', { data: { id_list } }),
  
  // 搜索文章
  searchArticles: (key: string, params: PageParams) => 
    request.get<PageData<Article>>('/articles/search', { 
      params: { ...params, key } 
    }),
  
  // 获取文章标签
  getArticleTags: () => 
    request.get<PageData<{ tag: string; count: number }>>('/articles/tags'),
  
  // 获取文章分类
  getArticleCategories: () => 
    request.get<{ label: string; value: string }[]>('/articles/categorys'),
  
  // 获取文章日历
  getArticleCalendar: () => 
    request.get<{ date: string; count: number }[]>('/articles/calendar'),
  
  // 收藏文章
  collectArticle: (id: string) => 
    request.post('/articles/collects', { id }),
  
  // 获取收藏列表
  getCollectList: (params: PageParams) => 
    request.get<PageData<Article>>('/articles/collects', { params }),
  
  // 取消收藏
  removeCollects: (id_list: string[]) => 
    request.delete('/articles/collects/batch', { data: { id_list } }),
  
  // 点赞文章
  diggArticle: (id: string) => 
    request.post('/digg', { id })
};
```

```typescript
// api/tag.ts
import request from '@/utils/request';
import { Tag, PageData, PageParams } from '@/types/api';

export const tagApi = {
  // 获取标签列表
  getTagList: (params: PageParams) => 
    request.get<PageData<Tag>>('/tags', { params }),
  
  // 获取标签名称列表
  getTagNames: () => 
    request.get<Pick<Tag, 'id' | 'tag'>[]>('/tags/names'),
  
  // 创建标签
  createTag: (tag: string) => 
    request.post('/tags', { tag }),
  
  // 更新标签
  updateTag: (id: number, tag: string) => 
    request.put(`/tags/${id}`, { tag }),
  
  // 删除标签
  removeTags: (id_list: number[]) => 
    request.delete('/tags', { data: { id_list } })
};
```

```typescript
// api/comment.ts
import request from '@/utils/request';
import { Comment } from '@/types/api';

interface CreateCommentParams {
  article_id: string;
  content: string;
  parent_comment_id?: number;
}

export const commentApi = {
  // 获取评论列表
  getCommentList: (article_id: string) => 
    request.get<Comment[]>('/comments', { params: { article_id } }),
  
  // 发表评论
  createComment: (params: CreateCommentParams) => 
    request.post('/comments', params),
  
  // 删除评论
  removeComment: (id: number) => 
    request.delete(`/comments/${id}`),
  
  // 点赞评论
  diggComment: (id: number) => 
    request.post(`/comments/digg/${id}`)
};
```

---

## React/Vue 集成示例

### React Hooks 封装

```typescript
// hooks/useUser.ts
import { useState, useEffect } from 'react';
import { userApi } from '@/api/user';
import { User } from '@/types/api';

export const useUser = () => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchUserInfo = async () => {
    setLoading(true);
    try {
      const data = await userApi.getUserInfo();
      setUser(data);
    } catch (err) {
      setError(err as Error);
    } finally {
      setLoading(false);
    }
  };

  const updateNickname = async (nick_name: string) => {
    await userApi.updateNickname({ nick_name });
    await fetchUserInfo();
  };

  useEffect(() => {
    fetchUserInfo();
  }, []);

  return {
    user,
    loading,
    error,
    refresh: fetchUserInfo,
    updateNickname
  };
};
```

```typescript
// hooks/useArticles.ts
import { useState, useEffect, useCallback } from 'react';
import { articleApi } from '@/api/article';
import { Article, PageData, PageParams } from '@/types/api';

export const useArticles = (initialParams: PageParams = {}) => {
  const [articles, setArticles] = useState<PageData<Article> | null>(null);
  const [loading, setLoading] = useState(false);
  const [params, setParams] = useState<PageParams>({
    page: 1,
    limit: 10,
    ...initialParams
  });

  const fetchArticles = useCallback(async () => {
    setLoading(true);
    try {
      const data = await articleApi.getArticleList(params);
      setArticles(data);
    } finally {
      setLoading(false);
    }
  }, [params]);

  const searchArticles = async (key: string) => {
    setLoading(true);
    try {
      const data = await articleApi.searchArticles(key, params);
      setArticles(data);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchArticles();
  }, [fetchArticles]);

  return {
    articles,
    loading,
    params,
    setParams,
    refresh: fetchArticles,
    searchArticles
  };
};
```

---

### Vue Composables 封装

```typescript
// composables/useUser.ts
import { ref, onMounted } from 'vue';
import { userApi } from '@/api/user';
import { User } from '@/types/api';

export const useUser = () => {
  const user = ref<User | null>(null);
  const loading = ref(false);
  const error = ref<Error | null>(null);

  const fetchUserInfo = async () => {
    loading.value = true;
    try {
      user.value = await userApi.getUserInfo();
    } catch (err) {
      error.value = err as Error;
    } finally {
      loading.value = false;
    }
  };

  const updateNickname = async (nick_name: string) => {
    await userApi.updateNickname({ nick_name });
    await fetchUserInfo();
  };

  onMounted(() => {
    fetchUserInfo();
  });

  return {
    user,
    loading,
    error,
    refresh: fetchUserInfo,
    updateNickname
  };
};
```

---

## 自动生成工具推荐

### 1. Orval（推荐用于 React/Vue）

```bash
# 安装
npm install orval -D

# 配置文件 orval.config.js
module.exports = {
  gvb: {
    output: {
      target: './src/api/generated.ts',
      client: 'react-query', // 或 'vue-query', 'swr', 'axios'
      mock: true,
    },
    input: {
      target: './docs/swagger.json',
    },
  },
};

# 生成代码
npx orval
```

---

### 2. RTK Query Codegen

```bash
# 安装
npm install @rtk-query/codegen-openapi -D

# 配置文件 openapi-config.ts
import { ConfigFile } from '@rtk-query/codegen-openapi';

const config: ConfigFile = {
  schemaFile: './docs/swagger.json',
  apiFile: './src/store/emptyApi.ts',
  apiImport: 'emptySplitApi',
  outputFile: './src/store/gvbApi.ts',
  exportName: 'gvbApi',
  hooks: true,
};

export default config;

# 生成代码
npx @rtk-query/codegen-openapi openapi-config.ts
```

---

### 3. Swagger Typescript API

```bash
# 安装
npm install swagger-typescript-api -D

# 生成代码
npx swagger-typescript-api \
  -p ./docs/swagger.json \
  -o ./src/api \
  -n gvbApi.ts \
  --axios
```

---

## 完整项目结构示例

```
frontend/
├── src/
│   ├── api/
│   │   ├── index.ts          # API 导出
│   │   ├── user.ts           # 用户 API
│   │   ├── article.ts        # 文章 API
│   │   ├── tag.ts            # 标签 API
│   │   ├── comment.ts        # 评论 API
│   │   └── ...               # 其他 API
│   ├── types/
│   │   ├── api.ts            # API 类型定义
│   │   └── index.ts          # 类型导出
│   ├── utils/
│   │   └── request.ts        # Axios 封装
│   ├── hooks/                # React Hooks
│   │   ├── useUser.ts
│   │   ├── useArticles.ts
│   │   └── ...
│   ├── composables/          # Vue Composables
│   │   ├── useUser.ts
│   │   └── ...
│   └── ...
├── docs/
│   ├── swagger.json          # Swagger 文档
│   └── swagger.yaml
└── ...
```

---

*文档生成时间: 2026-03-01*
