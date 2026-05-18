package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var Conf Config

func init() {
	// 读取文件
	data, err := readConfigFile()
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

	if Conf.Server.Mode == "debug" {
		debugPrint()
	}
}

func debugPrint() {
	// 1. 加载配置
	log.Println("配置文件加载成功！")

	// 2. 直接使用配置
	log.Printf("服务器端口：%d", Conf.Server.Port)
	log.Printf("运行模式：%s", Conf.Server.Mode)
	log.Printf("Redis地址：%s:%d", Conf.Redis.Host, Conf.Redis.Port)
	log.Printf("MySQL地址：%s:%d  数据库：%s", Conf.MySQL.Host, Conf.MySQL.Port, Conf.MySQL.DBName)
	log.Printf("登录Token有效期：%s", Conf.Auth.TokenDuration())
	log.Printf("邮箱验证码有效期：%s", Conf.Auth.EmailCodeDuration())
	log.Printf("邮件Provider：%s", Conf.Mail.Provider)
}

func readConfigFile() ([]byte, error) {
	if path := os.Getenv("STEINS_GATE_CONFIG"); path != "" {
		return os.ReadFile(path)
	}

	candidates := []string{
		filepath.Join("config", "config.yaml"),
		filepath.Join("..", "config", "config.yaml"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}

	return nil, os.ErrNotExist
}
