# AGENTS.md — reglas de este proyecto

Lo lee el asistente de IA de cada integrante (Claude Code, Cursor, Copilot…) antes de escribir código. Si el equipo cambia una convención y este archivo no lo refleja, el asistente va a seguir escribiendo con la vieja. Cada regla es verificable: si no se puede chequear mirando un diff, no va acá.

`frontend/AGENTS.md` lo genera `next dev` y solo avisa que la versión de Next.js tiene cambios: vale además de este archivo, no en su lugar.

## Qué es este proyecto

E-commerce de BC Importados: catálogo público de variantes, carrito, compra como invitado o cliente registrado, pago con Mercado Pago o efectivo con retiro, stock por lotes con reservas, y un panel administrativo con permisos por rol.

Flujo principal: el comprador busca en el catálogo → arma el carrito → hace el checkout → el servidor crea el pedido y reserva stock → paga con Mercado Pago → el webhook confirma el pedido → el comprador lo consulta → la administración lo despacha y lo entrega.

## La especificación

Lo que el sistema tiene que hacer está en [`docs/specs.md`](docs/specs.md); quién puede llamar a cada endpoint, en [`docs/api.md`](docs/api.md).

- **Antes de escribir lógica de dominio, leé la sección 6 de la spec.** Las reglas no se deducen del código.
- **Lo marcado como «Pendiente» no se resuelve en silencio.** Si la tarea lo toca, preguntá.
- **Las pantallas y sus direcciones salen del mapa de pantallas** (spec, sección 5). No inventes rutas: si falta una fila, preguntá antes de crear la pantalla.
- **Lo que está en «Fuera de alcance» (sección 9) no se implementa**, aunque parezca fácil.

## Stack

| Parte    | Tecnología                                                                 | Decisión            |
| -------- | -------------------------------------------------------------------------- | ------------------- |
| Backend  | Go 1.26, Chi, GORM + gormigrate, PostgreSQL 16, JWT propio                  | ADR-001 a ADR-006   |
| Frontend | Next.js 16 (App Router), React 19, TypeScript, Tailwind + shadcn/ui        | ADR-007             |
| Pruebas  | `go test` (dominio), Playwright (permisos contra la API y flujo en navegador) | este archivo     |
| Deploy   | VPS propia con Docker Compose y Nginx                                       | ADR-005             |

## Comandos

```bash
# backend
cd backend
make check        # go build + go vet + go test
make migrate      # migraciones
make run          # API en :8080

# frontend
cd frontend
npm run dev       # desarrollo
npm run build     # build de producción
npm run lint
npm run typecheck # tsc --noEmit
npm run test:e2e  # Playwright: base local, API, app compilada y navegador
```

**Antes de entregar un cambio, corré los sensores de la parte que tocaste** (`make check` en backend; `lint`, `typecheck` y `build` en frontend) y decí el resultado. No entregues algo que no viste compilar.

## Dónde va cada cosa

| Si vas a escribir…                                   | Va en…                                                                 |
| ---------------------------------------------------- | ---------------------------------------------------------------------- |
| Una regla de negocio                                 | `backend/internal/usecases/<modulo>/`, con su `_test.go` al lado      |
| Una entidad o un tipo del dominio                    | `backend/internal/core/domain/<modulo>.go`                             |
| Una consulta a la base                               | `backend/internal/adapters/db/postgresql/repositories/<modulo>/`      |
| Un cambio de esquema                                 | Una migración nueva en `.../migrations/schema/`. Nunca se edita una ya mergeada |
| Un endpoint                                          | `backend/internal/adapters/rest/handlers/<modulo>/`, montado en su `Montar` |
| Una pantalla de la tienda (sesión opcional)          | `frontend/src/app/(tienda)/`                                          |
| Una pantalla de la cuenta del comprador              | `frontend/src/app/(cuenta)/`                                          |
| Una pantalla del panel                               | `frontend/src/app/(admin)/admin/`                                     |
| El estado de carga o de error de una pantalla        | `loading.tsx` / `error.tsx` al lado de su `page.tsx`                   |
| Una Server Action                                    | `acciones.ts` al lado de la `page.tsx` que la usa                      |
| Una llamada a la API de Go                           | `frontend/src/features/<modulo>/api.ts`                                |
| Un componente que conoce el dominio                  | `frontend/src/features/<modulo>/components/` (o `src/components/` si lo usan varios módulos) |
| Un componente base (botón, tarjeta, badge, input)    | `frontend/src/components/ui/`, traído con `npx shadcn@latest add`      |
| Una prueba de permisos                               | `frontend/e2e/permisos.spec.ts`, una por fila de `docs/api.md`        |
| La prueba del flujo principal                        | `frontend/e2e/flujo-principal.spec.ts`                                 |

