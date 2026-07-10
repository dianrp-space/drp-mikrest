package handler

import (
	"github.com/drp-mikrest/backend/internal/middleware"
	"github.com/drp-mikrest/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type HotspotHandler struct {
	svc *service.HotspotService
}

func NewHotspotHandler(svc *service.HotspotService) *HotspotHandler {
	return &HotspotHandler{svc: svc}
}

func (h *HotspotHandler) ActiveUsers(c *fiber.Ctx) error {
	sid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	users, err := h.svc.ActiveUsers(c.UserContext(), sid)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "routeros_error", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"data": users})
}

func (h *HotspotHandler) Kick(c *fiber.Ctx) error {
	sid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	rosID := c.Params("rosId")
	if rosID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "bad_ros_id"})
	}
	uid, _ := uuid.Parse(middleware.UserID(c))
	if err := h.svc.Kick(c.UserContext(), sid, rosID, uid); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "kick_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"kicked": true})
}
