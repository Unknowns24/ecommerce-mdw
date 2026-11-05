# Plan de trabajo — ecommerce-mdw

Equipo: Yasmín Ruffinengo, Genaro Civilotti, Nicolás Di Bernardo, Agustín Bressan
Responsable del repositorio: Genaro (`Unknowns24`)
Última actualización: 05/11/2026 — reemplaza el plan del 12/09, que quedó con el calendario corrido (el Parcial I fue el 06/10, no el 29/09). La versión anterior está en el historial de git.

---

## 1. Dónde estamos parados — auditoría al 05/11

| Hito de la cátedra                                        | Fecha     | Estado                                                        |
| --------------------------------------------------------- | --------- | ------------------------------------------------------------- |
| URL pública funcionando                                   | 11/08     | Sí: `https://mdw.unkcode.com`                                 |
| Backend completo: CRUD, validaciones, auth, servicio externo | 22/09  | Sí, con los defectos de la tabla de abajo                     |
| Parcial I — defensa del backend                           | 06/10     | Rendido                                                       |
| Clase 10 — UI, design system, cinco estados               | 13/10     | **No empezado**                                               |
| Clase 11 — harness: `AGENTS.md`, sensores; mobile-first   | 20/10     | **No empezado** (este PR agrega `AGENTS.md`)                  |
| Clase 12 — formularios y mutaciones · *hito: flujo navegable desde la UI* | 27/10 | **No empezado**: el frontend es el template de `create-next-app` |
| Clase 13 — testing y calidad, e2e en CI, `main` protegida | 03/11     | **No empezado**: el CI solo corre `go test`                   |
| TP N° 1 (individual, Ultra)                               | **mar 10/11, 23:00** | Cada uno                                           |
| **Parcial II — defensa de la aplicación**                 | **mar 17/11** | Quedan 12 días                                             |
| Demo Day / recuperatorios                                 | mar 24/11 | —                                                             |

La clase 14 (el sistema bajo carga, 10/11) no suma trabajo al proyecto: el equipo decidió no incorporarla al desarrollo.

**El último commit del repo es del 06/10.** Todo lo que piden las clases 10 a 13 está pendiente, y el Parcial II se defiende recorriendo el flujo principal desde la interfaz, con pruebas corriendo en CI.

### Lo que falta, por clase

| Clase | Entregable de la cátedra                                                                 | Estado en el repo                                                        |
| ----- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| 10    | `components/ui/` instalado y usado; ningún borde a mano                                  | No hay Tailwind ni shadcn/ui                                             |
| 10    | Cinco estados en la pantalla principal                                                    | No hay pantallas                                                         |
| 10    | Barra con nombre del sistema, quién sos, tu rol y cómo salir                             | No hay; además `GET /api/cuenta` no devuelve roles ni permisos           |
| 10    | Mapa de pantallas en la spec, sección 5                                                  | **Agregado en este PR**                                                  |
| 10    | `/sistema` con todos los componentes y estados                                           | No hay                                                                   |
| 10    | Público y privado separados, login en su URL, 404 que no cuenta de más                   | No hay                                                                   |
| 11    | `AGENTS.md` con reglas verificables; el agente corre los sensores antes de entregar     | **Agregado en este PR** (raíz)                                           |
| 11    | Mobile-first                                                                             | Criterios **agregados a la spec**, sección 7                             |
| 12    | Flujo principal recorrible desde la UI, sin Postman                                      | No                                                                       |
| 12    | Server Actions que verifican sesión; ids desde sesión o URL; dos errores en dos lugares; formulario que no se vacía; botón deshabilitado; `revalidatePath` | No hay formularios. Criterios **agregados a la spec** y a `AGENTS.md` |
| 12    | Recorrido en un teléfono real contra producción                                         | No                                                                       |
| 13    | E2E del flujo principal en verde                                                          | No hay Playwright                                                        |
| 13    | Una prueba de permisos por fila de la columna «Rol» del contrato                         | No había contrato con esa columna: **agregado `docs/api.md`**            |
| 13    | Pruebas e2e contra base local; sin modo prueba                                           | No                                                                       |
| 13    | CI con las dos capas en cada PR y `main` protegida con ellas                              | El CI corre build/vet/test/gofmt del backend; nada del frontend          |
| 13    | Sección «Pruebas» en `AGENTS.md` y checklist en el template de PR                        | **Agregados en este PR**                                                 |

