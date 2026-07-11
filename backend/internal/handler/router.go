package handler

import (
	"time"

	"github.com/DRP-MikREST/backend/internal/middleware"
	"github.com/DRP-MikREST/backend/internal/util"
	"github.com/gofiber/fiber/v2"
)

// Deps berisi semua handler + util untuk memasang route web.
type Deps struct {
	JWT                 *util.JWTManager
	AuthHandler         *AuthHandler
	ServerHandler       *ServerHandler
	VoucherHandler      *VoucherHandler
	ProfileHandler      *ProfileHandler
	HotspotHandler      *HotspotHandler
	TokenHandler        *TokenHandler
	SettingHandler      *SettingHandler
	AuditHandler        *AuditHandler
	DisableRegistration bool
}

// RegisterRoutes memasang semua route web (autentikasi JWT cookie/bearer).
func RegisterRoutes(app *fiber.App, deps *Deps) {
	api := app.Group("/api/web")

	// health check publik
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "time": time.Now()})
	})

	// public auth
	if !deps.DisableRegistration {
		api.Post("/auth/register", deps.AuthHandler.Register)
	}
	api.Post("/auth/login", middleware.LoginRateLimit(), deps.AuthHandler.Login)
	api.Post("/auth/seed", deps.AuthHandler.EnsureSeed)

	// public settings (dibutuhkan oleh login page tanpa token)
	api.Get("/settings", deps.SettingHandler.GetAll)

	// protected (JWT)
	protected := api.Use(middleware.JWTAuth(deps.JWT))
	protected.Get("/auth/me", deps.AuthHandler.Me)

	// servers
	protected.Get("/servers", deps.ServerHandler.List)
	protected.Post("/servers", deps.ServerHandler.Create)
	protected.Get("/servers/:id", deps.ServerHandler.Get)
	protected.Put("/servers/:id", deps.ServerHandler.Update)
	protected.Delete("/servers/:id", deps.ServerHandler.Delete)
	protected.Post("/servers/:id/test", deps.ServerHandler.TestConnection)
	protected.Get("/servers/:id/resource", deps.ServerHandler.Resource)

	// vouchers
	protected.Get("/vouchers", deps.VoucherHandler.List)
	protected.Post("/vouchers/generate", deps.VoucherHandler.Generate)
	protected.Post("/vouchers/member", deps.VoucherHandler.CreateMember)
	protected.Get("/vouchers/:id", deps.VoucherHandler.Get)
	protected.Post("/vouchers/:id/disable", deps.VoucherHandler.Disable)
	protected.Post("/vouchers/:id/enable", deps.VoucherHandler.Enable)
	protected.Delete("/vouchers/:id", deps.VoucherHandler.Delete)
	protected.Post("/servers/:id/vouchers/sync", deps.VoucherHandler.Sync)

	// hotspot profiles
	protected.Get("/servers/:id/profiles", deps.ProfileHandler.List)
	protected.Post("/servers/:id/profiles/sync", deps.ProfileHandler.Sync)
	protected.Post("/servers/:id/profiles", deps.ProfileHandler.Create)
	protected.Delete("/servers/:id/profiles/:profileId", deps.ProfileHandler.Delete)

	// hotspot active users
	protected.Get("/servers/:id/active", deps.HotspotHandler.ActiveUsers)
	protected.Post("/servers/:id/active/:rosId/kick", deps.HotspotHandler.Kick)

	// api tokens (kelola via web)
	protected.Get("/tokens", deps.TokenHandler.List)
	protected.Post("/tokens", deps.TokenHandler.Create)
	protected.Delete("/tokens/:id", deps.TokenHandler.Delete)

	// profile / change email & password
	protected.Put("/auth/email", deps.AuthHandler.ChangeEmail)
	protected.Put("/auth/password", deps.AuthHandler.ChangePassword)

	// audit logs
	protected.Get("/audit-logs", deps.AuditHandler.List)

	// settings
	protected.Put("/settings", deps.SettingHandler.Update)
	protected.Post("/settings/upload-logo", deps.SettingHandler.UploadLogo)
	protected.Post("/settings/upload-favicon", deps.SettingHandler.UploadFavicon)
	protected.Get("/settings/scheduler", deps.SettingHandler.GetSchedulerInfo)
	protected.Post("/settings/scheduler/reload", deps.SettingHandler.ReloadScheduler)
}
