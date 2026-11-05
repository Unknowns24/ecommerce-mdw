# Agustín — back-office de pedidos y stock, infraestructura y CI

**Sos dueño de:** `app/(admin)/admin/{pedidos,proveedores,stock}/`, `features/backoffice/` y `features/stock/` (bajo `frontend/src/`); `.github/workflows/`, `docker-compose.yml`, `deploy/` y `frontend/Dockerfile`. En el backend siguen siendo tuyos inventario y pagos.

**Leé antes de empezar:** `AGENTS.md`, `docs/contratos-frontend.md`, `docs/specs.md` (H09, H15, H17, H18 y sección 7), ADR-005, ADR-007, `docs/api/inventario.http`, `docs/api/pedidos.http` (transiciones administrativas) y `docs/api/pagos.http`.

**Consumís de otros:** la base del día 0 de Genaro (desde A4); `sesion.ts` (firma congelada); para el job `e2e`, el comando `npm run test:e2e` y las variables de `.env.e2e` del harness de Yasmín. Hasta A4, tus tareas no necesitan ninguna pantalla.

**Los demás consumen de vos:** que el frontend esté en producción (A2) y los jobs del CI con los nombres del contrato (`Verificar frontend` y `e2e`), que Genaro exige en el ruleset de `main`.

## Tareas

### A0 — Revisar los arreglos de pagos y probarlos en el sandbox · vie 06/11
`backend/internal/usecases/pagos/`, `handlers/pagos/iniciar.go`, `webhook.go`, `vencimiento.go` y `adapters/external/mp/cliente.go`. Leelos hasta poder explicarlos y aprobalos. Después, en el sandbox de Mercado Pago, lo que no se pudo probar sin red:
- una compra **con envío** de punta a punta, en la que se cobre el total;
- que `GET /v1/payments/search?external_reference=<pedidoId>` responda con nuestro token (si falla, iniciar un pago responde 409);
- que un pedido sin pagar quede cancelado a los 15 minutos, con su stock liberado.

### A1 — README y ADR-005 · vie 06/11
Lo hace el PR de los parches (defecto 7), pero el mecanismo con el que la VPS toma la imagen nueva de GHCR solo lo sabés vos: contestáselo al asistente cuando te lo pregunte, y revisá que quede bien escrito.

### A2 — El frontend en producción · sáb 07 – dom 08/11 (no necesita la base)
- `frontend/Dockerfile` multi-stage con `output: "standalone"` (lo agrega Genaro en la base), usuario no root.
- Servicio `web` en `docker-compose.yml` con `API_BASE_URL=http://api:8080`.
- `deploy/nginx/`: `/api/` va a `api`, todo lo demás a `web`, bajo `mdw.unkcode.com` (ADR-007). Que el webhook de Mercado Pago siga entrando por `/api/pagos/webhook`.
- La imagen del frontend se publica en GHCR igual que la del backend.

**Listo cuando:** `https://mdw.unkcode.com/` muestra el frontend, `https://mdw.unkcode.com/api/salud` sigue respondiendo, y la cookie de sesión vale para los dos.

### A3 — CI del frontend y job e2e · lun 09 – vie 13/11
- Job **`Verificar frontend`** (desde el lunes, cuando la base tiene sus scripts): `npm ci`, `lint`, `typecheck` y `build`, con Node 22.
- Job **`e2e`** (cuando el harness de Yasmín funcione): Postgres 16 como *service*, Go 1.26 y Node 22, `npx playwright install --with-deps chromium`, `npm run test:e2e` con las variables de `.env.e2e` (solo se reemplaza `DATABASE_DSN` por la del *service*), y `test-results/` como artifact también cuando falla.
- Los dos corren en cada PR, en paralelo con `Verificar backend`.

**Listo cuando:** los tres jobs están en verde en un PR, y avisaste a Genaro para que active el ruleset (G7).

### A4 — Pedidos del panel · mar 10 – mié 11/11 · H17
`/admin/pedidos`: listado paginado con filtros por estado y fecha (`desde`, `hasta`), los cinco estados de pantalla, estado del pedido y del pago con texto. Cada página y cada acción empieza con `requerirPermiso("pedidos.gestionar")`.

### A5 — Detalle y transiciones · mié 11 – jue 12/11 · H15, H17, H18
`/admin/pedidos/[id]`: comprador, pago, productos y entrega. Acciones, cada una visible solo en el estado que la permite (el backend igual valida): «Marcar despachado», «Marcar entregado», «Completar pedido» (efectivo: registra cobro y entrega) y «Cancelar», con motivo obligatorio. Si el pedido estaba pagado, la cancelación pide confirmar que el reintegro se hizo fuera del sistema y muestra el aviso que devuelve la API. Después de cada acción, `revalidatePath` del detalle y del listado.
**Listo cuando:** repetir una acción muestra el 409 arriba, y cancelar un pedido pagado no se puede hacer sin la confirmación.

### A6 — Stock y proveedores · jue 12 – vie 13/11 · H08, H09
`/admin/proveedores` (listado y alta), `/admin/stock/[varianteId]` (disponible, lotes, movimientos y ajuste con motivo) y `/admin/stock/lotes/nuevo` (variante, proveedor, unidades, costo unitario y fecha). Permiso `stock.gestionar`. A `/admin/stock/[varianteId]` se llega desde el link «Ver stock» que pone Genaro en cada variante: la dirección ya está fijada.
**Listo cuando:** un ajuste que deja stock negativo muestra el 409 de la API arriba del formulario.

### A7 — Recorrido en un teléfono real · sáb 14/11
El checkout de Nicolás en tu teléfono, contra producción y con datos móviles (spec, sección 7). Lo que veas que DevTools no muestra va al PR de Nicolás. Después, tus pantallas con teclado y a 375px.

## Cada día
Una revisión al PR de un compañero (rotación en `docs/plan-de-trabajo.md`).
