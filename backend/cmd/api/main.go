package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DRP-MikREST/backend/internal/api/v1"
	"github.com/DRP-MikREST/backend/internal/config"
	"github.com/DRP-MikREST/backend/internal/db"
	"github.com/DRP-MikREST/backend/internal/handler"
	"github.com/DRP-MikREST/backend/internal/middleware"
	"github.com/DRP-MikREST/backend/internal/repository"
	"github.com/DRP-MikREST/backend/internal/scheduler"
	"github.com/DRP-MikREST/backend/internal/service"
	"github.com/DRP-MikREST/backend/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.With().Timestamp().Logger()

	seedEmail := flag.String("seed-email", "", "email admin awal jika users kosong")
	seedPass := flag.String("seed-pass", "", "password admin awal")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// Pastikan direktori uploads ada
	_ = os.MkdirAll("uploads", 0755)

	// DB pool
	pool, err := db.New(rootCtx, cfg.DB.DSN(), cfg.DB.MaxConns)
	if err != nil {
		log.Fatal().Err(err).Msg("konek db")
	}
	defer pool.Close()

	// Migrasi
	if err := db.Migrate(rootCtx, pool); err != nil {
		log.Fatal().Err(err).Msg("migrasi")
	}
	log.Info().Msg("migrasi selesai")

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	serverRepo := repository.NewServerRepository(pool)
	profileRepo := repository.NewProfileRepository(pool)
	voucherRepo := repository.NewVoucherRepository(pool)
	batchRepo := repository.NewBatchRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	auditRepo := repository.NewAuditRepository(pool)
	settingRepo := repository.NewSettingRepository(pool)

	if cfg.EncryptionKey == cfg.AppSecret {
		log.Warn().Msg("ENCRYPTION_KEY tidak di-set, menggunakan APP_SECRET (tidak aman). Set ENCRYPTION_KEY yang berbeda!")
	}

	// JWT manager
	jwtMgr := util.NewJWTManager(cfg.AppSecret, cfg.JWTAccessTTL)

	// Services
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	serverSvc := service.NewServerService(serverRepo, auditRepo, cfg.EncryptionKey)
	profileSvc := service.NewProfileService(profileRepo, serverSvc, auditRepo)
	voucherSvc := service.NewVoucherService(voucherRepo, batchRepo, profileRepo, serverSvc, auditRepo)
	hotspotSvc := service.NewHotspotService(serverSvc, auditRepo)
	tokenSvc := service.NewTokenService(tokenRepo, userRepo)
	expireSvc := service.NewExpirationService(voucherRepo, serverSvc, auditRepo)
	settingSvc := service.NewSettingService(settingRepo)
	auditSvc := service.NewAuditService(auditRepo)

	// Seed admin awal (opsional via flag / env)
	if *seedEmail != "" && *seedPass != "" {
		if err := authSvc.EnsureSeedUser(rootCtx, *seedEmail, *seedPass); err != nil {
			log.Error().Err(err).Msg("seed user")
		} else {
			log.Info().Str("email", *seedEmail).Msg("seed user dipastikan")
		}
	}

	// Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	app.Use(middleware.SecureHeaders())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigin,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Scheduler: auto-kick & auto-disable voucher expired
	sched := scheduler.New(expireSvc, settingSvc, auditSvc)
	sched.Start()
	defer sched.Stop(context.Background())

	// Web routes (JWT)
	// Static files for uploads
	app.Static("/uploads", "./uploads")

	handler.RegisterRoutes(app, &handler.Deps{
		JWT:                 jwtMgr,
		AuthHandler:         handler.NewAuthHandler(authSvc, tokenSvc, jwtMgr, auditRepo, *seedEmail, *seedPass),
		ServerHandler:       handler.NewServerHandler(serverSvc),
		VoucherHandler:      handler.NewVoucherHandler(voucherSvc),
		ProfileHandler:      handler.NewProfileHandler(profileSvc),
		HotspotHandler:      handler.NewHotspotHandler(hotspotSvc),
		TokenHandler:        handler.NewTokenHandler(tokenSvc, auditRepo),
		SettingHandler:      handler.NewSettingHandler(settingSvc, sched),
		AuditHandler:        handler.NewAuditHandler(auditSvc),
		DisableRegistration: cfg.DisableRegistration,
	})

	// External API /api/v1 (Bearer token per user)
	v1.Register(app, &v1.Deps{
		TokenService:   tokenSvc,
		VoucherService: voucherSvc,
		ProfileService: profileSvc,
		ServerService:  serverSvc,
	})

	// graceful shutdown
	go func() {
		log.Info().Str("port", cfg.AppPort).Msg("server mendengarkan")
		if err := app.Listen(":" + cfg.AppPort); err != nil {
			log.Fatal().Err(err).Msg("server berhenti")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("mematikan server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown")
	}
	serverSvc.CloseAll()
	log.Info().Msg("server berhenti")
}
