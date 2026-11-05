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

En `POST /api/pedidos`, la identidad sale de la sesión si la hay y el precio sale del catálogo, nunca del cuerpo: un cuerpo que mande precio o usuario no cambia el resultado.

## Pagos (`docs/api/pagos.http`)

| Método | Ruta                       | Quién puede                                        | Sin sesión | Sin permiso | Ajeno | Historia |
| ------ | -------------------------- | -------------------------------------------------- | ---------- | ----------- | ----- | -------- |
| POST   | `/api/pedidos/{id}/pago`   | su comprador: sesión del dueño del pedido o token del invitado | **Pendiente** (ver abajo) | — | 404 | H14 |
| GET    | `/api/pedidos/{id}/pago`   | sesión, su comprador                               | 401        | —           | 404   | H14      |
| POST   | `/api/pagos/webhook`       | Mercado Pago, con firma válida                     | 401 con firma inválida | — | — | H14   |

## Diferencias entre este contrato y el código al 05/11/2026

Encontradas al escribir este archivo. Cada una es trabajo, no una corrección del contrato.

1. **Las rutas de carrito y pedidos no están montadas.** `backend/internal/adapters/rest/handlers/pedidos/rutas.go` tiene `Montar` vacío (un `TODO`): los handlers existen y sus pruebas pasan porque arman su propio router, pero el servidor no los expone. Toda la sección «Carrito y pedidos» responde 404 hoy.
2. **`POST /api/pedidos/{id}/pago` no verifica quién paga.** Con sesión opcional y solo el id del pedido, cualquiera que conozca el UUID abre una preferencia nueva. La spec (H16) dice que un pedido ajeno no se alcanza modificando un identificador. **Pendiente:** el invitado se identifica con el token de su enlace privado; con sesión, el pedido tiene que ser suyo.
3. **El mismo endpoint no verifica el vencimiento de la reserva ni si ya hay un intento activo** (H14: «permite reintentar si la reserva sigue vigente y no existe otro pago activo que pueda producir un cobro duplicado»).
4. **La preferencia de Mercado Pago no incluye el envío.** Se arma con las líneas de productos, pero el intento registra el total con envío; y el webhook confirma el pedido sin comparar el monto acreditado con el total. Una compra con envío se cobra de menos y se confirma igual.
5. **No hay vencimiento automático** de los pedidos de Mercado Pago a los 15 minutos (H14): `vence_en` se guarda y nadie lo usa.
6. **`GET /api/cuenta` no devuelve roles ni permisos**, y la barra del panel tiene que mostrar el rol y solo los módulos habilitados (spec, sección 7, «Navegación»; ADR-007).
