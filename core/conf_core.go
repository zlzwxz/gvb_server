package core

import (
	"fmt"
	"gvb-server/config"
	"gvb-server/global"
	"io/fs"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// 文件路径
const ConfigFlis = "settings.yaml"

// 初始化读取yaml文件
func InitConf() {
	//指向对应config.enter里面的config结构体指针
	c := &config.Config{}
	//读取文件
	yamlConfig, err := ioutil.ReadFile(ConfigFlis)
	//报错直接退出
	if err != nil {
		panic(fmt.Errorf("get yaml file error : %s", err))
	}
	err = yaml.Unmarshal(yamlConfig, c)
	if err != nil {
		panic(fmt.Errorf("yaml unmarshal error : %s", err))
	}
	//赋值给全局变量global.Config
	global.Config = c
}

// 修改yaml文件方法
func SetYaml() {
	ByteData, err := yaml.Marshal(global.Config)
	if err != nil {
		global.Log.Error("转换配置文件失败,yaml.Marshal error:", err)
		return
	}
	err = ioutil.WriteFile(ConfigFlis, ByteData, fs.ModePerm)
	if err != nil {
		global.Log.Error("写入配置文件失败,yaml.Marshal error:", err)
		return
	}
	global.Log.Info("写入配置文件成功")
}