## Reglas

### Backend

- **Un solo contrato de errores.** Solo `apierr.Responder` decide el status: 400 `ErrValidacion` (forma del dato, `detalles.campos`), 409 `ErrRegla` (regla de negocio que miró la base), 401, 403, 404.
- **La identidad sale de `middleware.UsuarioDeContexto`.** Ningún DTO, header propio ni query string trae `usuarioID`, rol ni permiso.
- **Un recurso ajeno responde 404, no 403.** Toda consulta de datos de un comprador lleva su id en el `WHERE`.
- **Toda ruta lleva su middleware** (`RequiereSesion`, `RequierePermiso`, `SesionOpcional`) o un comentario que diga que es pública a propósito.
- **Una regla de negocio no vive en un handler.** El handler traduce HTTP ↔ caso de uso; la regla está en `usecases/`.
- **El dinero es `int64` en centavos** (ADR-004). Nada de `float64` salvo para leer la respuesta de Mercado Pago y convertirla.
- **Toda lista tiene límite** (`Limit`) y selecciona columnas explícitas.
- Ningún paquete lee variables de entorno fuera de `config/` (ADR-001).
- Las operaciones que el proveedor puede repetir (webhook, cancelación, confirmación) son idempotentes: repetirlas no duplica stock, cobros ni pedidos.

### Frontend: datos y sesión (ADR-007)

- **La API de Go se llama solo desde el servidor de Next** —Server Components y Server Actions—, a través de `src/features/<modulo>/api.ts`. Ningún Client Component hace `fetch` a la API ni lee la cookie.
- La cookie de sesión `sesion` es `HttpOnly`: el navegador no la ve. El servidor de Next la reenvía a la API en cada llamada.
- **La validación de verdad es la del backend.** El frontend no duplica reglas con otro schema: muestra los `detalles.campos` del 400 debajo de cada campo. Los atributos HTML (`required`, `type="email"`) son ayuda, no validación.
- **Lo que el cliente no necesita elegir, el cliente no lo manda.** El id del usuario sale de la sesión; el id del recurso, de la URL (`.bind` en la acción). Nunca un `input type="hidden"` con un id, un rol o un precio.
- Los precios se muestran desde centavos con una sola función de formato (`src/lib/formato.ts`); ningún componente divide por 100 a mano.
- Ninguna variable con secretos lleva el prefijo `NEXT_PUBLIC_`.

### Frontend: pantallas

- Server Components por defecto. `"use client"` solo si hay estado, efectos o eventos del navegador (un buscador, un formulario con `useActionState`, `error.tsx`).
- **Una pantalla que lee datos resuelve sus cinco estados:** cargando (`loading.tsx`, esqueleto con la forma del contenido), con datos, vacío, sin resultados y error (`error.tsx`). Vacío y sin resultados dicen cosas distintas.
- **La pantalla de error no muestra el mensaje de la excepción:** el detalle va a `console.error`.
- **El 404 no cuenta de más:** «no existe» y «no es tuyo» muestran el mismo texto, sin la palabra «permiso».
- **Redirigir al login desde un layout no es autorización.** Cada página igual llama a la API con la sesión, y la API decide.
- Ningún estado se comunica solo con color: siempre hay texto.
- **Mobile-first:** primero el caso de 375px, después `sm:` hacia arriba. Botón principal a lo ancho en el teléfono, `<select>` nativo, `inputMode="numeric"` para DNI y teléfono, nada que solo aparezca con hover.

### Frontend: formularios

