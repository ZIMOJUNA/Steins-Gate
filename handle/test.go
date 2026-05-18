package handle

import (
	"github.com/gofiber/fiber/v3"
)

func HelloWorld(c fiber.Ctx) error {
	return c.SendString("Steins Gate cloud save service")
}

func Health(c fiber.Ctx) error {
	return ok(c, fiber.Map{"status": "ok"})
}
