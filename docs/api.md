# Contrato de la API — quién puede qué

Una fila por endpoint. Es la fuente de las pruebas de permisos (`e2e/permisos.spec.ts`, ver [`AGENTS.md`](../AGENTS.md), sección «Pruebas»): cada fila se prueba **dada vuelta** —quién no puede y qué tiene que contestar el backend—, mirando esta tabla y no el código. El código dice lo que hace; este archivo dice lo que debería hacer.

Los cuerpos, ejemplos y casos de cada endpoint siguen en `docs/api/<modulo>.http`. Las reglas que deciden quién puede qué están en la sección 2 y 6 de [`docs/specs.md`](specs.md).

## Cómo se lee

| Columna           | Qué dice                                                                                                     |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **Quién puede**   | `público` (sin sesión), `sesión` (cualquier cuenta), un permiso (`catalogo.escribir`…) o `dueño`             |
| **Sin sesión**    | Lo que responde si no hay sesión. `—` si el endpoint es público                                              |
| **Sin permiso**   | Lo que responde con una sesión válida que no tiene el permiso o el rol. `—` si no pide permiso               |
| **Ajeno**         | Lo que responde si el recurso existe pero es de otra persona. `—` si no hay dueño del recurso                 |

Las respuestas siguen el contrato de errores de `backend/internal/adapters/rest/errors`: 401 «no sé quién sos», 403 «sé quién sos y no podés», 404 «para vos, esto no existe». Un recurso ajeno responde **404, no 403**: un 403 confirmaría que existe.

**Cuentas de prueba** que necesitan las pruebas de permisos: anónimo; cliente registrado (rol `cliente`, sin permisos); un administrativo con un solo permiso (para probar el 403 de los demás); dueño (todos los permisos).

## Salud

| Método | Ruta          | Quién puede | Sin sesión | Sin permiso | Ajeno | Historia |
| ------ | ------------- | ----------- | ---------- | ----------- | ----- | -------- |
| GET    | `/api/salud`  | público     | —          | —           | —     | —        |

## Identidad (`docs/api/identidad.http`)

| Método | Ruta                               | Quién puede                    | Sin sesión | Sin permiso | Ajeno | Historia |
| ------ | ---------------------------------- | ------------------------------ | ---------- | ----------- | ----- | -------- |
| POST   | `/api/auth/bootstrap`              | público con clave bootstrap    | —          | —           | —     | H01      |
| POST   | `/api/auth/registro`               | público                        | —          | —           | —     | H02      |
| POST   | `/api/auth/login`                  | público                        | —          | —           | —     | H02      |
| POST   | `/api/auth/logout`                 | sesión                         | 401        | —           | —     | H02      |
| GET    | `/api/cuenta`                      | sesión                         | 401        | —           | —     | H02      |
| PATCH  | `/api/cuenta`                      | sesión                         | 401        | —           | —     | H02      |
| GET    | `/api/admin/usuarios`              | dueño con `usuarios.gestionar` | 401        | 403         | —     | H03      |
| POST   | `/api/admin/usuarios`              | dueño con `usuarios.gestionar` | 401        | 403         | —     | H03      |
| PATCH  | `/api/admin/usuarios/{id}/estado`  | dueño con `usuarios.gestionar` | 401        | 403         | —     | H03      |
| PUT    | `/api/admin/usuarios/{id}/roles`   | dueño con `usuarios.gestionar` | 401        | 403         | —     | H03      |
| GET    | `/api/admin/roles`                 | dueño con `usuarios.gestionar` | 401        | 403         | —     | H03      |
| POST   | `/api/admin/roles`                 | dueño con `usuarios.gestionar` | 401        | 403         | —     | H03      |

Reglas que también se prueban, aunque no son de permisos: el bootstrap con un dueño ya existente se rechaza; el registro público nunca crea una cuenta con permisos administrativos; nadie puede desactivar ni quitarle el rol dueño al dueño inicial (sección 2 de la spec).

## Catálogo (`docs/api/catalogo.http`)

