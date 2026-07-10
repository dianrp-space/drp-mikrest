package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/drp-mikrest/backend/internal/scheduler"
	"github.com/drp-mikrest/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SettingHandler struct {
	svc       *service.SettingService
	scheduler *scheduler.Scheduler
}

func NewSettingHandler(svc *service.SettingService, sched *scheduler.Scheduler) *SettingHandler {
	return &SettingHandler{svc: svc, scheduler: sched}
}

func (h *SettingHandler) GetAll(c *fiber.Ctx) error {
	settings, err := h.svc.GetAll(c.UserContext())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}
	return c.JSON(settings)
}

type updateSettingsReq struct {
	AppName  string `json:"app_name"`
	LogoPath string `json:"logo_path"`
}

func (h *SettingHandler) Update(c *fiber.Ctx) error {
	var in map[string]string
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	if err := h.svc.BatchSet(c.UserContext(), in); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *SettingHandler) UploadLogo(c *fiber.Ctx) error {
	return h.uploadFile(c, "logo_path", "logo")
}

func (h *SettingHandler) UploadFavicon(c *fiber.Ctx) error {
	return h.uploadFile(c, "favicon_path", "favicon")
}

func (h *SettingHandler) GetSchedulerInfo(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"cron_interval": h.scheduler.CurrentInterval(),
	})
}

func (h *SettingHandler) ReloadScheduler(c *fiber.Ctx) error {
	h.scheduler.Reload()
	return c.JSON(fiber.Map{"ok": true})
}

var allowedImageMIME = map[string]string{
	"image/png":          ".png",
	"image/jpeg":         ".jpg",
	"image/gif":          ".gif",
	"image/webp":         ".webp",
	"image/svg+xml":      ".svg",
	"image/x-icon":       ".ico",
	"image/vnd.microsoft": ".ico",
}

const maxUploadSize = 2 << 20 // 2 MB

func (h *SettingHandler) uploadFile(c *fiber.Ctx, settingKey, filePrefix string) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": "file tidak ditemukan"})
	}

	if file.Size > maxUploadSize {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": "file terlalu besar, maksimal 2 MB"})
	}

	fh, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": "gagal membaca file"})
	}
	defer fh.Close()

	buf := make([]byte, 512)
	if _, err := fh.Read(buf); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": "gagal membaca header file"})
	}

	mime := http.DetectContentType(buf)
	ext, ok := allowedImageMIME[mime]
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": "tipe file tidak diizinkan. Gunakan PNG, JPG, GIF, WebP, SVG, atau ICO"})
	}

	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": "gagal buat direktori upload"})
	}

	filename := filePrefix + "_" + uuid.New().String() + ext
	dest := filepath.Join(uploadDir, filename)

	// Rewind file before saving
	fh.Close()
	if err := c.SaveFile(file, dest); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": "gagal menyimpan file"})
	}

	// Hapus file lama jika ada
	settings, err := h.svc.GetAll(c.UserContext())
	if err == nil {
		if old := settings[settingKey]; old != "" {
			_ = os.Remove(old)
		}
	}

	if err := h.svc.Set(c.UserContext(), settingKey, dest); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}

	return c.JSON(fiber.Map{"ok": true, "path": dest})
}
