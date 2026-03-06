package service

import (
	"gvb-server/service/image_ser"
	"gvb-server/service/user_ser"
)

// ServiceGroup 聚合所有 service 层对象。
// API 层需要调用业务逻辑时，统一从这里取具体 service。
type ServiceGroup struct {
	ImageService image_ser.ImageService
	UserService  user_ser.UserService
}

// ServiceApp 是 service 层的全局单例入口。
var ServiceApp = new(ServiceGroup)
