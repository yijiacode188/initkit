package lib

import (
	"encoding/binary"
	"flag"
	"log"
	"net"
	"os"
	"time"
)

var timeLocation *time.Location
var timeFormat = "2006-01-02 15:04:05"
var dateFormat = "2006-01-02"
var localIP = net.ParseIP("127.0.0.1")

type ConfigParams struct {
	ConfigPath string     `json:"configPath"`
	Logger     LoggerType `json:"logger"`
}

// LoadConfig 加载配置项
func LoadConfig(params *ConfigParams) error {
	conf := flag.String("config", params.ConfigPath, "input config file like ./conf/config.dev.yml")
	flag.Parse()
	if *conf == "" {
		flag.Usage()
		os.Exit(1)
	}
	log.Println("------------------------------------------------------------------------")
	log.Printf("[INFO]  config=%s\n", *conf)
	log.Printf("[INFO] %s\n", " start loading resources.")
	// 解析配置文件目录

	//初始化配置文件
	err := initViperConf(*conf)
	if err != nil {
		return err
	}
	if params.Logger != nil {
		Log = params.Logger
	} else {
		Log = NewLogger()
		//初始化日志写入
		err = initLogger()
		if err != nil {
			return err
		}
	}

	// 是否有数据库配置，如果有则初始化数据库
	if IsSetConf("datasource") {
		//添加了数据库配置
		err = initDatasource()
		if err != nil {
			return err
		}
	}
	if IsSetConf("redis") {
		err = initRedis()
		if err != nil {
			return err
		}
	}
	if IsSetConf("i18n") {
		err = initI18n()
		if err != nil {
			return err
		}
	}

	if GetStringConf("base.local_ip") != "" {
		localIP = net.ParseIP(GetStringConf("base.local_ip"))
	}
	timeLocationStr := GetStringConf("base.time_location")
	if timeLocationStr == "" {
		timeLocationStr = "Asia/Chongqing"
	}
	location, err := time.LoadLocation(GetStringConf("base.time_location"))
	if err != nil {
		return err
	}
	timeLocation = location
	//初始化雪花算法
	err = initSnowflake()
	if err != nil {
		return err
	}
	log.Printf("[INFO] %s\n", " success loading resources.")
	log.Println("------------------------------------------------------------------------")
	return nil
}

// Destroy 公共销毁函数
func Destroy() {
	log.Println("------------------------------------------------------------------------")
	log.Printf("[INFO] %s\n", " start destroy resources.")
	closeDB()
	closeRedis()
	log.Printf("[INFO] %s\n", " success destroy resources.")
}

// IPv4地址转换为uint32
func ipv4ToUint16(ip net.IP) uint16 {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0
	}
	return binary.BigEndian.Uint16(ipv4)
}

// GetCurrentTime 获取当前时间
func GetCurrentTime() time.Time {
	b1 := time.Now().UTC()
	return b1.In(timeLocation)
}
