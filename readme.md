# BC Importados — E-commerce

Proyecto de Metodologías y Desarrollo Web (UAI) para centralizar catálogo, stock por lotes, pedidos web y operación administrativa de BC Importados.

La fuente funcional de esta entrega es [la especificación del sistema](docs/specs.md). El documento distingue decisiones confirmadas, propuestas para revisión y pendientes: estos últimos no deben resolverse de forma implícita al implementar.

## Alcance de v1

- Catálogo público de variantes, búsqueda, filtros, carrito y favoritos.
- Compra como invitado o cliente registrado, retiro y envío por distancia vial.
- Mercado Pago, efectivo para retiro, reservas de stock, promociones y cupones.
- Catálogo, lotes, movimientos, ventas externas, pedidos, roles, auditoría y estadísticas online para administración.

Quedan fuera de alcance, entre otros, venta sin stock, múltiples sucursales, órdenes de compra a proveedores, reintegros automáticos, logística con repartidores y app nativa. El detalle está en la sección 9 de la spec.

## Arquitectura y decisiones

Las decisiones aceptadas se mantienen como ADR:

- [ADR-001: configuración centralizada con Viper](docs/adr/ADR-001-configuracion.md)
- [ADR-002: GORM y gormigrate para persistencia](docs/adr/ADR-002-persistencia-y-migraciones.md)
- [ADR-003: API HTTP con Chi](docs/adr/ADR-003-api-http.md)

Las propuestas futuras y asuntos abiertos están separados como [RFC](docs/rfc/README.md). Un RFC no habilita código productivo hasta contar con una decisión aceptada.

## Estructura

```text
backend/          API en Go
  config/         único punto de lectura de variables de entorno
  cmd/            puntos de entrada y comandos operativos
  internal/       casos de uso, dominio y adaptadores privados
docs/
  specs.md        especificación funcional
  adr/            decisiones de arquitectura aceptadas
  rfc/            propuestas y decisiones pendientes
frontend/         tienda y panel en Next.js
```

## Backend

El backend usa Go 1.26. La configuración se carga con `config.Load()` desde [`backend/config/viper.go`](backend/config/viper.go); ningún otro paquete debe leer variables de entorno directamente.

Para validar el módulo actual:

```bash
cd backend
go test ./...
```

Las variables, defaults, obligatoriedad y normas para secretos están documentadas en [ADR-001](docs/adr/ADR-001-configuracion.md). No se incluye un archivo `.env` ni valores de ejemplo con secretos.

## Frontend

```bash
cd frontend
npm install
npm run dev
```

Antes de modificar el frontend, revisar [`frontend/AGENTS.md`](frontend/AGENTS.md): el proyecto usa una versión de Next.js con convenciones específicas.
