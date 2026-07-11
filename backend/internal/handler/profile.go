package handler

import (
	"github.com/DRP-MikREST/backend/internal/middleware"
	"github.com/DRP-MikREST/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	svc *service.ProfileService
}

func NewProfileHandler(svc *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) List(c *fiber.Ctx) error {
	sid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	list, err := h.svc.List(c.UserContext(), sid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *ProfileHandler) Sync(c *fiber.Ctx) error {
	sid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	n, err := h.svc.SyncFromRouter(c.UserContext(), sid)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "sync_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"synced": true, "count": n})
}

func (h *ProfileHandler) Create(c *fiber.Ctx) error {
	sid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	var in service.ProfileInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	if in.SharedUsers == 0 {
		in.SharedUsers = 1
	}
	uid, _ := uuid.Parse(middleware.UserID(c))
	p, err := h.svc.Create(c.UserContext(), sid, in, uid)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "create_failed", "message": err.Error()})
	}
	return c.Status(201).JSON(p)
}

func (h *ProfileHandler) Delete(c *fiber.Ctx) error {
	sid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	pid, err := uuid.Parse(c.Params("profileId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_profile_id"})
	}
	uid, _ := uuid.Parse(middleware.UserID(c))
	if err := h.svc.Delete(c.UserContext(), sid, pid, uid); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "delete_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"deleted": true})
}
