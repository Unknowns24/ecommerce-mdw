# Genaro — catálogo, sistema visual y base del frontend

**Sos dueño de:** `app/layout.tsx`, `app/globals.css`, `next.config.ts`, `components/ui/`, `components/` compartidos, `lib/`, `app/(tienda)/layout.tsx`, `app/(tienda)/not-found.tsx`, `app/(tienda)/page.tsx`, `app/(tienda)/variantes/`, `app/(admin)/admin/{productos,marcas,categorias,sistema}/` y `features/catalogo/` (todo bajo `frontend/src/`).

**Leé antes de empezar:** `AGENTS.md`, `docs/contratos-frontend.md` (lo escribís vos el viernes: es tu contrato con los demás), `docs/specs.md` (H04, H05, H08, sección 5 y sección 7), ADR-007 y `docs/api/catalogo.http`.

**Consumís de otros:** nada que tengas que esperar. El botón de agregar al carrito y el menú de sesión del encabezado los creás vos como *stub* en la base, y después los completan Nicolás y Yasmín sin cambiarles la firma.

**Los demás consumen de vos:** la base del día 0. Es la única tarea del equipo que alguien espera, y por eso es la primera.

## Tareas

### G0 — Base del frontend · **viernes 06/11, mergeada ese mismo día**

Exactamente lo de `docs/contratos-frontend.md`, sección 4:

1. Tailwind + `npx shadcn@latest init` y `add button card badge input label skeleton alert separator table select textarea`. **Este commit va solo**, sin pantallas.
2. Tokens en `globals.css`: color primario, radio y tipografía, con contraste 4,5:1 sobre el fondo.
3. `lib/api.ts`, `lib/acciones.ts`, `lib/formato.ts`, `components/campo.tsx`, `components/sin-datos.tsx` y `components/error-de-carga.tsx`, con las firmas exactas del contrato.
4. `features/identidad/sesion.ts` implementado de verdad (sobre `api("/api/cuenta")`; un 401 devuelve `null`), y los tres *stubs*: `MenuSesion` (solo «Ingresar»), `BotonAgregarAlCarrito` (deshabilitado) y `EnlaceCarrito` (link a `/carrito`).
5. Los layouts mínimos de `(tienda)`, `(cuenta)` y `(admin)`. Borrá `app/page.tsx` y `page.module.css` del template.
6. Scripts `typecheck` y `test:e2e`, `output: "standalone"` y `.env.example` con `API_BASE_URL`.

**Listo cuando:** `npm run lint`, `npm run typecheck` y `npm run build` pasan, `npm run dev` contra la API local muestra el encabezado, y el PR está mergeado en `develop`. Avisá en el grupo apenas se mergee.

### G1 — Encabezado de la tienda y su 404 · sáb 07/11
`(tienda)/layout.tsx` con el nombre de la tienda (link a `/`), `EnlaceCarrito` y `MenuSesion`. `(tienda)/not-found.tsx` con salida al catálogo.
**Listo cuando:** contesta «¿dónde estoy?» y «¿cómo salgo?» en todas las pantallas de la tienda, también a 375px.

### G2 — Catálogo `/` · sáb 07 – dom 08/11 · H04
Tarjetas por variante con imagen, nombre y precio (`formatearPrecio`); búsqueda por nombre, filtros por categoría y marca, orden por precio ascendente y descendente, paginación que conserva la selección (todo en la query string, contra `GET /api/catalogo/variantes`). «SIN STOCK» visible en las agotadas.
**Listo cuando:** resuelve los cinco estados. **Vacío** («todavía no hay productos») y **sin resultados** («ningún producto coincide con "…"», con un link para limpiar los filtros) dicen cosas distintas. `loading.tsx` es un esqueleto con forma de grilla y `error.tsx` usa `ErrorDeCarga`.

### G3 — Detalle `/variantes/[id]` · lun 09/11 · H05
Descripción, marca, categorías, imágenes (si una falla, una alternativa visual que no rompe la página), precio, disponibilidad, links a las otras variantes del producto y `BotonAgregarAlCarrito` con el `disponible` de la API. Una variante inactiva o inexistente es `notFound()`.
**Listo cuando:** tiene su `loading.tsx`, y un id inválido muestra el 404 de la tienda.

### G4 — Productos del panel · mié 11/11 · H08
`/admin/productos`: listado y alta. `/admin/productos/[id]`: edición, variantes (alta y edición con precios minorista y mayorista), imágenes (agregar, quitar, elegir la principal) y un link «Ver stock» de cada variante a `/admin/stock/[varianteId]` (la pantalla es de Agustín; alcanza con la dirección). Cada página empieza con `requerirPermiso("catalogo.escribir")`, y cada acción también.
**Listo cuando:** cada formulario muestra el 400 debajo del campo y el 409 arriba, no se vacía al fallar, deshabilita el botón mientras guarda y hace `revalidatePath` de lo que quedó viejo.

### G5 — Marcas y categorías · jue 12/11 · H08
`/admin/marcas` y `/admin/categorias` (con subcategorías: el padre se elige con un `<select>` nativo).
**Listo cuando:** lo mismo que G4.

### G6 — Galería `/admin/sistema` · jue 12/11
Cada componente de `components/ui/` con sus variantes, más `SinDatos`, `ErrorDeCarga`, los esqueletos, un `Campo` con error y un texto larguísimo.
**Listo cuando:** se puede chequear contraste y foco de todo el sistema en esa sola pantalla.

### G7 — Ruleset de `main` · vie 13/11
Cuando estén en verde los jobs `Verificar frontend` y `e2e` de Agustín: Settings → Rulesets → regla para `main` que exige `Verificar backend`, `Verificar frontend` y `e2e`, con *Enforcement* en *Active*.
**Listo cuando:** un PR con una prueba roja no se puede mergear (probalo con un PR de prueba y cerralo).

### G8 — Recorrido de tus pantallas · sáb 14/11
Las tuyas, con teclado y a 375px. Lo que encuentres va en un PR.

## Cada día
Una revisión al PR de un compañero (rotación en `docs/plan-de-trabajo.md`).
