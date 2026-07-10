package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		lat := time.Since(start)
		ev := log.Info()
		if err != nil {
			ev = log.Error().Err(err)
		}
		ev.Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", lat).
			Str("ip", c.IP()).
			Msg("request")
		return err
	}
}

func Recoverer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Str("path", c.Path()).Msg("panic recovered")
				_ = c.Status(500).JSON(fiber.Map{"error": "internal", "message": "kesalahan server"})
			}
		}()
		return c.Next()
	}
}
