# Plan de trabajo — ecommerce-mdw

Equipo: Yasmín Ruffinengo, Genaro Civilotti, Nicolás Di Bernardo, Agustín Bressan
Responsable del repositorio: Genaro (`Unknowns24`)
Última actualización: 12/09/2026

---

## 1. Dónde estamos parados

| Hito de la cátedra | Fecha | Estado |
|---|---|---|
| Repo creado y **URL pública funcionando** | 11/08 | Repo sí, URL pública **no** |
| Modelo de datos migrado en la base | 25/08 | **No** |
| API respondiendo en producción (clase 4) | 01/09 | **No** |
| Validación, errores y capas (clase 5) | 08/09 | **No** |
| **Backend completo**: CRUD, validaciones, auth con 2 roles, servicio externo | **22/09** | 10 días |
| **Parcial I — defensa del backend** | **29/09** | 17 días |
| Flujo principal navegable desde la UI | 20/10 | — |
| Testing, performance y seguridad revisados | 10/11 | — |
| **Parcial II — defensa de la aplicación** | **17/11** | — |
| **Demo Day** con datos reales | **24/11** | — |

Quedan **11 clases** (martes 15/09 → 24/11). El repo hoy tiene solo la estructura de carpetas.

**Condición dura del Parcial I:** si la URL pública no responde y el login no funciona, el parcial no se toma y pasa a recuperación del 24/11. Es lo primero que hay que resolver.

---

## 2. Cómo nos dividimos: rebanadas verticales, no capas

La cátedra desaconseja explícitamente repartir por rol fijo (arquitectura / datos / frontend / devops):

> "El cuatrimestre es backend hasta la clase 8 y frontend después, así que quien fuera 'el del frontend' se pasaría media cursada sin trabajo propio."

Por eso el reparto es por **módulo de dominio**: cada persona es dueña de su módulo de punta a punta — entidades, migraciones, casos de uso, endpoints, validaciones, pantallas y tests. Hasta el 29/09 eso significa backend; desde el 06/10, frontend del mismo módulo.

### Principios de desacople

1. **Un dueño por módulo, un módulo por carpeta.** Nadie edita archivos de otro módulo. Si necesitás un cambio ahí, abrís un Issue y lo asignás al dueño.
2. **Los contratos se congelan el 15/09 y no se discuten más.** Modelo de datos, contrato de API y puertos compartidos. Después de esa sesión nadie necesita preguntar nada para avanzar.
3. **Dependés de interfaces, no de personas.** Si tu módulo necesita algo de otro, se declara como interfaz en `backend/internal/core/ports/` y trabajás contra un stub en memoria hasta que el dueño lo implemente. El repo ya es hexagonal: esto sale gratis.

### Archivos con conflicto garantizado y su regla

| Archivo / zona | Regla |
|---|---|
| Migraciones | Un archivo por HU, prefijo `AAAAMMDDHHMM_<modulo>_`. **Una migración mergeada nunca se edita**, se corrige con otra nueva. |
| Registro de rutas | Cada módulo expone `RegisterRoutes(r)` en su propio archivo. Las 4 líneas del bootstrap se agregan **todas juntas el 15/09** y no se tocan más. |
| Seeds | `seed_<modulo>` — uno por persona. |
| `.env.example` | Sección por módulo, solo se agrega al final. Nunca se reordena. |
| `go.mod` / `package.json` | Dependencias nuevas en un **commit aparte**, avisando en el grupo. Nunca en el mismo commit que la feature. |
| `src/components/ui/` (kit) | Dueño: Genaro. Para el resto es **solo lectura**. |
| `docs/api.yaml` | Un archivo por módulo (`api-iam.yaml`, `api-catalogo.yaml`, …), nunca uno solo compartido. |

### Git

- `main` protegida — mergea solo Genaro. `develop` es la rama de integración.
- Una rama por HU: `feat/<modulo>/hu-XX-slug`. Ejemplo: `feat/catalogo/hu-08-busqueda`.
- PR chico, contra `develop`, con 1 review aprobada. `git pull --rebase origin develop` **antes** de abrir el PR: el conflicto lo resuelve quien lo genera.
- **Cada uno commitea con su propia cuenta.** La nota es individual y se mide sobre el historial del repo (commits, PRs propios y reviews, con fecha y autor). Nada de commits colectivos desde una máquina.

---

## 3. Los cuatro módulos

### Yasmín — Identidad y Acceso (IAM)

- **HU:** 01, 02, 03, 04, 05, 06, 33, 48, 49, 50, 51, 52 · más HU-53 (contrato de errores)
- **Entidades:** `User`, `Role`, `Permission`, `UserRole` (N-N), `RolePermission` (N-N)
- **Expone:** middleware `RequireAuth` y `RequirePermission(perm)`, usuario en contexto
- **Consume:** nada — **cero dependencias, arranca primera**
- **Carpetas:** `core/domain/{user,role,permission}`, `usecases/iam/`, `adapters/db/iam/`, `adapters/rest/iam/`, `adapters/jwt/` · front: `features/auth/`, rutas `login`, `registro`, `perfil`, `admin/usuarios`, `admin/roles`
- **Es el desbloqueante de los otros tres.** El middleware real tiene que estar el 22/09; hasta entonces los demás usan un stub.

