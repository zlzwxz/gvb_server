# GVB Server API Documentation (Latest)

This file is generated from `routers/*_router.go` and reflects the current backend route list.

## Base

- Base URL: `http://127.0.0.1:8080/api`
- Auth headers: `token: <token>` or `Authorization: Bearer <token>`
- Permission labels: `public`, `auth`, `admin`
- Detailed guide: `docs/API_DOCUMENTATION_DETAILED.md`
- Detailed Postman collection: `docs/GVB_Server_Detailed.postman_collection.json`
- Postman environment: `docs/GVB_Server.postman_environment.json`

## Endpoint List

| Module | Method | Path | Auth | Type |
| --- | --- | --- | --- | --- |
| advert | DELETE | /api/adverts | admin | write |
| advert | GET | /api/adverts | public | query |
| advert | POST | /api/adverts | admin | write |
| advert | PUT | /api/adverts/{id} | admin | write |
| announcement | DELETE | /api/announcements | admin | write |
| announcement | GET | /api/announcements | public | query |
| announcement | GET | /api/announcements/manage | admin | query |
| announcement | POST | /api/announcements | admin | write |
| announcement | PUT | /api/announcements/{id} | admin | write |
| article | DELETE | /api/articles | auth | write |
| article | DELETE | /api/articles/collects/batch | auth | write |
| article | DELETE | /api/articles/collects/manage | auth | write |
| article | GET | /api/article/text | public | query |
| article | GET | /api/articles | public | query |
| article | GET | /api/articles/calendar | public | query |
| article | GET | /api/articles/categorys | public | query |
| article | GET | /api/articles/collects | auth | query |
| article | GET | /api/articles/collects/manage | auth | query |
| article | GET | /api/articles/content/{id} | public | query |
| article | GET | /api/articles/detail | public | query |
| article | GET | /api/articles/insights | public | query |
| article | GET | /api/articles/reports | auth | query |
| article | GET | /api/articles/tags | public | query |
| article | GET | /api/articles/{id} | public | query |
| article | POST | /api/articles | auth | write |
| article | POST | /api/articles/collects | auth | write |
| article | POST | /api/articles/reports | auth | write |
| article | PUT | /api/articles | auth | write |
| article | PUT | /api/articles/reports | auth | write |
| article | PUT | /api/articles/review | auth | write |
| board | DELETE | /api/boards | admin | write |
| board | GET | /api/boards | public | query |
| board | POST | /api/boards | admin | write |
| board | PUT | /api/boards | admin | write |
| chat | GET | /api/chat_groups | public | query |
| chat | GET | /api/chat_groups_records | public | query |
| comment | DELETE | /api/comments/{id} | auth | write |
| comment | GET | /api/comments | public | query |
| comment | GET | /api/comments/{id} | public | query |
| comment | POST | /api/comments | auth | write |
| data | GET | /api/data_login | public | query |
| data | GET | /api/data_sum | public | query |
| data | GET | /api/data_sum/admin | admin | query |
| digg | POST | /api/article/digg | auth | write |
| file | GET | /api/files/{id}/download | auth | query |
| file | POST | /api/files | auth | write |
| images | DELETE | /api/images | auth | write |
| images | GET | /api/images | auth | query |
| images | POST | /api/images | auth | write |
| images | PUT | /api/images | auth | write |
| log | DELETE | /api/logs | admin | write |
| log | GET | /api/logs | admin | query |
| menu | DELETE | /api/menus | admin | write |
| menu | GET | /api/menu_names | public | query |
| menu | GET | /api/menus | public | query |
| menu | GET | /api/menus/{id} | public | query |
| menu | POST | /api/menus | admin | write |
| menu | PUT | /api/menus | admin | write |
| message | GET | /api/messages | auth | query |
| message | GET | /api/messages/all | admin | query |
| message | GET | /api/messages/record | auth | query |
| message | GET | /api/messages_all | admin | query |
| message | GET | /api/messages_record | auth | query |
| message | POST | /api/messages | auth | write |
| new | GET | /api/news | public | query |
| new | GET | /api/news/sources | public | query |
| settings | GET | /api/settings/es/export | admin | query |
| settings | GET | /api/settings/es/indices | admin | query |
| settings | GET | /api/settings/public/site_info | public | query |
| settings | GET | /api/settings/site_info/sync_fengfeng_images_preview | admin | query |
| settings | GET | /api/settings/site_info/sync_fengfeng_preview | admin | query |
| settings | GET | /api/settings/{name} | admin | query |
| settings | POST | /api/settings/es/import | admin | write |
| settings | POST | /api/settings/site_info/sync_fengfeng | admin | write |
| settings | POST | /api/settings/site_info/sync_fengfeng_images | admin | write |
| settings | PUT | /api/settings/{name} | admin | write |
| social | DELETE | /api/social/blocks/{id} | auth | write |
| social | DELETE | /api/social/follows/{id} | auth | write |
| social | DELETE | /api/social/groups/{id}/members/{user_id} | auth | write |
| social | GET | /api/social/blocks | auth | query |
| social | GET | /api/social/calls | auth | query |
| social | GET | /api/social/conversations | auth | query |
| social | GET | /api/social/direct/messages | auth | query |
| social | GET | /api/social/discovery | auth | query |
| social | GET | /api/social/files/{id}/download | auth | query |
| social | GET | /api/social/friends | auth | query |
| social | GET | /api/social/groups | auth | query |
| social | GET | /api/social/groups/{id} | auth | query |
| social | GET | /api/social/groups/{id}/messages | auth | query |
| social | GET | /api/social/manage/blocks | admin | query |
| social | GET | /api/social/manage/follows | admin | query |
| social | GET | /api/social/manage/groups | admin | query |
| social | GET | /api/social/manage/summary | admin | query |
| social | GET | /api/social/messages/search | auth | query |
| social | GET | /api/social/relations/{id} | auth | query |
| social | GET | /api/social/summary | auth | query |
| social | GET | /api/social/ws | public | query |
| social | POST | /api/social/blocks/{id} | auth | write |
| social | POST | /api/social/direct/messages | auth | write |
| social | POST | /api/social/files | auth | write |
| social | POST | /api/social/follows/{id} | auth | write |
| social | POST | /api/social/groups | auth | write |
| social | POST | /api/social/groups/join | auth | write |
| social | POST | /api/social/groups/{id}/members | auth | write |
| social | POST | /api/social/groups/{id}/messages | auth | write |
| social | POST | /api/social/messages/{id}/recall | auth | write |
| social | PUT | /api/social/groups/{id}/members/{user_id}/role | auth | write |
| social | PUT | /api/social/groups/{id}/transfer-owner | auth | write |
| social | PUT | /api/social/presence | auth | write |
| tag | DELETE | /api/tags | admin | write |
| tag | GET | /api/tags | public | query |
| tag | GET | /api/tags/names | public | query |
| tag | POST | /api/tags | admin | write |
| tag | PUT | /api/tags/{id} | admin | write |
| user | DELETE | /api/space/messages/{id} | auth | write |
| user | DELETE | /api/space/posts/{id} | auth | write |
| user | DELETE | /api/users | admin | write |
| user | GET | /api/qq_login_path | public | query |
| user | GET | /api/user_check_in_status | auth | query |
| user | GET | /api/user_info | auth | query |
| user | GET | /api/user_level_rank | public | query |
| user | GET | /api/users | auth | query |
| user | GET | /api/users/{id}/profile | public | query |
| user | GET | /api/users/{id}/space/messages | public | query |
| user | GET | /api/users/{id}/space/posts | public | query |
| user | POST | /api/email_login | public | write |
| user | POST | /api/logout | auth | write |
| user | POST | /api/qq_login | public | write |
| user | POST | /api/space/messages | auth | write |
| user | POST | /api/space/posts | auth | write |
| user | POST | /api/user_bind_email | auth | write |
| user | POST | /api/user_check_in | auth | write |
| user | POST | /api/user_create | public | write |
| user | POST | /api/user_register_email_code | public | write |
| user | PUT | /api/user_password | admin | write |
| user | PUT | /api/user_role | admin | write |
| user | PUT | /api/user_update_nick_name | auth | write |

## Notes

1. `docs/API_DOCUMENTATION_DETAILED.md` contains test data, request examples, response examples, logic explanations, and call precautions for the newly completed and commonly used endpoints.
2. `docs/GVB_Server_Detailed.postman_collection.json` can be imported directly into Postman.
3. `docs/GVB_Server.postman_environment.json` contains local variables for `baseUrl`, `token`, `adminToken`, IDs, and upload file paths.
