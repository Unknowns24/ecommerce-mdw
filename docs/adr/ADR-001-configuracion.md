# ADR-001 — Configuración centralizada con Viper

- Estado: aceptado
- Fecha: 2026-09-18

## Contexto

El backend requiere secretos, conexión a base de datos y parámetros operativos. Leer variables de entorno en handlers, casos de uso o adaptadores dispersaría el contrato de despliegue, dificultaría las pruebas y elevaría el riesgo de exponer secretos.

## Decisión

Toda variable de entorno del backend se declara, bindea y carga exclusivamente en [`backend/config/viper.go`](../../backend/config/viper.go). `config.Load()` crea una instancia independiente de Viper, bindea las etiquetas `mapstructure` de `Config`, aplica defaults no sensibles y devuelve una configuración validada o un error.

Ningún otro paquete debe llamar a `os.Getenv`, `os.LookupEnv` o Viper. La raíz de composición carga `Config` una sola vez y pasa dependencias ya construidas a las capas siguientes. Los secretos no se registran en logs ni se devuelven en errores HTTP.

## Contrato actual

| Variable | Requerida | Uso |
| --- | --- | --- |
| `ENVIRONMENT` | No (`development`) | Entorno de ejecución |
| `APP_NAME` | No (`bc-importados`) | Identificación de la aplicación |
| `APP_PORT` | No (`8080`) | Puerto HTTP |
| `APP_SECRET_KEY` | Sí | Secreto de aplicación/sesiones |
| `DATABASE_DSN` | Sí | DSN de la base de datos usada por GORM |
| `BOOTSTRAP_MASTER_KEY` | Al ejecutar bootstrap | Habilita el dueño inicial; no se expone |
| `ORDER_ACCESS_TOKEN_SECRET` | Sí | Firma de enlaces privados de pedidos invitados |
| `PAYMENT_RESERVATION_TTL` | No (`15m`) | Duración positiva de la reserva de Mercado Pago |
| `MERCADOPAGO_ACCESS_TOKEN` | Al habilitar Mercado Pago | Credencial de Mercado Pago |
| `MERCADOPAGO_WEBHOOK_SECRET` | Al habilitar webhook | Verificación de notificaciones |
| `FISCAL_API_BASE_URL`, `FISCAL_API_TOKEN` | Al integrar facturación | Adaptador fiscal pendiente de contrato |
| `OPENWA_BASE_URL`, `OPENWA_TOKEN` | Al integrar WhatsApp | Adaptador de confirmaciones |
| `MAIL_HOST`, `MAIL_PORT`, `MAIL_USERNAME`, `MAIL_PASSWORD`, `MAIL_FROM` | Al habilitar correo | Adaptador de correo |

Las integraciones pendientes no se activan por la mera presencia de estas variables: su validación específica debe ocurrir cuando se incorpore el adaptador y su contrato esté definido.

## Consecuencias

- Agregar una variable exige añadir su campo y etiqueta en `Config`, actualizar esta tabla y validar sus reglas en el mismo archivo.
- `PAYMENT_RESERVATION_TTL` se parsea explícitamente como `time.Duration`; valores vacíos, inválidos o no positivos fallan al iniciar.
- No se carga automáticamente un archivo `.env`. Las herramientas locales pueden cargarlo antes de iniciar el proceso, pero producción usa el entorno del proceso.
