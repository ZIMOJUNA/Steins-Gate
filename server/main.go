package main

import (
	"fmt"
	"server/mongodb"
	"server/redis"
	"time"

	"server/config"
	"server/handle"

	"github.com/gofiber/fiber/v3"
)

func test() {
	mongodb.GetClient()
	redis.GenerateToken("Apple", 100*time.Second)
	redis.GenerateToken("Applsaddsasde", 100*time.Second)
	redis.GenerateToken("Applsaddsas32423de", 100*time.Second)
}

func main() {

	test()

	// 创建 Fiber 应用
	app := fiber.New()

	app.Get("/", handle.HelloWorld)
	app.Post("/user", handle.HelloWorld)

	port := config.Conf.Server.Port
	addr := fmt.Sprintf(":%d", port)

	// 启动服务
	err := app.Listen(addr)
	if err != nil {
		panic("服务启动失败：" + err.Error())
	}
}
