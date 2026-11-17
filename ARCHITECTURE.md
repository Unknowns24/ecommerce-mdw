# ARCHITECTURE.md — arquitectura real del e-commerce de BC Importados

Resumen técnico para quien llega al repo. Las decisiones están en `docs/adr/` (ADR-001 a ADR-007); el comportamiento, en `docs/specs.md`; los permisos, en `docs/api.md`.

## 1. Vista general

```
Navegador ──► Next.js 16 (frontend/)            ──► API Go (backend/)                  ──► PostgreSQL 16
              Server Components, Server Actions      Chi · casos de uso · GORM             RLS: rol bc_app por request
              la cookie «sesion» (HttpOnly)          JWT HS256 propio                      dueño: migraciones, seeds, tareas
              nunca llama a la API desde el                 │
              navegador (ADR-007)                           └──► Mercado Pago (preferencias, webhook, conciliación)
```

Mismo dominio para los dos: `/api/*` va a la API y el resto a Next (Nginx en la VPS, Vercel Services en la copia personal). Así la cookie vale para los dos y no hay pedidos entre orígenes distintos.

## 2. Stack

| Capa | Tecnología | Dónde |
| --- | --- | --- |
| Frontend | Next.js 16.4 (App Router, Turbopack, React Compiler), React 19, TypeScript, Tailwind v4, shadcn/ui (Radix) | `frontend/` |
| Backend | Go 1.26, Chi, GORM + gormigrate, validator, JWT propio | `backend/` |
| Base | PostgreSQL 16 con Row Level Security | migraciones en `backend/internal/adapters/db/postgresql/migrations/` |
| Pagos | Mercado Pago (Checkout Pro y webhook con firma) | `backend/internal/adapters/external/mp/` |
| Pruebas | `go test` (dominio y handlers), pruebas de integración de RLS, Playwright (permisos, flujo) | `backend/**/*_test.go`, `frontend/e2e/` |
| CI | GitHub Actions: `Verificar backend`, `Verificar frontend`, `e2e` | `.github/workflows/cicd.yaml` |
| Despliegue | VPS: Docker Compose + Nginx + Watchtower (ADR-005). Copia personal: Vercel Services + Neon | `docker-compose.yml`, `deploy/`, `vercel.json` |

## 3. Backend: capas

```
cmd/api            arranque, subcomando migrate, middlewares globales
cmd/seed*          datos de demostración (como dueño)
internal/core/domain        entidades y estados
internal/core/ports/out     interfaces que los casos de uso necesitan
internal/usecases/<modulo>  reglas de negocio (probadas como funciones)
internal/adapters/
  rest/handlers/<modulo>    HTTP ↔ caso de uso; Montar(...) registra rutas y middlewares
  rest/middleware           sesión (JWT), permisos, Cache-Control
  rest/errors               contrato único de errores (400/401/403/404/409)
  db/postgresql             conexión, migraciones, repositorios, rls (contexto por request)
  external/mp               cliente de Mercado Pago (timeout de 10 s)
  jwt                       emisor y verificador HS256
```

Módulos: identidad, catálogo, inventario, pedidos (carrito, checkout, consulta, gestión) y pagos.

### Un request, de punta a punta

1. **Middlewares globales**:
   - RequestID, RealIP, Recoverer y Logger;
   - `SinCacheConSesion`: `private, no-store` si hay credenciales;
   - CORS;
   - **`rls.Pool.Contexto`**: abre una transacción, `SET ROLE bc_app`, `app.usuario_id` y `app.permisos` desde el JWT.
2. **Middlewares de la ruta**:
   - `RequiereSesion`, `RequierePermiso`, `SesionOpcional`;
   - `rls.Sistema` en las rutas que necesitan los privilegios del dueño.
3. **Handler → caso de uso → repositorio**: GORM manda las consultas a la transacción del request (`rls.Pool`), y las transacciones anidadas son savepoints.
4. **Fin**: `COMMIT` si la respuesta es 2xx o 3xx, `ROLLBACK` si es 4xx o 5xx. Recién después se envía la respuesta guardada en el buffer.

