package middleware

import (
	"strings"

	"github.com/drp-mikrest/backend/internal/repository"
	"github.com/drp-mikrest/backend/internal/util"
	"github.com/gofiber/fiber/v2"
)

func JWTAuth(jm *util.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized", "message": "token tidak ada"})
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized", "message": "format Authorization salah"})
		}
		claims, err := jm.Parse(parts[1])
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized", "message": err.Error()})
		}
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.SetUserContext(repository.WithAuthSource(c.UserContext(), "web"))
		return c.Next()
	}
}

// OptionalJWT mengisi context jika token ada, tapi tidak menolak jika tidak ada.
func OptionalJWT(jm *util.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Next()
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Next()
		}
		if claims, err := jm.Parse(parts[1]); err == nil {
			c.Locals("user_id", claims.UserID)
			c.Locals("email", claims.Email)
			c.Locals("role", claims.Role)
			c.SetUserContext(repository.WithAuthSource(c.UserContext(), "web"))
		}
		return c.Next()
	}
}

// UserID helper mengambil user_id dari context.
func UserID(c *fiber.Ctx) string {
	if v, ok := c.Locals("user_id").(string); ok {
		return v
	}
	return ""
}

// RequireRole memastikan role sesuai.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, _ := c.Locals("role").(string)
		if r != role {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden", "message": "akses ditolak"})
		}
		return c.Next()
	}
}
