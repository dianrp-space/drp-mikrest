package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv     string
	AppPort    string
	AppSecret  string
	EncryptionKey string
	CORSOrigin string

	DB DBConfig

	JWTAccessTTL      time.Duration
	JWTRefreshTTL     time.Duration
	DisableRegistration bool
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

func Load() (*Config, error) {
	// muat file .env jika ada (tidak menimpa env yang sudah ada)
	if err := loadDotEnv(".env"); err != nil {
		// file .env opsional, abaikan jika tidak ada
		_ = err
	}

	disableReg := true
	if getEnv("DISABLE_REGISTRATION", "") == "false" {
		disableReg = false
	}

	c := &Config{
		AppEnv:     getEnv("APP_ENV", "development"),
		AppPort:    getEnv("APP_PORT", "8080"),
		AppSecret:  os.Getenv("APP_SECRET"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		CORSOrigin: getEnv("CORS_ORIGIN", "http://localhost:5173"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "drp_mikrest"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 10),
		},
		JWTAccessTTL:      getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:     getEnvDuration("JWT_REFRESH_TTL", 168*time.Hour),
		DisableRegistration: disableReg,
	}

	if c.AppSecret == "" {
		return nil, fmt.Errorf("APP_SECRET wajib di-set (min 32 char acak)")
	}
	if len(c.AppSecret) < 32 {
		return nil, fmt.Errorf("APP_SECRET terlalu pendek, gunakan min 32 char")
	}
	if c.EncryptionKey == "" {
		c.EncryptionKey = c.AppSecret
	}
	if c.DB.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD wajib di-set")
	}
	if c.AppEnv == "production" && strings.Contains(c.CORSOrigin, "*") {
		return nil, fmt.Errorf("CORS_ORIGIN tidak boleh mengandung '*' di production")
	}

	return c, nil
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		// hanya set jika belum ada di environment
		if _, ok := os.LookupEnv(key); !ok {
			_ = os.Setenv(key, val)
		}
	}
	return sc.Err()
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
