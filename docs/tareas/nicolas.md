# Nicolás — carrito, checkout, consulta de pedidos y prueba del flujo

**Sos dueño de:** `app/(tienda)/{carrito,checkout,pago,seguimiento}/`, `app/(cuenta)/cuenta/pedidos/`, `features/carrito/` (incluidos `BotonAgregarAlCarrito` y `EnlaceCarrito` desde el día 1), `features/pedidos/` y `e2e/flujo-principal.spec.ts` (todo bajo `frontend/`). En el backend, sigue siendo tuyo el módulo de pedidos.

**Leé antes de empezar:** `AGENTS.md`, `docs/contratos-frontend.md`, `docs/specs.md` (H07, H13–H16, el flujo principal y el mapa de pantallas de la sección 5, y la sección 7), ADR-007, `docs/api/pedidos.http` y `docs/api/pagos.http`.

**Consumís de otros:** la base del día 0 de Genaro (desde N2); `sesion.ts` (firma congelada); el harness de Yasmín (solo para N7); y las pantallas de los demás (solo para N7, la última tarea). Iniciar el pago es un endpoint de Agustín que ya existe: lo usás tal como está en `pagos.http`.

**Los demás consumen de vos:** `BotonAgregarAlCarrito` (lo dibuja el detalle de Genaro) y `EnlaceCarrito` (el encabezado). Sus firmas no cambian.

## Tareas

### N0 — Revisar el arreglo de las rutas · vie 06/11
`backend/internal/adapters/rest/handlers/pedidos/rutas.go` y `rutas_test.go`: las 15 rutas montadas con los dos adaptadores hacia catálogo e inventario. Leelo hasta poder explicarlo y aprobalo. Después del deploy, corré `docs/api/pedidos.http` entero contra producción.

### N1 — Configuración pública de la tienda (backend) · vie 06/11 (no necesita la base)
`GET /api/tienda/configuracion`, público: `efectivoHabilitado`, `distanciaMaximaMetros` y las tarifas `{desdeMetros, hastaMetros, costoCentavos}`, desde `RepositorioConfiguracion`. Agregá el caso en `docs/api/pedidos.http`, mantené la fila de `docs/api.md` y sacale la nota de «todavía no existe».
**Listo cuando:** tiene su prueba de handler y responde en producción. Sin configuración cargada responde 404, que el checkout entiende como «el envío no está disponible».

### N2 — El carrito, como módulo · vie 06 – sáb 07/11
`features/carrito/`: las mismas operaciones (ver, agregar, cambiar unidades, quitar, vaciar) para los dos compradores. El **invitado** guarda en `localStorage` solo `{varianteId, unidades}`; el **cliente** usa `/api/carrito`. Los precios nunca se guardan: se leen de la API. Completá `BotonAgregarAlCarrito` y `EnlaceCarrito` sin cambiarles la firma. Antes de esta tarea, resolvé la decisión abierta 3 del plan: al iniciar sesión, el carrito del invitado se descarta o se conserva sin mezclar.
**Listo cuando:** agregar desde el detalle funciona con y sin sesión, y pedir más de lo disponible muestra el 409 de la API.

### N3 — `/carrito` · dom 08/11 · H07
Líneas con variante, precio, cantidad editable y quitar; subtotal y total estimado. Avisa que el carrito no reserva unidades ni congela precios.
**Listo cuando:** el vacío ofrece ir al catálogo, y cambiar cantidades no hace saltar la página.

### N4 — `/checkout` · lun 09 – mié 11/11 · H13, H14, H15
Datos del comprador (nombre, apellido, correo, teléfono y DNI con `inputMode="numeric"`); retiro o envío (domicilio y distancia en metros, que sigue siendo un dato del formulario: spec, sección 9); efectivo **solo** si se eligió retiro y `efectivoHabilitado`; resumen y «Confirmar». La Server Action llama a `POST /api/pedidos` (el invitado manda sus ítems en el cuerpo). Un 409 de stock enumera variante → disponibles arriba del formulario.
- **Efectivo:** redirige al detalle del pedido.
- **Mercado Pago:** guarda el token del invitado en una cookie propia de corta duración (spec, sección 5), muestra el enlace privado, llama a `POST /api/pedidos/{id}/pago` (con `X-Token-Acceso` o con la sesión) y redirige al `initPoint`. Si responde 409 («ya hay un pago en curso», «la reserva venció», «no pudimos iniciar el pago»), el mensaje va arriba y el pedido no se da por pagado.

**Listo cuando:** apretar dos veces «Confirmar» crea un solo pedido, y el formulario no se vacía al fallar.

### N5 — `/pago/retorno` · mié 11/11 · H14
Nunca muestra «pagado» por el parámetro `estado`: dice que el pago se está verificando y lleva al detalle del pedido (cookie del invitado → `/seguimiento/[token]`; sesión → `/cuenta/pedidos/[id]`).

### N6 — Consulta de pedidos · jue 12/11 · H16
`/seguimiento/[token]` con su propio `not-found.tsx` («No encontramos ese pedido»); `/cuenta/pedidos` (paginado, con su vacío) y `/cuenta/pedidos/[id]`. Productos, precios históricos, total, entrega, y los estados del pedido y del pago, cada uno con texto y no solo con color.
**Listo cuando:** un pedido ajeno y uno inexistente muestran el mismo 404.

### N7 — Prueba de punta a punta · vie 13 – sáb 14/11
`e2e/flujo-principal.spec.ts`, en el proyecto `flujo` del harness de Yasmín: un `test.step` por fila del mapa de pantallas, buscando por rol y por etiqueta. El caso de error en el medio es pedir más unidades que las disponibles y ver el rechazo que las enumera. Para el pago, seguí la propuesta de ADR-007, «Pendiente»: la compra completa por efectivo con retiro, y la rama de Mercado Pago hasta la redirección.
**Listo cuando:** está en verde en el CI y la viste fallar a propósito. Es la última tarea porque recorre las pantallas de los cuatro: los pasos de tus propias pantallas los podés escribir antes.

## Cada día
Una revisión al PR de un compañero (rotación en `docs/plan-de-trabajo.md`).
