package handle

import (
	"strings"

	"github.com/Future-Game-Laboratory/Steins-Gate/service"
	"github.com/gofiber/fiber/v3"
)

const (
	localAccountID = "account_id"
	localAccount   = "account"
)

func AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		token := bearerToken(c.Get(fiber.HeaderAuthorization))
		account, err := service.AuthenticateToken(c.Context(), token)
		if err != nil {
			return serviceError(c, err)
		}

		c.Locals(localAccountID, account.ID)
		c.Locals(localAccount, account)

		return c.Next()
	}
}

func accountID(c fiber.Ctx) uint64 {
	value, _ := c.Locals(localAccountID).(uint64)
	return value
}

func currentAccount(c fiber.Ctx) *service.Account {
	account, _ := c.Locals(localAccount).(*service.Account)
	return account
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(header, prefix))
	}
	return header
}