### Defectos encontrados en lo que ya está hecho

Ordenados por gravedad. Los defectos 1 a 5 **se corrigieron el 05/11, con sus pruebas** (detalle en `docs/api.md`, sección «Diferencias»): quedan para que su dueño los revise y los pueda defender, porque en la defensa le van a preguntar a él por su módulo. Los 6 a 8 siguen abiertos.

| # | Defecto                                                                                                    | Dónde                                                     | Dueño    |
| - | ---------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- | -------- |
| 1 | **Las rutas de carrito y pedidos no están montadas.** `Montar` está vacío: el checkout, la consulta y la gestión de pedidos responden 404 en producción, aunque los handlers y sus pruebas existen | `backend/internal/adapters/rest/handlers/pedidos/rutas.go` | Nicolás · **corregido** |
| 2 | **Mercado Pago cobra sin el envío y el webhook no compara montos.** La preferencia lleva solo las líneas de productos y el pedido se confirma con lo que se haya acreditado | `handlers/pagos/iniciar.go`, `webhook.go`                 | Agustín · **corregido** |
| 3 | `POST /api/pedidos/{id}/pago` no verifica quién paga, ni el vencimiento de la reserva, ni si ya hay un intento activo (H14) | `handlers/pagos/iniciar.go`                               | Agustín · **corregido** |
| 4 | No hay vencimiento automático de los pedidos de Mercado Pago a los 15 minutos: una reserva abandonada bloquea stock para siempre (H14) | `handlers/pagos/vencimiento.go`                           | Agustín · **corregido** |
| 5 | `GET /api/cuenta` no devuelve roles ni permisos (la barra los necesita)                                     | `handlers/identidad/rutas.go`                             | Yasmín · **corregido** |
| 6 | Sin pruebas de dominio en catálogo (`usecases/catalogo` no tiene `_test.go`)                               | `backend/internal/usecases/catalogo/`                     | Genaro   |
| 7 | Documentación desactualizada: el README menciona `deploy/nginx/` (no existe) y un CD por SSH que el workflow ya no hace; ADR-005 describe ese mismo job | `readme.md`, `docs/adr/ADR-005-despliegue.md`             | Agustín  |
| 8 | La rama `docs/diagrams` (diagrama entidad-relación) no está mergeada                                       | rama `docs/diagrams`                                      | Genaro   |

### Una alerta sobre la nota individual

Los commits del repo se concentran en cuatro días (12/09, 18/09, 04/10 y 06/10). La nota de concepto mira la **distribución de los commits a lo largo del cuatrimestre, PRs propios y reviews hechas a otros**, y la cátedra lo dice explícitamente: «no todo la noche antes del parcial». De acá al 17/11: **un PR chico por persona por día hábil, con su propia cuenta, y una review a un compañero por día.**

---

## 2. Cómo nos dividimos

Se mantiene el reparto por **módulo de dominio, de punta a punta**: cada uno es dueño del backend y de las pantallas de su módulo. Lo que hace posible trabajar sin esperarse ya está escrito:

1. **Las direcciones de las pantallas están congeladas** en el mapa de pantallas (spec, sección 5). Nadie necesita preguntarle a otro a dónde enlazar.
2. **El contrato de la API está congelado** en `docs/api.md` y `docs/api/*.http`. Cada pantalla consume endpoints que ya existen.
3. **La forma de consumir la API está fijada** en ADR-007: un cliente compartido (`src/lib/api.ts`) y un `src/features/<modulo>/api.ts` por módulo.
4. **Las pruebas se escriben desde la spec y el contrato**, no desde el código del compañero: se pueden escribir en paralelo.