### Genaro — Catálogo

- **HU:** 07, 08, 09, 10, 11, 12, 13, 34, 35, 36, 37, 38, 39, 40 · más HU-54 y HU-55 (kit de UI, estados de carga, responsive)
- **Entidades:** `Product`, `Category`, `ProductImage`
- **Expone:** `CatalogReader` — precio y disponibilidad de un producto (lo consume Checkout)
- **Consume:** middleware de Yasmín para los endpoints de admin
- **Carpetas:** `core/domain/{product,category,product_image}`, `usecases/catalogo/`, `adapters/db/catalogo/`, `adapters/rest/catalogo/` · front: `features/catalogo/`, rutas `/`, `productos/[id]`, `admin/productos`, `admin/categorias`, y `components/ui/`
- **Además, como responsable del repo:** protección de ramas, CODEOWNERS, plantilla de PR, convención de commits y los merges a `main`.

### Nicolás — Carrito, Checkout y Pedidos del cliente

- **HU:** 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 28, 29, 30
- **Entidades:** `Cart`, `CartItem`, `Order`, **`OrderItem`** (la N-N con datos propios que exige el núcleo obligatorio), `Address`, `ShippingMethod`
- **Expone:** `OrderReader` para el back-office
- **Consume:** `CatalogReader` (Genaro), `StockService` (Agustín), `PaymentService` (Agustín), auth (Yasmín) — **el módulo más acoplado por contrato**, por eso trabaja contra stubs hasta la semana del 22/09
- **Carpetas:** `core/domain/{cart,order,order_item,address,shipping_method}`, `usecases/checkout/`, `adapters/db/pedidos/`, `adapters/rest/pedidos/` · front: `features/carrito/`, rutas `carrito`, `checkout`, `mis-pedidos`

### Agustín — Pagos, Stock y Back-office

- **HU:** 24, 25, 26, 27, 31, 32, 41, 42, 43, 44, 45, 46, 47, 56
- **Entidades:** `Payment`, `StockMovement`, `Invoice` · dueño de la **máquina de estados de la orden**
- **Expone:** `StockService` (descontar / restaurar), `PaymentService` (crear preferencia), transiciones de estado
- **Consume:** `OrderReader` (Nicolás), auth (Yasmín)
- **Carpetas:** `core/domain/{payment,stock_movement,invoice}`, `usecases/pagos/`, `adapters/external/mercadopago/`, `adapters/rest/backoffice/` · front: `features/backoffice/`, rutas `admin/pedidos`, `admin/stock`, `pago/retorno`, `comprobante/[id]`
- **Además, infraestructura:** VPS, base de datos, `docker-compose`, CI/CD (el workflow está vacío), variables de entorno y la **URL pública** — hito vencido del 11/08, es lo primero.

### Cómo se cubre el núcleo obligatorio

| Requisito de la cátedra | Cómo lo cumplimos | Dueño |
|---|---|---|
| Auth con 2 roles que hagan cosas distintas | `cliente` y `admin` con RBAC por permisos | Yasmín |
| 5+ entidades propias, con 1-N y N-N | 1-N: `Category → Product`. N-N: `Order ↔ Product` vía `OrderItem` con cantidad y precio propios | Genaro / Nicolás |
| CRUD completo sobre al menos una | `Product` | Genaro |
| Flujo de negocio de punta a punta | catálogo → carrito → checkout → pago → pedido → estado | los cuatro |
| Integración externa | Mercado Pago (checkout + webhook) | Agustín |
| En producción, URL fija | VPS propia | Agustín |

---

## 4. Calendario

### Sprint 0 — sábado 13/09 al lunes 15/09

Única sesión sincrónica obligatoria del cuatrimestre: 3 horas los cuatro juntos. Se sale con los contratos congelados y **4 PRs, uno por persona** (para que quede evidencia individual en el historial).

| Quién | Entrega |
|---|---|
| Los cuatro | Diagrama del modelo de datos completo, en pizarra. Después cada uno commitea las entidades **de su módulo**. |
| Yasmín | `docs/spec.md` con el recorte del MVP (qué HU entran y cuáles van a *Fuera de alcance*) + contrato de errores de la API |
| Genaro | Protección de `main`, rama `develop`, `CODEOWNERS`, plantilla de PR, convención de commits |
| Nicolás | `docs/api-*.yaml` (contrato REST por módulo) + interfaces en `core/ports/` + stubs en memoria |
| Agustín | ADRs 0001 (Go en vez de Route Handlers), 0002 (Postgres), 0003 (VPS en vez de Vercel) + `docker-compose` + CI |

### Semana a semana

