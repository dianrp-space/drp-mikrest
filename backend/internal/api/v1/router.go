package v1

import (
	"github.com/drp-mikrest/backend/internal/handler"
	"github.com/drp-mikrest/backend/internal/middleware"
	"github.com/drp-mikrest/backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

// Deps berisi dependency untuk external API /api/v1 (Bearer token).
type Deps struct {
	TokenService   *service.TokenService
	VoucherService *service.VoucherService
	ProfileService *service.ProfileService
	ServerService  *service.ServerService
}

// Register memasang route /api/v1/* dengan autentikasi API token.
func Register(app *fiber.App, deps *Deps) {
	r := app.Group("/api/v1", middleware.APITokenAuth(deps.TokenService), middleware.TokenRateLimit())

	vh := handler.NewVoucherHandler(deps.VoucherService)
	sh := handler.NewServerHandler(deps.ServerService)
	ph := handler.NewProfileHandler(deps.ProfileService)

	// voucher generate (per user, via token)
	r.Post("/vouchers", middleware.RequireScope("vouchers:rw"), vh.Generate)
	r.Post("/vouchers/member", middleware.RequireScope("vouchers:rw"), vh.CreateMember)
	r.Get("/vouchers", vh.List)
	r.Get("/vouchers/:id", vh.Get)
	r.Post("/vouchers/:id/disable", middleware.RequireScope("vouchers:rw"), vh.Disable)
	r.Post("/vouchers/:id/enable", middleware.RequireScope("vouchers:rw"), vh.Enable)
	r.Delete("/vouchers/:id", middleware.RequireScope("vouchers:rw"), vh.Delete)

	// servers & profiles (read-only untuk API)
	r.Get("/servers", sh.List)
	r.Get("/servers/:id", sh.Get)
	r.Get("/servers/:id/profiles", ph.List)
}