### Archivos compartidos del frontend y su regla

| Archivo / zona                                   | Dueño   | Regla para los demás                                                       |
| ------------------------------------------------ | ------- | -------------------------------------------------------------------------- |
| `src/components/ui/`, `src/app/globals.css`, `components.json` | Genaro | Solo lectura. Un componente base nuevo se pide por Issue                  |
| `src/lib/api.ts`, `src/lib/formato.ts`, `src/components/campo.tsx`, `src/components/sin-datos.tsx` | Genaro | Solo lectura |
| `src/app/layout.tsx`, `src/app/(tienda)/layout.tsx` | Genaro | Solo lectura                                                                |
| `src/app/(admin)/layout.tsx`, `src/app/(cuenta)/layout.tsx`, barra del panel | Yasmín | Solo lectura. Los módulos del panel salen del mapa de pantallas            |
| `src/features/<modulo>/` y las rutas del mapa de su módulo | el dueño del módulo | No se edita una ajena                                       |
| `frontend/package.json`                          | quien agrega la dependencia | Commit aparte, avisando en el grupo. Nunca en el commit de la feature |
| `playwright.config.ts`, `e2e/preparar.ts`, workflows de CI, `docker-compose.yml` | Agustín | Solo lectura                                                  |
| `e2e/sesiones.ts`, `e2e/permisos.spec.ts`        | Yasmín  | Un endpoint nuevo se agrega con su fila: quien crea el endpoint le pasa la fila |
| `e2e/flujo-principal.spec.ts`                    | Nicolás | —                                                                           |

### Git

- `main` protegida: solo se mergea desde `develop`, y con los checks del CI obligatorios (sección 4, viernes 13/11).
- Una rama por tarea: `feat/<modulo>/<descripcion>` o `fix/<modulo>/<descripcion>`. PR chico contra `develop`, con una review aprobada de alguien que no lo escribió.
- `git pull --rebase origin develop` antes de abrir el PR: el conflicto lo resuelve quien lo genera.
- Cada uno commitea con su propia cuenta.

---

## 3. Los cuatro paquetes

El viernes 06/11, mientras Genaro deja lista la base del frontend, los otros tres revisan el arreglo de backend de su módulo (defectos 1 a 5, ya corregidos): lo leen línea por línea, lo prueban contra producción y lo aprueban en el PR. «Ship it, but own it»: lo que el dueño no pueda explicar no cuenta en la defensa.

### Genaro — Catálogo y sistema visual

**Día 0 (vie 06/11), y es lo único que los demás esperan:** PR de la base del frontend, mergeado ese mismo día.
- Tailwind + `npx shadcn@latest init` y `add button card badge input label skeleton alert separator table select textarea`. Ese commit solo, sin pantallas.
- Tokens en `globals.css` con contraste 4,5:1.
- `src/lib/api.ts` (ADR-007), `src/lib/formato.ts` (centavos → pesos), `src/components/campo.tsx`, `src/components/sin-datos.tsx`.
- Los tres grupos de ruta con su layout mínimo, y el script `typecheck` en `package.json`.

**Después:**
- `(tienda)/layout.tsx`: encabezado con nombre, búsqueda, link al carrito e «Ingresar» o el nombre de quien tiene sesión.
- `/` catálogo: búsqueda, filtros por categoría y marca, orden por precio, paginación; los cinco estados, «SIN STOCK» visible (H04).
- `/variantes/[id]`: detalle, imágenes con alternativa si fallan, otras variantes, «Agregar al carrito» (el botón llama a la función que expone Nicolás en `features/carrito/`) (H05).
- `/admin/productos`, `/admin/productos/[id]` (variantes e imágenes), `/admin/marcas`, `/admin/categorias` (H08).
- `/admin/sistema`: la galería.
- Backend: pruebas de dominio de catálogo (defecto 6). Mergear `docs/diagrams` (defecto 8).
- Como responsable del repo: ruleset de `main` (viernes 13/11).

### Yasmín — Identidad, panel y pruebas de permisos