| Método              | Ruta                                          | Quién puede         | Sin sesión | Sin permiso | Ajeno | Historia |
| ------------------- | --------------------------------------------- | ------------------- | ---------- | ----------- | ----- | -------- |
| GET                 | `/api/catalogo/variantes`                     | público             | —          | —           | —     | H04      |
| GET                 | `/api/catalogo/variantes/{id}`                | público             | —          | —           | —     | H05      |
| GET                 | `/api/catalogo/categorias`                    | público             | —          | —           | —     | H04      |
| GET                 | `/api/catalogo/marcas`                        | público             | —          | —           | —     | H04      |
| GET, POST           | `/api/admin/productos`                        | `catalogo.escribir` | 401        | 403         | —     | H08      |
| GET, PATCH, DELETE  | `/api/admin/productos/{id}`                   | `catalogo.escribir` | 401        | 403         | —     | H08      |
| GET, POST           | `/api/admin/variantes`                        | `catalogo.escribir` | 401        | 403         | —     | H08      |
| GET, PATCH, DELETE  | `/api/admin/variantes/{id}`                   | `catalogo.escribir` | 401        | 403         | —     | H08      |
| POST                | `/api/admin/variantes/{id}/imagenes`          | `catalogo.escribir` | 401        | 403         | —     | H08      |
| PUT                 | `/api/admin/variantes/{id}/imagen-principal`  | `catalogo.escribir` | 401        | 403         | —     | H08      |
| DELETE              | `/api/admin/imagenes/{id}`                    | `catalogo.escribir` | 401        | 403         | —     | H08      |
| GET, POST           | `/api/admin/marcas`                           | `catalogo.escribir` | 401        | 403         | —     | H08      |
| GET, PATCH, DELETE  | `/api/admin/marcas/{id}`                      | `catalogo.escribir` | 401        | 403         | —     | H08      |
| GET, POST           | `/api/admin/categorias`                       | `catalogo.escribir` | 401        | 403         | —     | H08      |
| GET, PATCH, DELETE  | `/api/admin/categorias/{id}`                  | `catalogo.escribir` | 401        | 403         | —     | H08      |

Una variante inactiva no aparece en `/api/catalogo/variantes` y `/api/catalogo/variantes/{id}` responde 404 para ella (H04, H05).

## Inventario (`docs/api/inventario.http`)

| Método    | Ruta                                     | Quién puede       | Sin sesión | Sin permiso | Ajeno | Historia |
| --------- | ---------------------------------------- | ----------------- | ---------- | ----------- | ----- | -------- |
| GET, POST | `/api/admin/proveedores`                 | `stock.gestionar` | 401        | 403         | —     | H08      |
| POST      | `/api/admin/lotes`                       | `stock.gestionar` | 401        | 403         | —     | H09      |
| GET       | `/api/admin/variantes/{id}/stock`        | `stock.gestionar` | 401        | 403         | —     | H09      |
| GET       | `/api/admin/variantes/{id}/movimientos`  | `stock.gestionar` | 401        | 403         | —     | H09      |
| POST      | `/api/admin/variantes/{id}/ajustes`      | `stock.gestionar` | 401        | 403         | —     | H09      |

## Carrito y pedidos (`docs/api/pedidos.http`)

| Método        | Ruta                                          | Quién puede                    | Sin sesión | Sin permiso | Ajeno | Historia |
| ------------- | --------------------------------------------- | ------------------------------ | ---------- | ----------- | ----- | -------- |
| GET           | `/api/tienda/configuracion`                   | público                        | —          | —           | —     | H13      |
| GET, DELETE   | `/api/carrito`                                | sesión                         | 401        | —           | —     | H07      |
| POST          | `/api/carrito/items`                          | sesión                         | 401        | —           | —     | H07      |
| PATCH, DELETE | `/api/carrito/items/{id}`                     | sesión                         | 401        | —           | 404   | H07      |
| POST          | `/api/pedidos`                                | público (sesión opcional)      | —          | —           | —     | H13, H15 |
| GET           | `/api/pedidos`                                | sesión                         | 401        | —           | —     | H16      |
| GET           | `/api/pedidos/{id}`                           | sesión, su comprador           | 401        | —           | 404   | H16      |
| GET           | `/api/pedidos/publico/{token}`                | quien tenga el token           | —          | —           | 404 con un token inválido | H16 |
| GET           | `/api/admin/pedidos`                          | `pedidos.gestionar`            | 401        | 403         | —     | H17      |
| GET           | `/api/admin/pedidos/{id}`                     | `pedidos.gestionar`            | 401        | 403         | —     | H17      |
| POST          | `/api/admin/pedidos/{id}/despacho`            | `pedidos.gestionar`            | 401        | 403         | —     | H17      |
| POST          | `/api/admin/pedidos/{id}/entrega`             | `pedidos.gestionar`            | 401        | 403         | —     | H17      |
| POST          | `/api/admin/pedidos/{id}/cobro-efectivo`      | `pedidos.gestionar`            | 401        | 403         | —     | H15, H17 |
| POST          | `/api/admin/pedidos/{id}/cancelacion`         | `pedidos.gestionar`            | 401        | 403         | —     | H18      |

