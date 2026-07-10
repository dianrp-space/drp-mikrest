package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/drp-mikrest/backend/internal/repository"
	"github.com/drp-mikrest/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// APITokenAuth middleware untuk /api/v1/* (Bearer token user).
func APITokenAuth(ts *service.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(fiber.Map{"error": "invalid_token", "message": "token tidak ada"})
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(401).JSON(fiber.Map{"error": "invalid_token", "message": "format salah"})
		}
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()
		user, tok, err := ts.AuthByToken(ctx, parts[1])
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid_token", "message": err.Error()})
		}
		c.Locals("user_id", user.ID.String())
		c.Locals("email", user.Email)
		c.Locals("role", user.Role)
		c.Locals("token_id", tok.ID.String())
		c.Locals("token_scopes", tok.Scopes)
		c.Locals("token_rate", tok.RateLimit)
		c.SetUserContext(repository.WithAuthSource(c.UserContext(), "api"))
		return c.Next()
	}
}

// RequireScope memastikan token memiliki scope tertentu.
func RequireScope(scope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scopes, _ := c.Locals("token_scopes").([]string)
		for _, s := range scopes {
			if s == scope || s == "*" {
				return c.Next()
			}
		}
		return c.Status(403).JSON(fiber.Map{"error": "insufficient_scope", "message": "token tidak punya scope: " + scope})
	}
}

// TokenUserID mengambil user_id dari context sebagai uuid.UUID.
func TokenUserID(c *fiber.Ctx) (uuid.UUID, error) {
	s := UserID(c)
	return uuid.Parse(s)
}