**Día 0:** revisar el defecto 5 (`GET /api/cuenta` con roles y permisos de la sesión, `dtos/identidad/cuenta.go`) y arrancar `/ingresar`.

**Después:**
- `/ingresar` y `/registro` con Server Actions; la cookie de sesión según ADR-007; «Salir».
- `/cuenta`: datos y edición (H02).
- `(cuenta)/layout.tsx` y `(admin)/layout.tsx`: redirección sin sesión y barra del panel con nombre, roles, los módulos que habilitan los permisos y «Salir».
- `/admin` (inicio del panel), `/admin/usuarios`, `/admin/roles` (H03).
- `e2e/sesiones.ts`: fabrica el JWT con el `APP_SECRET_KEY` de prueba para cada cuenta de prueba de `docs/api.md`.
- `e2e/permisos.spec.ts`: **una prueba por fila de `docs/api.md`**, escrita desde el contrato. Las filas de carrito y pedidos van a fallar hasta que se arregle el defecto 1: eso es lo esperado. Ver una fallar a propósito y contarlo en el PR.

### Nicolás — Carrito, checkout, consulta y flujo de punta a punta

**Día 0:** revisar el defecto 1 (las rutas de pedidos montadas en `handlers/pedidos/rutas.go`, con los dos adaptadores hacia catálogo e inventario). Después del deploy, correr `docs/api/pedidos.http` entero contra producción.

**Después:**
- `src/features/carrito/`: el carrito del invitado en `localStorage` y el del cliente contra la API, detrás de las mismas funciones (las usa el botón del detalle de Genaro).
- `/carrito` (H07).
- `/checkout`: datos del comprador, retiro o envío, medio de pago, resumen y confirmación; el 409 de stock enumera variante → disponibles arriba del formulario (H13, H15). Con Mercado Pago, inicia el pago y redirige al `init_point`.
- `/pago/retorno`: nunca muestra «pagado» por el parámetro; lleva al detalle del pedido (H14).
- `/seguimiento/[token]` con su propio 404, `/cuenta/pedidos`, `/cuenta/pedidos/[id]` (H16).
- `e2e/flujo-principal.spec.ts`: un `test.step` por fila del mapa, con el caso de error de stock en el medio (sábado 14/11, cuando todas las pantallas del flujo estén mergeadas).

### Agustín — Pagos, stock, back-office e infraestructura

**Día 0:** revisar los defectos 2, 3 y 4 (`usecases/pagos/reglas.go`, `handlers/pagos/iniciar.go`, `webhook.go` y `vencimiento.go`) y probar en el sandbox de Mercado Pago lo que acá no se pudo: una compra con envío de punta a punta, y que la búsqueda de pagos por `external_reference` responda con el token de producción. Después, arrancar la infraestructura.

**Después:**
- `/admin/pedidos` con filtros por fecha y estado, y `/admin/pedidos/[id]` con despacho, entrega, «Completar pedido» en efectivo y cancelación con motivo y confirmación explícita del reintegro (H17, H18, H15).
- `/admin/proveedores`, `/admin/stock/[varianteId]` (lotes, movimientos y ajuste), `/admin/stock/lotes/nuevo` (H09).
- Infraestructura: servicio `web` en `docker-compose.yml`, regla de Nginx `/api/` → API y el resto → Next (ADR-007), imagen del frontend en el workflow.
- CI: job del frontend (`lint`, `typecheck`, `build`) y job `e2e` (Postgres como servicio, migraciones, seed, API, `next build && next start`, Playwright), con `playwright.config.ts` que se niega a correr contra una base no local y `e2e/preparar.ts` que siembra los datos de la demo.
- Defecto 7: README y ADR-005 al día.

---

## 4. Calendario hasta el Parcial II