### Procesos fuera de un request

- `pagos.IniciarVencimiento`: cada minuto cancela los pedidos de Mercado Pago con la reserva vencida y concilia con Mercado Pago. Corre como dueño.
- Migraciones (`./api migrate`) y seeds: como dueño.

## 4. Frontend: estructura

```
src/app/(tienda)   catálogo /, /variantes/[id], /carrito, /checkout, /pago/retorno, /seguimiento/[token], /ingresar, /registro
src/app/(cuenta)   /cuenta, /cuenta/pedidos, /cuenta/pedidos/[id]          (requiere sesión)
src/app/(admin)    /admin, pedidos, productos, marcas, categorías, proveedores, stock, usuarios, roles, sistema
src/app/metricas   recepción de Web Vitals (route handler)
src/features/<modulo>   api.ts (llamadas a la API), tipos, componentes del dominio
src/lib            api.ts (única puerta a la API), acciones.ts (400/409 → formulario), formato.ts
src/components     ui/ (shadcn, sin negocio), Campo, Formulario, SinDatos, ErrorDeCarga, encabezado, barra fija
```

- **Datos**: solo desde el servidor, con `cache: "no-store"`. Timeout de 25 s en `api()`. Next memoiza los GET repetidos dentro de un mismo request.
- **Escrituras**: Server Actions con `useActionState`. Los ids salen de `.bind`, nunca de campos ocultos.
- **Estados**: `loading.tsx` (esqueletos), `error.tsx` y `global-error.tsx`, y `SinDatos` para vacío y sin resultados.
- **Diseño**: tokens en `globals.css`, la dirección de `DESIGN.md`, Bodoni Moda y Public Sans con `next/font`.

## 5. Datos

23 tablas: identidad (`usuario`, `rol`, `permiso`, `usuario_rol`, `rol_permiso`), catálogo (`marca`, `categoria`, `producto`, `producto_categoria`, `variante`, `imagen_variante`), inventario (`proveedor`, `lote`, `movimiento_stock`, `reserva_stock`), pedidos (`configuracion_tienda`, `tarifa_distancia`, `carrito`, `item_carrito`, `pedido`, `detalle_pedido`), pagos (`intento_pago`) y `migrations`.

- **Dinero**: en centavos (`bigint`).
- **Stock**: por lotes con FIFO/LIFO y reservas; la disponibilidad se calcula, no se guarda.
- **Pedido**: guarda una copia histórica (comprador, precios y entrega).
- **Seguridad**: ver `SECURITY_AUDIT.md`.

## 6. Despliegue

| Entorno | Cómo | Base |
| --- | --- | --- |
| Producción del equipo | VPS: `api`, `web` y `db` en Docker Compose; Nginx con TLS; Watchtower toma `:latest` de GHCR | Postgres en el compose |
| Copia personal | Vercel Services (`frontend` en Next y `backend` como contenedor desde su Dockerfile) | Neon (Marketplace) |
| Local | `make run` y `npm run dev`, o `docker compose up` | Postgres local |

**Antes de desplegar este código en cualquier entorno, hay que correr la migración `20261111_05_seguridad_rls`** (`./api migrate`), con el dueño de la base.

## 7. Deuda técnica y limitaciones conocidas

- El usuario de conexión de la API es dueño de las tablas: el siguiente paso de seguridad es separar un usuario de login para la app (`SECURITY_AUDIT.md`, sección 9).
- En Vercel, la tarea de vencimiento solo corre mientras la API está despierta (se apaga a los 5 minutos sin tráfico).
- El panel de producto hace un pedido de imágenes por variante (P2, decisión del 10/11).
- Las fotos de producto, la geocodificación, las notificaciones y la factura están fuera del alcance (spec, sección 9).
- 12 avisos moderados o bajos en dependencias de desarrollo (typescript-eslint, CLI de shadcn), que piden versiones major.
