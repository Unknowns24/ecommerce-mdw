# BC Importados — E-commerce

Proyecto de Metodologías y Desarrollo Web (UAI) para centralizar catálogo, stock por lotes, pedidos web y operación administrativa de BC Importados.

La fuente funcional de esta entrega es [la especificación del sistema](docs/specs.md). El documento distingue decisiones confirmadas, propuestas para revisión y pendientes: estos últimos no deben resolverse de forma implícita al implementar.

## Alcance de v1

- Catálogo público de variantes, búsqueda, filtros, carrito y favoritos.
- Compra como invitado o cliente registrado, retiro y envío por distancia vial.
- Mercado Pago, efectivo para retiro, reservas de stock, promociones y cupones.
- Catálogo, lotes, movimientos, ventas externas, pedidos, roles, auditoría y estadísticas online para administración.

Quedan fuera de alcance, entre otros, venta sin stock, múltiples sucursales, órdenes de compra a proveedores, reintegros automáticos, logística con repartidores y app nativa. El detalle está en la sección 9 de la spec.

**Esta entrega (Parcial I, 06/10/2026)** recorta además promociones, cupones, facturación fiscal, estadísticas, favoritos, avisos de reposición, ventas externas, auditoría, WhatsApp, correo, geocodificación real y frontend — ver "Fuera de alcance de esta entrega" en [`docs/specs.md`](docs/specs.md#9-fuera-de-alcance). Esas historias siguen documentadas: vuelven para el Parcial II (17/11/2026).

## Arquitectura y decisiones

Las decisiones aceptadas se mantienen como ADR:

- [ADR-001: configuración centralizada con Viper](docs/adr/ADR-001-configuracion.md)
- [ADR-002: GORM y gormigrate para persistencia](docs/adr/ADR-002-persistencia-y-migraciones.md)
- [ADR-003: API HTTP con Chi](docs/adr/ADR-003-api-http.md)
- [ADR-004: dinero en centavos (`int64`) e identificadores `uuid.UUID`](docs/adr/ADR-004-tipos-de-datos.md)
- [ADR-005: despliegue en VPS propia, no en Vercel](docs/adr/ADR-005-despliegue.md)
- [ADR-006: login propio y JWT con permisos adentro](docs/adr/ADR-006-sesion.md)

Las propuestas futuras y asuntos abiertos están separados como [RFC](docs/rfc/README.md). Un RFC no habilita código productivo hasta contar con una decisión aceptada.

## Estructura

```text
backend/          API en Go
  config/         único punto de lectura de variables de entorno
  cmd/            puntos de entrada y comandos operativos (api, seed)
  internal/       casos de uso, dominio y adaptadores privados
  Dockerfile      build multi-stage, usuario no root
deploy/
  nginx/          referencia del reverse proxy con TLS para la VPS
docs/
  specs.md        especificación funcional
  adr/            decisiones de arquitectura aceptadas
  rfc/            propuestas y decisiones pendientes
  api/            requests .http de prueba, por módulo
docker-compose.yml  servicio api + Postgres 16 con volumen persistente
.github/workflows/  CI (build/vet/test/gofmt) y CD (deploy por SSH a main)
frontend/         tienda y panel en Next.js
```

## Producción

**URL pública:** https://mdw.unkcode.com

```bash
curl https://mdw.unkcode.com/api/salud
# {"estado":"ok","version":"<git-sha-corto>"}
```

## Backend

El backend usa Go 1.26. La configuración se carga con `config.Load()` desde [`backend/config/viper.go`](backend/config/viper.go); ningún otro paquete debe leer variables de entorno directamente. Las variables, defaults, obligatoriedad y normas para secretos están documentadas en [ADR-001](docs/adr/ADR-001-configuracion.md).

### Arrancar en local

```bash
cp backend/.env.example backend/.env   # completar DATABASE_DSN y los secretos
cd backend
make check     # go build + go vet + go test
make migrate   # aplica las migraciones de los cuatro módulos
make run       # go run ./cmd/api, escucha en :8080
```

Con Docker, además del backend hace falta levantar Postgres:

```bash
cp .env.example .env               # POSTGRES_USER / PASSWORD / DB del contenedor db
cp backend/.env.example backend/.env
docker compose up -d --build
docker compose run --rm api /api migrate
```

### Datos de prueba

```bash
cd backend
go run ./cmd/seed          # inventario: un proveedor y lotes con stock de sobra / una unidad / cero
# go run ./cmd/seed-identidad, ./cmd/seed-catalogo: los siembran Yasmín y Genaro
```

### Validar sólo el backend

```bash
cd backend
go test ./...
```

## Frontend

```bash
cd frontend
npm install
npm run dev
```

Antes de modificar el frontend, revisar [`frontend/AGENTS.md`](frontend/AGENTS.md): el proyecto usa una versión de Next.js con convenciones específicas.
