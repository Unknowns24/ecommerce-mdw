// Package config centralizes the application's runtime configuration.
//
// Environment variables must not be read anywhere else in the backend. Add a
// field with its mapstructure tag here, document it in ADR-001, and use Load
// at the composition root to obtain the validated configuration.
package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config contains every environment-backed setting used by the backend.
// Secrets are intentionally kept as values only; they must never be logged.
type Config struct {
	Environment  string `mapstructure:"ENVIRONMENT"`
	AppName      string `mapstructure:"APP_NAME"`
	AppPort      string `mapstructure:"APP_PORT"`
	AppSecretKey string `mapstructure:"APP_SECRET_KEY"`

	DatabaseDSN string `mapstructure:"DATABASE_DSN"`

	BootstrapMasterKey     string        `mapstructure:"BOOTSTRAP_MASTER_KEY"`
	OrderAccessTokenSecret string        `mapstructure:"ORDER_ACCESS_TOKEN_SECRET"`
	PaymentReservationTTL  time.Duration `mapstructure:"-"`

	MercadoPagoAccessToken   string `mapstructure:"MERCADOPAGO_ACCESS_TOKEN"`
	MercadoPagoWebhookSecret string `mapstructure:"MERCADOPAGO_WEBHOOK_SECRET"`

	FiscalAPIBaseURL string `mapstructure:"FISCAL_API_BASE_URL"`
	FiscalAPIToken   string `mapstructure:"FISCAL_API_TOKEN"`
	OpenWABaseURL    string `mapstructure:"OPENWA_BASE_URL"`
	OpenWAToken      string `mapstructure:"OPENWA_TOKEN"`
	MailHost         string `mapstructure:"MAIL_HOST"`
	MailPort         int    `mapstructure:"MAIL_PORT"`
	MailUsername     string `mapstructure:"MAIL_USERNAME"`
	MailPassword     string `mapstructure:"MAIL_PASSWORD"`
	MailFrom         string `mapstructure:"MAIL_FROM"`
}

// bindFromStructTags makes the Config tags the single source of truth for the
// variable names Viper reads. Fields tagged "-" are deliberately excluded.
func bindFromStructTags[T any](v *viper.Viper) error {
	var zero T
	typ := reflect.TypeOf(zero)
	for i := 0; i < typ.NumField(); i++ {
		key := typ.Field(i).Tag.Get("mapstructure")
		if key == "" || key == "-" {
			continue
		}
		if err := v.BindEnv(key); err != nil {
			return fmt.Errorf("bind environment variable %s: %w", key, err)
		}
	}
	return nil
}

func parsePositiveDuration(name, raw string) (time.Duration, error) {
	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		if err != nil {
			return 0, fmt.Errorf("%s must be a positive Go duration: %w", name, err)
		}
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return value, nil
}

func parsePort(name, raw string) (string, error) {
	value := strings.TrimSpace(raw)
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		if err != nil {
			return "", fmt.Errorf("%s must be a valid TCP port between 1 and 65535: %w", name, err)
		}
		return "", fmt.Errorf("%s must be a valid TCP port between 1 and 65535", name)
	}
	return value, nil
}

// Load reads the environment through Viper and returns a validated config.
// It does not read dotenv files: loading those, if needed for local tooling,
// belongs outside the application process. Production configuration is always
// supplied by the process environment.
func Load() (Config, error) {
	v := viper.New()
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetDefault("ENVIRONMENT", "development")
	v.SetDefault("APP_NAME", "bc-importados")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("PAYMENT_RESERVATION_TTL", "15m")
	v.SetDefault("MAIL_PORT", 587)

	if err := bindFromStructTags[Config](v); err != nil {
		return Config{}, err
	}
	if err := v.BindEnv("PAYMENT_RESERVATION_TTL"); err != nil {
		return Config{}, fmt.Errorf("bind PAYMENT_RESERVATION_TTL: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}

	appPort, err := parsePort("APP_PORT", cfg.AppPort)
	if err != nil {
		return Config{}, err
	}
	cfg.AppPort = appPort

	ttl, err := parsePositiveDuration("PAYMENT_RESERVATION_TTL", v.GetString("PAYMENT_RESERVATION_TTL"))
	if err != nil {
		return Config{}, err
	}
	cfg.PaymentReservationTTL = ttl

	if strings.TrimSpace(cfg.DatabaseDSN) == "" {
		return Config{}, fmt.Errorf("DATABASE_DSN is required")
	}
	if strings.TrimSpace(cfg.AppSecretKey) == "" {
		return Config{}, fmt.Errorf("APP_SECRET_KEY is required")
	}
	if strings.TrimSpace(cfg.OrderAccessTokenSecret) == "" {
		return Config{}, fmt.Errorf("ORDER_ACCESS_TOKEN_SECRET is required")
	}

	return cfg, nil
}
