package handler

import (
	"time"

	"github.com/DRP-MikREST/backend/internal/middleware"
	"github.com/DRP-MikREST/backend/internal/repository"
	"github.com/DRP-MikREST/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TokenHandler struct {
	svc   *service.TokenService
	audit *repository.AuditRepository
}

func NewTokenHandler(svc *service.TokenService, audit *repository.AuditRepository) *TokenHandler {
	return &TokenHandler{svc: svc, audit: audit}
}

type tokenCreateReq struct {
	Label     string     `json:"label" validate:"required"`
	Scopes    []string   `json:"scopes"`
	RateLimit int        `json:"rate_limit"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (h *TokenHandler) List(c *fiber.Ctx) error {
	uid, err := uuid.Parse(middleware.UserID(c))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	list, err := h.svc.List(c.UserContext(), uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *TokenHandler) Create(c *fiber.Ctx) error {
	uid, err := uuid.Parse(middleware.UserID(c))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var in tokenCreateReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	res, err := h.svc.Create(c.UserContext(), uid, service.TokenCreateInput{
		Label:     in.Label,
		Scopes:    in.Scopes,
		RateLimit: in.RateLimit,
		ExpiresAt: in.ExpiresAt,
	})
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "create_failed", "message": err.Error()})
	}
	_ = h.audit.Log(c.UserContext(), &uid, nil, "token.create", in.Label, nil)
	return c.Status(201).JSON(res)
}

func (h *TokenHandler) Delete(c *fiber.Ctx) error {
	uid, _ := uuid.Parse(middleware.UserID(c))
	tid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	if err := h.svc.Delete(c.UserContext(), tid, uid); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "delete_failed", "message": err.Error()})
	}
	_ = h.audit.Log(c.UserContext(), &uid, nil, "token.delete", tid.String(), nil)
	return c.JSON(fiber.Map{"deleted": true})
}
