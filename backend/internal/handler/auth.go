package handler

import (
	"context"
	"time"

	"github.com/drp-mikrest/backend/internal/middleware"
	"github.com/drp-mikrest/backend/internal/repository"
	"github.com/drp-mikrest/backend/internal/service"
	"github.com/drp-mikrest/backend/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type changeEmailReq struct {
	Password string `json:"password" validate:"required"`
	NewEmail string `json:"new_email" validate:"required,email"`
}

type changePasswordReq struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type AuthHandler struct {
	svc      *service.AuthService
	tokens   *service.TokenService
	jwt      *util.JWTManager
	audit    *repository.AuditRepository
	seedEmail string
	seedPass  string
}

func NewAuthHandler(svc *service.AuthService, tokens *service.TokenService, jwt *util.JWTManager, audit *repository.AuditRepository, seedEmail, seedPass string) *AuthHandler {
	return &AuthHandler{svc: svc, tokens: tokens, jwt: jwt, audit: audit, seedEmail: seedEmail, seedPass: seedPass}
}

type registerReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var in registerReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	var errs []fieldError
	if err := validateRequired(in.Email, "email"); err != nil {
		errs = append(errs, *err)
	} else if err := validateEmail(in.Email); err != nil {
		errs = append(errs, *err)
	}
	if err := validateMinLen(in.Password, 8, "password"); err != nil {
		errs = append(errs, *err)
	}
	if len(errs) > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "validation_failed", "errors": errs})
	}
	u, err := h.svc.Register(c.UserContext(), in.Email, in.Password, in.Role)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "register_failed", "message": err.Error()})
	}
	return c.Status(201).JSON(u)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var in loginReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	var errs []fieldError
	if err := validateRequired(in.Email, "email"); err != nil {
		errs = append(errs, *err)
	}
	if err := validateRequired(in.Password, "password"); err != nil {
		errs = append(errs, *err)
	}
	if len(errs) > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "validation_failed", "errors": errs})
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()
	pair, err := h.svc.Login(ctx, in.Email, in.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "login_failed", "message": err.Error()})
	}
	return c.JSON(pair)
}

type loginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	uidStr := middleware.UserID(c)
	if _, err := uuid.Parse(uidStr); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	return c.JSON(fiber.Map{
		"id":    uidStr,
		"email": c.Locals("email"),
		"role":  c.Locals("role"),
	})
}

func (h *AuthHandler) ChangeEmail(c *fiber.Ctx) error {
	uid, err := uuid.Parse(middleware.UserID(c))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var in changeEmailReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	if err := h.svc.ChangeEmail(c.UserContext(), uid, in.Password, in.NewEmail); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "change_email_failed", "message": err.Error()})
	}

	_ = h.audit.Log(c.UserContext(), &uid, nil, "auth.change_email", in.NewEmail, nil)

	role, _ := c.Locals("role").(string)
	token, expiresAt, err := h.jwt.Generate(uid.String(), in.NewEmail, role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token_generation_failed"})
	}

	return c.JSON(fiber.Map{
		"ok":           true,
		"access_token": token,
		"expires_at":   expiresAt.Format(time.RFC3339),
		"token_type":   "Bearer",
		"user": fiber.Map{
			"id":    uid.String(),
			"email": in.NewEmail,
			"role":  role,
		},
	})
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	uid, err := uuid.Parse(middleware.UserID(c))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var in changePasswordReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	if err := h.svc.ChangePassword(c.UserContext(), uid, in.OldPassword, in.NewPassword); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "change_password_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// EnsureSeed endpoint internal (dev) untuk memastikan admin awal ada.
func (h *AuthHandler) EnsureSeed(c *fiber.Ctx) error {
	if h.seedEmail == "" || h.seedPass == "" {
		return c.JSON(fiber.Map{"seeded": false, "message": "SEED_EMAIL/SEED_PASS tidak di-set"})
	}
	if err := h.svc.EnsureSeedUser(c.UserContext(), h.seedEmail, h.seedPass); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "seed_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"seeded": true, "email": h.seedEmail})
}
