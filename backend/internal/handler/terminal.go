package handler

import (
	"github.com/DRP-MikREST/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TerminalHandler struct {
	serverSvc *service.ServerService
}

func NewTerminalHandler(serverSvc *service.ServerService) *TerminalHandler {
	return &TerminalHandler{serverSvc: serverSvc}
}

type terminalReq struct {
	Command string `json:"command"`
}

func (h *TerminalHandler) Exec(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	var in terminalReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	if in.Command == "" {
		return c.Status(400).JSON(fiber.Map{"error": "validation_failed", "message": "command wajib diisi"})
	}

	cl, err := h.serverSvc.GetClient(c.UserContext(), id)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "connection_failed", "message": err.Error()})
	}

	output, err := cl.Exec(in.Command)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "command_failed", "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"output": output,
	})
}
