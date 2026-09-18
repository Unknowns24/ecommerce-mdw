package config

import (
	"testing"
	"time"
)

func TestLoadReadsBoundEnvironmentVariables(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/bc")
	t.Setenv("APP_SECRET_KEY", "application-secret")
	t.Setenv("ORDER_ACCESS_TOKEN_SECRET", "order-secret")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("PAYMENT_RESERVATION_TTL", "20m")
	t.Setenv("MAIL_PORT", "2525")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppPort != "9090" || cfg.MailPort != 2525 {
		t.Fatalf("environment values were not decoded: AppPort=%q MailPort=%d", cfg.AppPort, cfg.MailPort)
	}
	if cfg.PaymentReservationTTL != 20*time.Minute {
		t.Fatalf("PaymentReservationTTL = %s, want 20m", cfg.PaymentReservationTTL)
	}
}

func TestLoadRejectsInvalidReservationTTL(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/bc")
	t.Setenv("APP_SECRET_KEY", "application-secret")
	t.Setenv("ORDER_ACCESS_TOKEN_SECRET", "order-secret")
	t.Setenv("PAYMENT_RESERVATION_TTL", "0")

	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded with a non-positive PAYMENT_RESERVATION_TTL")
	}
}

func TestLoadRejectsEmptyReservationTTL(t *testing.T) {
	t.Setenv("DATABASE_DSN", "******localhost:5432/bc")
	t.Setenv("APP_SECRET_KEY", "application-secret")
	t.Setenv("ORDER_ACCESS_TOKEN_SECRET", "order-secret")
	t.Setenv("PAYMENT_RESERVATION_TTL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded with an empty PAYMENT_RESERVATION_TTL")
	}
}

func TestLoadRejectsInvalidAppPort(t *testing.T) {
	t.Setenv("DATABASE_DSN", "******localhost:5432/bc")
	t.Setenv("APP_SECRET_KEY", "application-secret")
	t.Setenv("ORDER_ACCESS_TOKEN_SECRET", "order-secret")
	t.Setenv("APP_PORT", "70000")

	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded with invalid APP_PORT")
	}
}
