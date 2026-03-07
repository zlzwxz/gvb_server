# GVB Server API Documentation (Markdown)

This file is generated from `routers/*_router.go` and covers all backend endpoints.

## Base

- Base URL: `http://127.0.0.1:8080/api`
- Auth headers: `Authorization: Bearer <token>` and `token: <token>`
- Permission label: `public`, `auth`, `admin`

## Endpoint List

| Module | Method | Path | Auth | Type |
| --- | --- | --- | --- | --- |
| advert | DELETE | /api/adverts | admin | write |
| advert | GET | /api/adverts | public | query |
| advert | POST | /api/adverts | admin | write |
| advert | PUT | /api/adverts/{{id}} | admin | write |
| article | GET | /api/article/text | public | query |
| article | DELETE | /api/articles | auth | write |
| article | GET | /api/articles | public | query |
| article | POST | /api/articles | auth | write |
| article | PUT | /api/articles | auth | write |
| article | GET | /api/articles/{{id}} | public | query |
| article | GET | /api/articles/calendar | public | query |
| article | GET | /api/articles/categorys | public | query |
| article | GET | /api/articles/collects | auth | query |
| article | POST | /api/articles/collects | auth | write |
| article | DELETE | /api/articles/collects/batch | auth | write |
| article | DELETE | /api/articles/collects/manage | auth | write |
| article | GET | /api/articles/collects/manage | auth | query |
| article | GET | /api/articles/content/{{id}} | public | query |
| article | GET | /api/articles/detail | public | query |
| article | GET | /api/articles/insights | public | query |
| article | PUT | /api/articles/review | admin | write |
| article | GET | /api/articles/tags | public | query |
| chat | GET | /api/chat_groups | public | query |
| chat | GET | /api/chat_groups_records | public | query |
| comment | GET | /api/comments | public | query |
| comment | POST | /api/comments | auth | write |
| comment | DELETE | /api/comments/{{id}} | auth | write |
| comment | GET | /api/comments/{{id}} | public | query |
| data | GET | /api/data_login | public | query |
| data | GET | /api/data_sum | public | query |
| digg | POST | /api/article/digg | auth | write |
| file | POST | /api/files | auth | write |
| file | GET | /api/files/{{id}}/download | auth | query |
| images | DELETE | /api/images | auth | write |
| images | GET | /api/images | auth | query |
| images | POST | /api/images | auth | write |
| images | PUT | /api/images | auth | write |
| log | DELETE | /api/logs | admin | write |
| log | GET | /api/logs | admin | query |
| menu | GET | /api/menu_names | public | query |
| menu | DELETE | /api/menus | admin | write |
| menu | GET | /api/menus | public | query |
| menu | POST | /api/menus | admin | write |
| menu | PUT | /api/menus | admin | write |
| menu | GET | /api/menus/{{id}} | public | query |
| message | GET | /api/messages | auth | query |
| message | POST | /api/messages | auth | write |
| message | GET | /api/messages_all | admin | query |
| message | GET | /api/messages_record | auth | query |
| message | GET | /api/messages/all | admin | query |
| message | GET | /api/messages/record | auth | query |
| new | GET | /api/news | public | query |
| new | GET | /api/news/sources | public | query |
| settings | GET | /api/settings/{{name}} | admin | query |
| settings | PUT | /api/settings/{{name}} | admin | write |
| settings | GET | /api/settings/public/site_info | public | query |
| settings | POST | /api/settings/site_info/sync_fengfeng | admin | write |
| tag | DELETE | /api/tags | admin | write |
| tag | GET | /api/tags | public | query |
| tag | POST | /api/tags | admin | write |
| tag | PUT | /api/tags/{{id}} | admin | write |
| tag | GET | /api/tags/names | public | query |
| user | POST | /api/email_login | public | write |
| user | POST | /api/logout | auth | write |
| user | POST | /api/qq_login | public | write |
| user | GET | /api/qq_login_path | public | query |
| user | POST | /api/user_bind_email | auth | write |
| user | POST | /api/user_check_in | auth | write |
| user | GET | /api/user_check_in_status | auth | query |
| user | POST | /api/user_create | public | write |
| user | GET | /api/user_info | auth | query |
| user | GET | /api/user_level_rank | public | query |
| user | PUT | /api/user_password | admin | write |
| user | POST | /api/user_register_email_code | public | write |
| user | PUT | /api/user_role | admin | write |
| user | PUT | /api/user_update_nick_name | auth | write |
| user | DELETE | /api/users | admin | write |
| user | GET | /api/users | auth | query |

## Postman

1. Import: `docs/GVB_Server.postman_collection.json`.
2. Fill variables: `baseUrl`, `token`, `articleId`.
3. Upload endpoints need: `uploadFilePath`, `uploadImagePath`.
4. Run login first, then run protected/admin requests.