# Yasmín — identidad, panel y pruebas de permisos

**Sos dueña de:** `app/(tienda)/ingresar/`, `app/(tienda)/registro/`, `app/(cuenta)/layout.tsx`, `app/(cuenta)/cuenta/page.tsx`, `app/(admin)/layout.tsx`, `app/(admin)/not-found.tsx`, `app/(admin)/admin/page.tsx`, `app/(admin)/admin/{usuarios,roles}/`, `features/identidad/` (desde el día 1, incluidos `sesion.ts` y `MenuSesion`), y el harness de pruebas: `playwright.config.ts`, `e2e/preparar.ts`, `e2e/sesiones.ts`, `e2e/permisos.spec.ts` y `.env.e2e` (todo bajo `frontend/`).

**Leé antes de empezar:** `AGENTS.md` (sección «Pruebas»), `docs/api.md` (es la fuente de tus pruebas), `docs/contratos-frontend.md`, `docs/specs.md` (H01–H03, sección 2 y sección 7), ADR-006, ADR-007 y `docs/api/identidad.http`.

**Consumís de otros:** la base del día 0 de Genaro, recién para Y4. Hasta ahí tus tareas no necesitan ninguna pantalla.

**Los demás consumen de vos:** `sesion.ts` (las firmas no cambian), la barra del panel y el harness de Playwright, que Nicolás usa para la prueba del flujo y Agustín corre en el CI. **La configuración del harness es contrato:** el comando es `npm run test:e2e` y las variables que necesita están en `.env.e2e`.

## Tareas

### Y0 — Revisar el arreglo de `GET /api/cuenta` · vie 06/11
`backend/internal/adapters/rest/dtos/identidad/cuenta.go` y el cambio en `handlers/identidad/rutas.go`: devuelve `roles` y `permisos` de la sesión (del token, no de la base; ADR-006). Leelo hasta poder explicarlo y aprobalo en el PR.

### Y1 — Harness de Playwright · vie 06 – sáb 07/11 (no necesita la base)
1. `npm i -D @playwright/test` en un commit aparte.
2. `playwright.config.ts` con **el cinturón:** si `DATABASE_DSN` no apunta a `localhost` o `127.0.0.1`, no arranca.
3. Dos proyectos: `permisos` (sin navegador, le pega a la API) y `flujo` (Chromium, para la prueba de Nicolás).
4. `webServer`: la API de Go compilada y el frontend con `next build && next start` en otro puerto, cada uno con las variables de `.env.e2e`.
5. `e2e/preparar.ts`: migraciones (`/api migrate`) y seeds (`cmd/seed-catalogo`, después `cmd/seed`, después `cmd/seed-pedidos`: ese orden importa), más las cuentas de prueba. Verificá en la documentación de Playwright si `globalSetup` corre antes o después de `webServer`, y ordená según eso.
6. `.env.e2e` con un `APP_SECRET_KEY` que solo existe en las pruebas, y nada de producción.

**Listo cuando:** `npm run test:e2e` levanta todo contra un Postgres local y corre una prueba trivial en verde. Pasale a Agustín la lista de variables para el job de CI.

### Y2 — Sesiones de prueba · sáb 07/11
`e2e/sesiones.ts` fabrica el JWT con el mismo formato que `backend/internal/adapters/jwt/emisor.go` (HS256, claims `sub`, `correo`, `roles`, `permisos`, `exp`, `iat`), firmado con el secreto de `.env.e2e`. Las cuatro cuentas de `docs/api.md`: anónimo, cliente, administrativo con un solo permiso y dueño. **Nada de modo prueba ni usuarios mágicos en la app.**
**Listo cuando:** una prueba con la sesión de cliente obtiene 200 en `GET /api/cuenta` y 403 en `GET /api/admin/pedidos`.

### Y3 — Pruebas de permisos · dom 08 – lun 09/11
`e2e/permisos.spec.ts`: **una prueba por fila de `docs/api.md`**, dada vuelta (sin sesión → 401, sin permiso → 403, ajeno → 404). Para los 404 de recursos ajenos, la prueba crea el recurso con la cuenta A y lo pide con la cuenta B. El valor esperado sale del contrato: si una prueba falla, encontraste un bug. No cambies el número esperado: avisale al dueño del endpoint.
**Listo cuando:** están todas las filas, y viste fallar una a propósito (sacando un `RequierePermiso`) y lo contaste en el PR.

### Y4 — Ingresar, registrarse y salir · mar 10/11 (necesita la base)
`/ingresar` y `/registro` con Server Actions y `useActionState`. El ingreso llama a `POST /api/auth/login` y escribe la cookie `sesion` con los atributos de ADR-007. «Salir» es una acción que llama a `POST /api/auth/logout` y borra la cookie. Completá `MenuSesion` (nombre, «Mi cuenta», «Salir») sin cambiarle la firma.
**Listo cuando:** credenciales inválidas muestran el error arriba sin revelar si el correo existe; un campo vacío, debajo del campo; y el formulario no se vacía.

### Y5 — Mi cuenta · mié 11/11 · H02
`(cuenta)/layout.tsx` con `requerirSesion()` y `/cuenta` para ver y editar los datos (`PATCH /api/cuenta`).

### Y6 — Layout del panel · mié 11/11
`(admin)/layout.tsx`: nombre del sistema, quién sos, tus roles, los módulos que habilitan tus permisos (los links salen de la tabla «Otras pantallas» del mapa) y «Salir». `(admin)/not-found.tsx`, que no dice «permiso». `/admin`: accesos según permisos.
**Listo cuando:** un administrativo con solo `stock.gestionar` ve solo los links de stock y proveedores.

### Y7 — Cuentas y roles · jue 12/11 · H03
`/admin/usuarios` (listar, invitar o crear, activar y desactivar, asignar roles) y `/admin/roles` (listar y crear con permisos). Cada página y cada acción empieza con `requerirPermiso("usuarios.gestionar")`; la regla de dueño la decide el backend.
**Listo cuando:** intentar desactivar al dueño inicial muestra el 409 de la API arriba del formulario.

### Y8 — Recorrido de tus pantallas · sáb 14/11
Las tuyas, con teclado y a 375px.

## Cada día
Una revisión al PR de un compañero (rotación en `docs/plan-de-trabajo.md`). Cuando alguien agregue un endpoint, pedile su fila en `docs/api.md` y sumá su prueba.
