package handle

import (
	"encoding/json"
	"strconv"

	"github.com/Future-Game-Laboratory/Steins-Gate/service"
	"github.com/gofiber/fiber/v3"
)

type PlayerDataHandler struct {
	data *service.PlayerDataService
}

func NewPlayerDataHandler(data *service.PlayerDataService) *PlayerDataHandler {
	return &PlayerDataHandler{data: data}
}

type upsertPlayerDataRequest struct {
	GameKey string          `json:"game_key"`
	SlotKey string          `json:"slot_key"`
	Data    json.RawMessage `json:"data"`
}

type updatePlayerDataRequest struct {
	Data json.RawMessage `json:"data"`
}

func (h *PlayerDataHandler) List(c fiber.Ctx) error {
	list, err := h.data.List(c.Context(), accountID(c), c.Query("game_key"))
	if err != nil {
		return serviceError(c, err)
	}

	return ok(c, fiber.Map{"items": list})
}

func (h *PlayerDataHandler) Get(c fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return serviceError(c, service.ErrInvalidInput)
	}

	item, err := h.data.Get(c.Context(), accountID(c), id)
	if err != nil {
		return serviceError(c, err)
	}

	return ok(c, item)
}

func (h *PlayerDataHandler) Upsert(c fiber.Ctx) error {
	var req upsertPlayerDataRequest
	if err := c.Bind().JSON(&req); err != nil {
		return serviceError(c, service.ErrInvalidInput)
	}

	item, err := h.data.Upsert(c.Context(), accountID(c), req.GameKey, req.SlotKey, req.Data)
	if err != nil {
		return serviceError(c, err)
	}

	return ok(c, item)
}

func (h *PlayerDataHandler) Update(c fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return serviceError(c, service.ErrInvalidInput)
	}

	var req updatePlayerDataRequest
	if err := c.Bind().JSON(&req); err != nil {
		return serviceError(c, service.ErrInvalidInput)
	}

	item, err := h.data.Update(c.Context(), accountID(c), id, req.Data)
	if err != nil {
		return serviceError(c, err)
	}

	return ok(c, item)
}

func (h *PlayerDataHandler) Delete(c fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return serviceError(c, service.ErrInvalidInput)
	}

	if err := h.data.Delete(c.Context(), accountID(c), id); err != nil {
		return serviceError(c, err)
	}

	return noContent(c)
}

func parseUintParam(c fiber.Ctx, name string) (uint64, error) {
	value := c.Params(name)
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, service.ErrInvalidInput
	}
	return id, nil
}
