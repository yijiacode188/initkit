package lib

import (
	"github.com/spf13/viper"
	"time"
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

// IsSetConf 是否设置了key
func IsSetConf(key string) bool {
	return viperConf.IsSet(key)
}

// GetStringConf 获取配置
func GetStringConf(key string) string {
	if key == "" {
		return ""
	}
	return viperConf.GetString(key)
}

// GetStringMapConf 获取get配置信息
func GetStringMapConf(key string) map[string]interface{} {
	if key == "" {
		return nil
	}
	conf := viperConf.GetStringMap(key)
	return conf
}

// GetConf 获取get配置信息
func GetConf(key string) interface{} {
	conf := viperConf.Get(key)
	return conf
}

// GetBoolConf 获取get配置信息
func GetBoolConf(key string) bool {
	conf := viperConf.GetBool(key)
	return conf
}

// GetFloat64Conf 获取get配置信息
func GetFloat64Conf(key string) float64 {
	conf := viperConf.GetFloat64(key)
	return conf
}

// GetIntConf 获取get配置信息
func GetIntConf(key string) int {
	conf := viperConf.GetInt(key)
	return conf
}

// GetStringMapStringConf 获取get配置信息
func GetStringMapStringConf(key string) map[string]string {

	conf := viperConf.GetStringMapString(key)
	return conf
}

// GetStringSliceConf 获取get配置信息
func GetStringSliceConf(key string) []string {

	conf := viperConf.GetStringSlice(key)
	return conf
}

// GetTimeConf 获取get配置信息
func GetTimeConf(key string) time.Time {
	conf := viperConf.GetTime(key)
	return conf
}
