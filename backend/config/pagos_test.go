package config

import "testing"

func TestProduccionExigeCredencialesDeMercadoPago(t *testing.T) {
	for _, falta := range []string{"MERCADOPAGO_ACCESS_TOKEN", "MERCADOPAGO_WEBHOOK_SECRET"} {
		t.Run(falta, func(t *testing.T) {
			t.Setenv("ENVIRONMENT", "production")
			t.Setenv("DATABASE_DSN", "postgres://localhost/pruebas")
			t.Setenv("APP_SECRET_KEY", "secreto-local")
			t.Setenv("ORDER_ACCESS_TOKEN_SECRET", "token-local")
			t.Setenv("PUBLIC_BASE_URL", "https://tienda.example")
			t.Setenv("MERCADOPAGO_ACCESS_TOKEN", "credencial-local")
			t.Setenv("MERCADOPAGO_WEBHOOK_SECRET", "firma-local")
			t.Setenv(falta, "   ")
			if _, err := Load(); err == nil {
				t.Fatalf("producción arrancó sin %s", falta)
			}
		})
	}
}
