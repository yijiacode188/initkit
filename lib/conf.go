package lib

import (
	"github.com/spf13/viper"
)

var viperConf *viper.Viper

// InitViperConf 初始化配置文件
func initViperConf(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	viperConf = v
	return nil
}

// GetStringConf 获取配置
func GetStringConf(key string) string {
	if key == "" {
		return ""
	}
	return viperConf.GetString(key)
}

// IsSetConf 是否设置了key
func IsSetConf(key string) bool {
	return viperConf.IsSet(key)
}
