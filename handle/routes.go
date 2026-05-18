package handle

import (
	"github.com/Future-Game-Laboratory/Steins-Gate/service"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(app fiber.Router, authSvc *service.AuthService, dataSvc *service.PlayerDataService) {
	authHandler := NewAuthHandler(authSvc)
	dataHandler := NewPlayerDataHandler(dataSvc)

	app.Get("/", HelloWorld)
	app.Get("/health", Health)

	api := app.Group("/api/v1")
	api.Post("/auth/email-code", authHandler.SendEmailCode)
	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/password/reset", authHandler.ResetPassword)

	protected := api.Group("", AuthMiddleware())
	protected.Get("/me", authHandler.Me)
	protected.Post("/auth/logout", authHandler.Logout)
	protected.Get("/player-data", dataHandler.List)
	protected.Post("/player-data", dataHandler.Upsert)
	protected.Get("/player-data/:id", dataHandler.Get)
	protected.Put("/player-data/:id", dataHandler.Update)
	protected.Delete("/player-data/:id", dataHandler.Delete)
}
