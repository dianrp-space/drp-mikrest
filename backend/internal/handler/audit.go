package handler

import (
	"time"

	"github.com/drp-mikrest/backend/internal/repository"
	"github.com/drp-mikrest/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

type auditListQuery struct {
	Action   string `json:"action"`
	Search   string `json:"search"`
	ServerID string `json:"server_id"`
	UserID   string `json:"user_id"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

func (h *AuditHandler) List(c *fiber.Ctx) error {
	var q auditListQuery
	if err := c.QueryParser(&q); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request"})
	}

	f := repository.AuditFilter{
		Action: q.Action,
		Search: q.Search,
		Limit:  q.Limit,
		Offset: q.Offset,
	}

	if q.ServerID != "" {
		id, err := uuid.Parse(q.ServerID)
		if err == nil {
			f.ServerID = &id
		}
	}
	if q.UserID != "" {
		id, err := uuid.Parse(q.UserID)
		if err == nil {
			f.UserID = &id
		}
	}
	if q.DateFrom != "" {
		t, err := time.Parse(time.RFC3339, q.DateFrom)
		if err == nil {
			f.DateFrom = &t
		}
	}
	if q.DateTo != "" {
		t, err := time.Parse(time.RFC3339, q.DateTo)
		if err == nil {
			f.DateTo = &t
		}
	}

	list, total, err := h.svc.List(c.UserContext(), f)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}

	return c.JSON(fiber.Map{"data": list, "total": total})
}
