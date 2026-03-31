package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var Conf Config

// LoadConfig 加载配置文件
func LoadConfig(path string) {
	// 读取文件
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("加载配置文件失败：%v", err)
		return
	}
	// 解析yaml到结构体
	err = yaml.Unmarshal(data, &Conf)
	if err != nil {
		log.Fatalf("解析配置文件失败：%v", err)
		return
	}

	// debugPrint()
}

func debugPrint() {
	// 1. 加载配置
	log.Println("配置文件加载成功！")

	// 2. 直接使用配置
	log.Printf("服务器端口：%d", Conf.Server.Port)
	log.Printf("运行模式：%s", Conf.Server.Mode)
	log.Printf("Redis地址：%s:%d", Conf.Redis.Host, Conf.Redis.Port)
	log.Printf("MySQL DSN：%s", Conf.MySQL.DSN)
}