`GET /api/tienda/configuracion` **todavía no existe** (lo agrega Nicolás, `docs/tareas/nicolas.md`): devuelve `efectivoHabilitado`, `distanciaMaximaMetros` y las tarifas `{desdeMetros, hastaMetros, costoCentavos}`, para que el checkout ofrezca efectivo solo si está habilitado y anticipe el costo de envío (H13). El costo que vale es el que calcula `POST /api/pedidos`.

En `POST /api/pedidos`, la identidad sale de la sesión si la hay y el precio sale del catálogo, nunca del cuerpo: un cuerpo que mande precio o usuario no cambia el resultado.

## Pagos (`docs/api/pagos.http`)

| Método | Ruta                       | Quién puede                                        | Sin sesión | Sin permiso | Ajeno | Historia |
| ------ | -------------------------- | -------------------------------------------------- | ---------- | ----------- | ----- | -------- |
| POST   | `/api/pedidos/{id}/pago`   | su comprador: sesión del dueño del pedido o header `X-Token-Acceso` del invitado | 404 | — | 404 | H14 |
| GET    | `/api/pedidos/{id}/pago`   | sesión, su comprador                               | 401        | —           | 404   | H14      |
| POST   | `/api/pagos/webhook`       | Mercado Pago, con firma válida                     | 401 con firma inválida | — | — | H14   |

`POST /api/pedidos/{id}/pago` sin sesión ni token responde 404 y no 401: el pedido no existe para quien no prueba ser su comprador. Además rechaza con 409 un pedido que ya no espera el pago, una reserva vencida y un pedido que ya tiene un pago en curso en Mercado Pago (H14).

## Diferencias entre este contrato y el código

Al escribir este archivo (05/11/2026) aparecieron seis diferencias con el código. Las seis se corrigieron en el mismo PR, con sus pruebas:

| # | Diferencia | Corrección |
| - | ---------- | ---------- |
| 1 | Las rutas de carrito y pedidos no estaban montadas (`Montar` vacío): respondían 404 | `handlers/pedidos/rutas.go` monta las 15 rutas con el control de esta tabla; `rutas_test.go` lo verifica |
| 2 | `POST /api/pedidos/{id}/pago` no verificaba quién paga | Sesión del dueño o `X-Token-Acceso`; si no, 404 (`usecases/pagos.PuedeIniciarPago`) |
| 3 | No verificaba el vencimiento de la reserva ni si había un pago en curso | 409 para reserva vencida o pedido no pendiente; consulta a Mercado Pago por `external_reference` antes de abrir otro pago |
| 4 | La preferencia no incluía el envío y el webhook confirmaba sin comparar montos | El envío entra como una línea más; el webhook solo confirma si lo acreditado es el total y el pedido sigue pendiente (`DecidirConfirmacion`) |
| 5 | No había vencimiento automático a los 15 minutos | `handlers/pagos/vencimiento.go`, cada minuto: cancela y libera stock, salvo que Mercado Pago tenga un pago en curso o no responda |
| 6 | `GET /api/cuenta` no devolvía roles ni permisos | Devuelve los de la sesión (`dtos/identidad.RespuestaCuenta`) |
