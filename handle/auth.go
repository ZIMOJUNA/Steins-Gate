package handle

import (
	"github.com/Future-Game-Laboratory/Steins-Gate/service"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type sendEmailCodeRequest struct {
	Email string `json:"email"`
	Scene string `json:"scene"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Code     string `json:"code"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) SendEmailCode(c fiber.Ctx) error {
	var req sendEmailCodeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return serviceError(c, service.ErrInvalidInput)
	}

	if err := h.auth.SendEmailCode(c.Context(), req.Email, req.Scene); err != nil {
		return serviceError(c, err)
	}

	return ok(c, fiber.Map{"sent": true})
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req registerRequest
	if err := c.Bind().JSON(&req); err != nil {
		return serviceError(c, service.ErrInvalidInput)
	}

	result, err := h.auth.Register(c.Context(), req.Email, req.Password, req.Nickname, req.Code)
	if err != nil {
		return serviceError(c, err)
	}

	return ok(c, result)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return serviceError(c, service.ErrInvalidInput)
	}

	result, err := h.auth.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return serviceError(c, err)
	}

	return ok(c, result)
}

func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req resetPasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return serviceError(c, service.ErrInvalidInput)
	}

	result, err := h.auth.ResetPassword(c.Context(), req.Email, req.Code, req.NewPassword)
	if err != nil {
		return serviceError(c, err)
	}

	return ok(c, result)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	token := bearerToken(c.Get(fiber.HeaderAuthorization))
	if err := h.auth.Logout(c.Context(), token); err != nil {
		return serviceError(c, err)
	}

	return noContent(c)
}

func (h *AuthHandler) Me(c fiber.Ctx) error {
	account := currentAccount(c)
	if account == nil {
		return serviceError(c, service.ErrUnauthorized)
	}

	return ok(c, account)
}