| Martes | Clase | Yasmín | Genaro | Nicolás | Agustín |
|---|---|---|---|---|---|
| **15/09** | C6 · Auth y autorización | Migraciones IAM, registro, login, logout, hash, JWT, middleware real, seed de roles | Migraciones catálogo, CRUD de productos y categorías | Migraciones `order`/`order_item`, carrito server-side, validaciones | **URL pública + CI + base en la VPS**, migraciones de stock y pagos |
| **22/09** | C7 · Servicios externos — *hito: backend completo* | Perfil, admin de usuarios, bloqueo, roles y permisos, tests de autorización | Listado con búsqueda, filtro, orden y paginación; imágenes; deshabilitar | Checkout completo: dirección, envío, creación de orden, revalidación de stock, mis pedidos | **Mercado Pago sandbox + webhook idempotente**, descuento y restauración de stock, back-office de pedidos |
| **29/09** | **C8 · PARCIAL I** | Ensayo el viernes 26: cada uno explica el módulo **de otro** | | | |
| **06/10** | C9 · React en Next | Rutas y fetching de auth y perfil | Rutas de catálogo y detalle | Rutas de carrito y checkout | Rutas de back-office y retorno de pago |
| **13/10** | C10 · UI y design system | Migra sus pantallas al kit | **Entrega el kit de UI** + responsive | Migra sus pantallas al kit | Migra sus pantallas al kit |
| **20/10** | C11 · Formularios y mobile-first — *hito: flujo navegable* | Formularios de registro y perfil | Filtros y buscador en mobile | **Lidera el flujo E2E**: catálogo → carrito → checkout → pago | Formularios de admin, ajuste de stock |
| **27/10** | C12 · Testing y calidad | Tests de autorización (los que no se ven en la UI) | Tests de catálogo y filtros | **E2E del flujo principal** | Tests del webhook y de la máquina de estados |
| **03/11** | C13 · IA en el producto | — | Dueño de la funcionalidad de IA *(a definir)* | — | Integración y costos |
| **10/11** | C14 · El sistema bajo carga — *hito: perf y seguridad* | Revisión de autorización en todos los caminos | Revisión de N+1 en listados | Revisión de la transacción de checkout | Índices, logs, rate limit |
| **17/11** | **C15 · PARCIAL II** | 10 minutos recorriendo el flujo principal desde la UI, los cuatro | | | |
| **24/11** | **C16 · DEMO DAY** | Exponen los cuatro. Retrospectiva: lo planificado vs. lo real | | | |

Todas las semanas, además: **PR subido el martes** y el cuestionario del eje metodológico entregado en UAI Online Ultra **antes** de la clase (es nota individual).

### Rotación de code review

La cátedra pide que en los últimos 15 minutos de cada clase alguien revise el PR de otro. Rotando así, en tres semanas todos revisaron a todos:

| Semana | Yasmín revisa a | Genaro revisa a | Nicolás revisa a | Agustín revisa a |
|---|---|---|---|---|
| 1 · 4 · 7 · 10 | Genaro | Nicolás | Agustín | Yasmín |
| 2 · 5 · 8 · 11 | Nicolás | Agustín | Yasmín | Genaro |
| 3 · 6 · 9 | Agustín | Yasmín | Genaro | Nicolás |

Esto también cubre el requisito de la defensa: se le pregunta a cualquiera sobre cualquier parte del sistema, y *"eso lo hizo otro"* no es una respuesta válida.

---

## 5. Decisiones abiertas

| # | Decisión | Quién define | Para cuándo |
|---|---|---|---|
| 1 | Recorte del MVP: ¿entran las 56 HU o se recortan? Propuesta abajo. | los cuatro | 15/09 |
| 2 | Qué funcionalidad de IA lleva el producto (clase 13). Propuesta: búsqueda semántica del catálogo, o generación de descripciones de producto. | Genaro | 06/10 |
| 3 | Preview deployments por PR en la VPS. La rúbrica lo puntúa como *destacado* en Producción; en Vercel es gratis, en VPS hay que armarlo. | Agustín | 22/09 |
| 4 | ADR que declare los equivalentes del stack sugerido: migraciones versionadas en vez de Prisma, validación en servidor en vez de Zod, sesiones en vez de Auth.js. La rúbrica se lee en el equivalente, pero **exige el ADR**. | Agustín | 15/09 |
| 5 | ¿Conseguimos un Product Owner real? La cátedra lo valora fuerte para el Demo Day. | los cuatro | 22/09 |

### Recorte del MVP propuesto

Fuera del alcance de los parciales, a `docs/spec.md` → *Fuera de alcance*:

- **HU-04** recuperación de contraseña — no está en el núcleo obligatorio.
- **HU-50, 51, 52** ABM de roles y permisos por pantalla — el RBAC va seedeado; el núcleo pide 2 roles funcionando, no un ABM.
- **HU-56** factura fiscal — queda solo el comprobante de HU-32.
- **HU-21** métodos de entrega — un envío fijo y retiro en local, sin cotizador.
- **HU-28**, criterio *"los pedidos se pueden volver a repetir"*.
- **HU-12 / HU-37** galería de imágenes — una principal, y adicionales si sobra tiempo.

Con eso quedan **~46 HU para cuatro personas en once semanas**, que es lo que entra.
