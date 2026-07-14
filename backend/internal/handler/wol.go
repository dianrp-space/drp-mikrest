package handler

import (
	"github.com/DRP-MikREST/backend/internal/models"
	"github.com/DRP-MikREST/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WOLHandler struct {
	svc *service.WOLService
}

func NewWOLHandler(svc *service.WOLService) *WOLHandler {
	return &WOLHandler{svc: svc}
}

func (h *WOLHandler) List(c *fiber.Ctx) error {
	list, err := h.svc.ListAll(c.UserContext())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *WOLHandler) Create(c *fiber.Ctx) error {
	var in models.WOLTargetInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	if in.Name == "" || in.MACAddress == "" || in.InterfaceName == "" || in.ServerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "validation_failed", "message": "Semua field wajib diisi"})
	}
	t, err := h.svc.Create(c.UserContext(), in)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "create_failed", "message": err.Error()})
	}
	return c.Status(201).JSON(t)
}

func (h *WOLHandler) Send(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	if err := h.svc.SendWOL(c.UserContext(), id); err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "wol_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"sent": true})
}

func (h *WOLHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "delete_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"deleted": true})
}
