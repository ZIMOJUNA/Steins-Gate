package main

import (
	"fmt"
	"log"

	"github.com/Future-Game-Laboratory/Steins-Gate/service"

	"github.com/Future-Game-Laboratory/Steins-Gate/config"
	"github.com/Future-Game-Laboratory/Steins-Gate/handle"
	"github.com/Future-Game-Laboratory/Steins-Gate/mailer"
	"github.com/Future-Game-Laboratory/Steins-Gate/mysql"
	"github.com/Future-Game-Laboratory/Steins-Gate/redis"
	fiber "github.com/gofiber/fiber/v3"
)

func main() {
	if err := mysql.Init(); err != nil {
		log.Fatal(err)
	}
	defer mysql.Close()

	if err := redis.Init(); err != nil {
		log.Fatal(err)
	}

	mailSender, err := mailer.NewSender(config.Conf.Mail)
	if err != nil {
		log.Fatal(err)
	}
	authSvc := service.NewAuthService(mailSender)
	dataSvc := service.NewPlayerDataService()

	// 创建 Fiber 应用
	app := fiber.New()

	handle.RegisterRoutes(app, authSvc, dataSvc)

	port := config.Conf.Server.Port
	addr := fmt.Sprintf(":%d", port)

	// 启动服务
	err = app.Listen(addr)
	if err != nil {
		panic("服务启动失败：" + err.Error())
	}
}
