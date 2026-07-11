package handler

import (
	"strconv"

	"github.com/DRP-MikREST/backend/internal/middleware"
	"github.com/DRP-MikREST/backend/internal/repository"
	"github.com/DRP-MikREST/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type VoucherHandler struct {
	svc *service.VoucherService
}

func NewVoucherHandler(svc *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{svc: svc}
}

func (h *VoucherHandler) List(c *fiber.Ctx) error {
	f := repository.VoucherFilter{
		Limit:  parseInt(c.Query("limit"), 50),
		Offset: parseInt(c.Query("offset"), 0),
		Status: c.Query("status"),
		Search: c.Query("q"),
	}
	if sid := c.Query("server_id"); sid != "" {
		if id, err := uuid.Parse(sid); err == nil {
			f.ServerID = &id
		}
	}
	if bid := c.Query("batch_id"); bid != "" {
		if id, err := uuid.Parse(bid); err == nil {
			f.BatchID = &id
		}
	}
	if t := c.Query("type"); t == "voucher" || t == "member" {
		f.Type = t
	}
	list, total, err := h.svc.List(c.UserContext(), f)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"data": list, "total": total, "limit": f.Limit, "offset": f.Offset})
}

func (h *VoucherHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	v, err := h.svc.Get(c.UserContext(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "server_error", "message": err.Error()})
	}
	if v == nil {
		return c.Status(404).JSON(fiber.Map{"error": "not_found"})
	}
	return c.JSON(v)
}

func (h *VoucherHandler) Generate(c *fiber.Ctx) error {
	var in service.GenerateInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	if in.Count == 0 {
		in.Count = 1
	}
	if err := validateCount(in.Count); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation_failed", "errors": []fieldError{*err}})
	}
	uid, err := uuid.Parse(middleware.UserID(c))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	res, err := h.svc.Generate(c.UserContext(), in, uid)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "generate_failed", "message": err.Error()})
	}
	return c.Status(201).JSON(res)
}

func (h *VoucherHandler) Disable(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	uid, _ := uuid.Parse(middleware.UserID(c))
	if err := h.svc.Disable(c.UserContext(), id, uid); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "disable_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"disabled": true})
}

func (h *VoucherHandler) Enable(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	uid, _ := uuid.Parse(middleware.UserID(c))
	if err := h.svc.Enable(c.UserContext(), id, uid); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "enable_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{"enabled": true})
}

func (h *VoucherHandler) Delete(c *fiber.Ctx) error {
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

func (h *VoucherHandler) CreateMember(c *fiber.Ctx) error {
	var in service.MemberInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_request", "message": err.Error()})
	}
	var errs []fieldError
	if err := validateRequired(in.Username, "username"); err != nil {
		errs = append(errs, *err)
	}
	if err := validateRequired(in.Password, "password"); err != nil {
		errs = append(errs, *err)
	}
	if len(errs) > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "validation_failed", "errors": errs})
	}
	uid, err := uuid.Parse(middleware.UserID(c))
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	v, err := h.svc.CreateMember(c.UserContext(), in, uid)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "create_member_failed", "message": err.Error()})
	}
	return c.Status(201).JSON(v)
}

func (h *VoucherHandler) Sync(c *fiber.Ctx) error {
	sid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad_id"})
	}
	uid, _ := uuid.Parse(middleware.UserID(c))
	res, err := h.svc.SyncFromRouter(c.UserContext(), sid, &uid)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "sync_failed", "message": err.Error()})
	}
	return c.JSON(fiber.Map{
		"synced":   true,
		"updated":  res.Updated,
		"imported": res.Imported,
		"removed":  res.Removed,
	})
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
