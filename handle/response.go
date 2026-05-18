package handle

import (
	"errors"

	"github.com/Future-Game-Laboratory/Steins-Gate/service"
	"github.com/gofiber/fiber/v3"
)

type response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func ok(c fiber.Ctx, data any) error {
	return c.JSON(response{
		Code:    "ok",
		Message: "ok",
		Data:    data,
	})
}

func noContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func fail(c fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(response{
		Code:    code,
		Message: message,
	})
}

func serviceError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return fail(c, fiber.StatusBadRequest, "invalid_input", "请求参数不合法")
	case errors.Is(err, service.ErrEmailExists):
		return fail(c, fiber.StatusConflict, "email_exists", "邮箱已注册")
	case errors.Is(err, service.ErrAccountNotFound):
		return fail(c, fiber.StatusNotFound, "account_not_found", "账号不存在")
	case errors.Is(err, service.ErrAccountDisabled):
		return fail(c, fiber.StatusForbidden, "account_disabled", "账号不可用")
	case errors.Is(err, service.ErrInvalidCredentials):
		return fail(c, fiber.StatusUnauthorized, "invalid_credentials", "邮箱或密码错误")
	case errors.Is(err, service.ErrTooManyRequests):
		return fail(c, fiber.StatusTooManyRequests, "too_many_requests", "操作过于频繁，请稍后再试")
	case errors.Is(err, service.ErrInvalidCode):
		return fail(c, fiber.StatusBadRequest, "invalid_code", "验证码错误")
	case errors.Is(err, service.ErrCodeExpired):
		return fail(c, fiber.StatusBadRequest, "code_expired", "验证码不存在或已过期")
	case errors.Is(err, service.ErrCodeTooManyAttempts):
		return fail(c, fiber.StatusTooManyRequests, "code_too_many_attempts", "验证码错误次数过多，请重新获取")
	case errors.Is(err, service.ErrUnauthorized):
		return fail(c, fiber.StatusUnauthorized, "unauthorized", "未登录或登录已过期")
	case errors.Is(err, service.ErrSaveNotFound):
		return fail(c, fiber.StatusNotFound, "player_data_not_found", "存档不存在")
	default:
		return fail(c, fiber.StatusInternalServerError, "internal_error", "服务器内部错误")
	}
}
