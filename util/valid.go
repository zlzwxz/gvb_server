package util

import (
	"github.com/go-playground/validator/v10"
	"reflect"
)

// GetValidMsg 返回结构体中的msg参数
func GetValidMsg(err error, obj any) string {
	//使用时需要传obj指针
	getObj := reflect.TypeOf(obj)
	if errs, ok := err.(validator.ValidationErrors); ok {
		//断言成功
		for _, e := range errs {
			//遍历结构体的字段
			if f, exits := getObj.Elem().FieldByName(e.Field()); exits {
				msg := f.Tag.Get("msg")
				//返回msg字段的值
				return msg
			}
		}
	}
	return err.Error()
}
