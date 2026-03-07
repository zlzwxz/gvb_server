package core

import (
	"fmt"
	"gvb-server/config"
	"gvb-server/global"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// ConfigFlis 指向项目默认配置文件路径。
// 当前项目约定配置文件固定叫 `settings.yaml`，放在后端根目录。
const ConfigFlis = "settings.yaml"

// InitConf 负责把 YAML 配置读进内存，并保存到 `global.Config`。
// 它是后端最早执行的初始化步骤之一，因为后面的日志、数据库、Redis 等都要依赖配置。
func InitConf() {
	// 创建一个配置结构体实例，后面 YAML 会反序列化到这里。
	c := &config.Config{}

	// 从磁盘读取 YAML 文件内容。
	yamlConfig, err := ioutil.ReadFile(ConfigFlis)
	if err != nil {
		panic(fmt.Errorf("get yaml file error : %s", err))
	}

	// 把 YAML 字节解析成 Go 结构体。
	err = yaml.Unmarshal(yamlConfig, c)
	if err != nil {
		panic(fmt.Errorf("yaml unmarshal error : %s", err))
	}

	// 允许用环境变量覆盖敏感配置，避免把生产密钥硬编码进仓库。
	applySecurityEnvOverrides(c)
	// 对关键安全项做最低限度校验，防止程序带着危险配置启动。
	validateSecurityConfig(c)

	// 保存到全局变量，供后续模块统一读取。
	global.Config = c
}

// SetYaml 把当前内存中的 `global.Config` 重新写回 `settings.yaml`。
// 这个函数通常用于“后台修改设置后回写到本地配置文件”的场景。
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

// applySecurityEnvOverrides 用环境变量覆盖配置中的敏感字段。
// 这样可以做到：仓库里保留通用配置，生产环境再通过环境变量注入真正密钥。
func applySecurityEnvOverrides(c *config.Config) {
	if c == nil {
		return
	}
	if value := strings.TrimSpace(os.Getenv("GVB_JWT_SECRET")); value != "" {
		c.Jwt.Secret = value
	}
	if value := strings.TrimSpace(os.Getenv("GVB_EMAIL_PASSWORD")); value != "" {
		c.Email.Password = value
	}
	if value := strings.TrimSpace(os.Getenv("GVB_MYSQL_PASSWORD")); value != "" {
		c.Mysql.Password = value
	}
}

// validateSecurityConfig 对关键安全配置做兜底检查。
// 这里不追求“非常全面”，主要避免最危险的低级配置错误。
func validateSecurityConfig(c *config.Config) {
	if c == nil {
		panic("配置为空")
	}
	jwtSecret := strings.TrimSpace(c.Jwt.Secret)
	if jwtSecret == "" {
		panic("jwt.secret 不能为空，请在 settings.yaml 或环境变量 GVB_JWT_SECRET 中配置")
	}
	if len(jwtSecret) < 16 {
		panic("jwt.secret 长度至少 16 位，建议使用高强度随机字符串")
	}
	if c.Jwt.Expires <= 0 {
		c.Jwt.Expires = 12
	}
}