- **Server Actions con `useActionState`**, no `fetch` desde el cliente. El estado del formulario es lo que devolvió la acción: nada de un `useState` por campo.
- **La primera línea de cada acción verifica la sesión** cuando la operación la exige. Una Server Action es un endpoint público con otra sintaxis: `"use server"` no la hace privada.
- **Dos errores, dos lugares:** 400 debajo del campo; 409 arriba del formulario, con lo que la historia pide enumerar.
- **Un formulario que falla no se vacía:** la acción devuelve lo que se había escrito.
- **El botón se deshabilita mientras la acción corre** y dice qué está pasando.
- **Después de escribir: `revalidatePath` de cada pantalla que quedó vieja, y `redirect` al final** (lo que va después de `redirect` no se ejecuta). A dónde se redirige sale del mapa de pantallas.
- Todo campo usa el componente `Campo`: etiqueta con `htmlFor`/`id` (de `useId()`), error conectado con `aria-describedby` y `aria-invalid`.

### Estilos

- Solo Tailwind. Nada de archivos `.css` nuevos ni `style={{}}` salvo valores calculados en runtime. Los tokens (color, radio, tipografía) se cambian en `src/app/globals.css`.
- Ningún borde, radio ni sombra escrito a mano en `src/app/`: se usan los componentes de `components/ui/`.
- `components/ui/` no sabe nada del negocio. Si un componente habla de pedidos, variantes o stock, va afuera.

### Pruebas

- **Las pruebas cambian cuando cambia la spec o el contrato, no cuando cambia el código.** Un refactor no toca las pruebas; si las rompe, cambió un comportamiento. **Nunca se «arregla» una prueba roja cambiando el valor esperado.**
- **Las reglas de dominio se prueban como funciones** en `_test.go` al lado del caso de uso, sin base ni red; la fecha «de hoy» entra por parámetro. Una regla nueva o modificada de la sección 6 entra con su prueba en el mismo PR.
- **Cada fila de `docs/api.md` tiene su caso en `frontend/e2e/permisos.spec.ts`** (401 sin sesión, 403 sin permiso, 404 ajeno). Un endpoint nuevo entra con su fila y su prueba en el mismo PR. Se escriben mirando el contrato, no el handler.
- **El flujo principal tiene su prueba en `frontend/e2e/flujo-principal.spec.ts`**, un `test.step` por fila del mapa de pantallas y al menos un caso de error en el medio. Si cambia el flujo en la spec, cambia la prueba en el mismo PR.
- Las pruebas buscan como una persona: `getByRole`, `getByLabel`, `getByText`. Nunca por clase de CSS ni por un id generado.
- **Las pruebas e2e nunca corren contra la base de producción.** `playwright.config.ts` se niega a arrancar si la base no es local.
- **No hay «modo prueba» en la aplicación:** ni usuarios mágicos ni `if` que dependa de estar en una prueba. La sesión de las pruebas se fabrica afuera, firmando el JWT con el `APP_SECRET_KEY` de prueba (`frontend/e2e/sesiones.ts`).
- Una prueba nueva se ve fallar al menos una vez (sacando lo que prueba) antes de darla por buena.

### Git

- Ramas: `feat/<modulo>/<descripcion-corta>` o `fix/<modulo>/<descripcion-corta>`. PR contra `develop`, con una revisión aprobada de alguien que no lo escribió.
- Commits en español y en imperativo, con el formato que ya usa el repo: `feat(pedidos): monta las rutas de carrito y checkout`.
- **Cada uno commitea con su propia cuenta.** La nota es individual y se mide sobre el historial.
- No se editan archivos de un módulo ajeno (ver `.github/CODEOWNERS`): si hace falta un cambio ahí, se abre un Issue para su dueño.
- Las dependencias nuevas (`go.mod`, `package.json`) van en un commit aparte.
- Nunca se commitean `.env`, `backend/.env` ni credenciales.

## Cómo quiero que trabajes

- Si la consigna es ambigua, **preguntá antes de escribir código**. No inventes reglas de negocio ni rutas.
- Cambios chicos y enfocados. No refactorices lo que no tiene que ver con la tarea.
- Antes de crear un helper, buscá si ya existe.
- Cuando toques seguridad, dinero o el modelo de datos, explicá el porqué: son las partes que se revisan línea por línea.