| Día            | Genaro                                         | Yasmín                                  | Nicolás                                   | Agustín                                            |
| -------------- | ---------------------------------------------- | --------------------------------------- | ----------------------------------------- | -------------------------------------------------- |
| **vie 06/11**  | **Base del frontend mergeada**                 | Revisa el defecto 5; `/ingresar`        | Revisa el defecto 1; `pedidos.http` en producción | Revisa 2, 3 y 4; prueba en el sandbox de MP |
| sáb 07 – dom 08 | Catálogo `/` y `/variantes/[id]`              | `/registro`, cookie de sesión           | `features/carrito`, `/carrito`            | Servicio `web` + Nginx en producción               |
| lun 09/11      | Detalle con «Agregar»                          | Layouts `(cuenta)` y `(admin)`, barra   | `/checkout` (retiro y efectivo)           | Job del frontend en el CI                          |
| **mar 10/11** — **TP N° 1** | Revisión del TP de cada uno; nadie mergea nada grande ese día | | | |
| mié 11/11      | `/admin/productos` y `[id]`                    | `/cuenta`, `/admin`, `/admin/usuarios`  | Checkout con envío y Mercado Pago, `/pago/retorno` | `/admin/pedidos` y `[id]` con transiciones  |
| jue 12/11      | Marcas, categorías, `/admin/sistema`           | `/admin/roles`, `e2e/sesiones.ts`       | `/seguimiento/[token]`, `/cuenta/pedidos` | Stock y proveedores; harness de Playwright         |
| **vie 13/11**  | **Ruleset de `main`**; pruebas de catálogo     | `e2e/permisos.spec.ts` completo         | Pantallas del flujo mergeadas             | **Jobs de frontend y e2e en el CI**                |
| sáb 14/11      | Recorrido con teclado y 375px de todo el panel | Visto fallar a propósito, en el PR      | `e2e/flujo-principal.spec.ts` en verde    | Recorrido en teléfono real del checkout            |
| dom 15/11      | **Congelamiento: desde acá solo arreglos.** Ensayo: cada uno explica el módulo de otro | | | |
| lun 16/11      | Datos de la demo cargados en producción; correr el flujo en producción una vez | | | |
| **mar 17/11**  | **PARCIAL II** — los cuatro, 10 minutos recorriendo el flujo desde la UI | | | |
| mar 24/11      | Demo Day: lo recortado en la sección 9 de la spec, si el flujo quedó completo y probado | | | |

Todas las semanas, además: el cuestionario del eje metodológico (el de la clase 13 vence el 10/11 a las 23:59) y el TP N° 1, que es individual y pide un **diagrama UML de despliegue**: conviene hacerlo una vez, en `docs/diagrams/`, con lo de ADR-005 y ADR-007 (Nginx, `web`, `api`, Postgres, Mercado Pago), y que cada uno lo explique en su informe.

### Rotación de code review

Cada uno revisa el PR de otro todos los días. Rotando así, en tres días todos revisaron a todos:

| Día        | Yasmín revisa a | Genaro revisa a | Nicolás revisa a | Agustín revisa a |
| ---------- | --------------- | --------------- | ---------------- | ---------------- |
| 1 · 4 · 7  | Genaro          | Nicolás         | Agustín          | Yasmín           |
| 2 · 5 · 8  | Nicolás         | Agustín         | Yasmín           | Genaro           |
| 3 · 6 · 9  | Agustín         | Yasmín          | Genaro           | Nicolás          |

---

## 5. Decisiones abiertas

| # | Decisión                                                                                     | Quién define | Para cuándo |
| - | -------------------------------------------------------------------------------------------- | ------------ | ----------- |
| 1 | Aprobar o corregir el recorte del Parcial II (spec, sección 9)                               | los cuatro   | vie 06/11   |
| 2 | Aceptar ADR-007 (cómo el frontend consume la API y el despliegue bajo el mismo dominio)       | los cuatro   | vie 06/11   |
| 3 | Qué pasa con el carrito del invitado al iniciar sesión: se descarta o se conserva sin mezclar | Nicolás      | lun 09/11   |
| 4 | Cómo prueba el e2e el paso de Mercado Pago (propuesta en ADR-007, «Pendiente»)               | Agustín      | jue 12/11   |
