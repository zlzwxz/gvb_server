package file_api

import "gvb-server/models/res"

var _ = res.Response{}

// FileUploadViewDoc 上传文章附件。
// @Tags 文件管理
// @Summary 上传文章附件
// @Description 仅允许上传附件白名单中的文件类型。上传成功后返回附件 ID 和下载地址，供文章 attachments 字段引用。
// @Accept multipart/form-data
// @Produce json
// @Param token header string true "token"
// @Param file formData file true "附件文件"
// @Success 200 {object} res.Response{data=FileUploadItem} "上传成功"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/files [post]
func FileUploadViewDoc() {}

// FileDownloadViewDoc 下载文章附件。
// @Tags 文件管理
// @Summary 下载文章附件
// @Description 文章作者、管理员，以及已公开文章引用的附件都可以下载。下载时会做权限校验，不建议直接暴露物理路径。
// @Accept json
// @Produce application/octet-stream
// @Param token header string true "token"
// @Param id path int true "附件 ID"
// @Success 200 {file} binary "文件流"
// @Failure 400 {object} res.Response "请求参数错误"
// @Failure 401 {object} res.Response "未授权"
// @Router /api/files/{id}/download [get]
func FileDownloadViewDoc() {}
