package handler

import (
	"context"
	"time"

	"github.com/DRP-MikREST/backend/internal/middleware"
	"github.com/DRP-MikREST/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ServerHandler struct {
	svc *service.ServerService
}

func NewServerHandler(svc *service.ServerService) *ServerHandler {
	return &ServerHandler{svc: svc}
}

func (h *ServerHandler) List(c *fiber.Ctx) error {
	list, err := h.svc.List(c.UserContext())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *ServerHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	s, err := h.svc.Get(c.UserContext(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}
	if s == nil {
		return c.Status(404).JSON(fiber.Map{"error": "not_found"})
	}
	return c.JSON(s)
}

func (h *ServerHandler) Create(c *fiber.Ctx) error {
	var in service.ServerInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	if in.APIPort == 0 {
		in.APIPort = 8728
	}
	if errs := validateServerInput(in); len(errs) > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "validation_failed", "errors": errs})
	}
	uid, err := uuid.Parse(middleware.UserID(c))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	s, err := h.svc.Create(c.UserContext(), in, uid)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "create_failed", "message": err.Error()})
	}
	return c.Status(201).JSON(s)
}

func validateServerInput(in service.ServerInput) []fieldError {
	var errs []fieldError
	if err := validateRequired(in.Name, "name"); err != nil {
		errs = append(errs, *err)
	}
	if err := validateRequired(in.Host, "host"); err != nil {
		errs = append(errs, *err)
	}
	if err := validateRequired(in.Username, "username"); err != nil {
		errs = append(errs, *err)
	}
	if err := validateRequired(in.Password, "password"); err != nil {
		errs = append(errs, *err)
	}
	if err := validatePort(in.APIPort); err != nil {
		errs = append(errs, *err)
	}
	return errs
}

func (h *ServerHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	var in service.ServerInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	if in.APIPort == 0 {
		in.APIPort = 8728
	}
	if errs := validateServerInputUpdate(in); len(errs) > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "validation_failed", "errors": errs})
	}
	uid, _ := uuid.Parse(middleware.UserID(c))
	s, err := h.svc.Update(c.UserContext(), id, in, uid)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "update_failed", "message": err.Error()})
	}
	return c.JSON(s)
}

func validateServerInputUpdate(in service.ServerInput) []fieldError {
	var errs []fieldError
	if err := validateRequired(in.Name, "name"); err != nil {
		errs = append(errs, *err)
	}
	if err := validateRequired(in.Host, "host"); err != nil {
		errs = append(errs, *err)
	}
	if err := validateRequired(in.Username, "username"); err != nil {
		errs = append(errs, *err)
	}
	if err := validatePort(in.APIPort); err != nil {
		errs = append(errs, *err)
	}
	return errs
}

var _ = time.Second

func (h *ServerHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	uid, _ := uuid.Parse(middleware.UserID(c))
	if err := h.svc.Delete(c.UserContext(), id, uid); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "delete_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"deleted": true})
}

func (h *ServerHandler) TestConnection(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 15*time.Second)
	defer cancel()
	name, err := h.svc.TestConnection(ctx, id)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "connection_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "online", "identity": name})
}

func (h *ServerHandler) Resource(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	cl, err := h.svc.GetClient(ctx, id)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "connection_failed", "message": err.Error()})
	}
	res, err := cl.SystemResource()
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "routeros_error", "message": err.Error()})
	}
	identity, _ := cl.SystemIdentity()
	return c.JSON(fiber.Map{"identity": identity, "resource": res})
}

func (h *ServerHandler) Interfaces(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	cl, err := h.svc.GetClient(ctx, id)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "connection_failed", "message": err.Error()})
	}
	ifaces, err := cl.Print("/interface")
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "routeros_error", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"data": ifaces})
}

var _ = time.Second
