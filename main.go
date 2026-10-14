package main

import (
	"log"

	"github.com/yijiacode188/initkit/lib"
)

func main() {
	err := lib.LoadConfig(&lib.ConfigParams{
		ConfigPath: "./conf/config.dev.yml",
	})
	if err != nil {
		log.Fatalf("获取配置文件失败: %v", err)
		return
	}
	defer lib.Destroy()
}
