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

1. **Las direcciones de las pantallas están congeladas** en el mapa de pantallas (spec, sección 5). Para enlazar a una pantalla de otro alcanza con su dirección.
2. **El contrato de la API está congelado** en `docs/api.md` y `docs/api/*.http`. Cada pantalla consume endpoints que ya existen.
3. **El contrato del frontend está congelado** en [`docs/contratos-frontend.md`](contratos-frontend.md): quién es dueño de cada carpeta, las firmas de las piezas compartidas y los cuatro puntos de encuentro entre módulos, que nacen como *stub* en la base del día 0.
4. **Las pruebas se escriben desde la spec y el contrato**, no desde el código del compañero: se pueden escribir en paralelo.

La única espera del equipo es la base del frontend (Genaro, viernes 06/11). Mientras tanto, los otros tres tienen trabajo que no la necesita.

### Git

- `main` protegida: solo se mergea desde `develop`, y con los checks del CI obligatorios (sección 4, viernes 13/11).
- Una rama por tarea: `feat/<modulo>/<descripcion>` o `fix/<modulo>/<descripcion>`. PR chico contra `develop`, con una review aprobada de alguien que no lo escribió.
- `git pull --rebase origin develop` antes de abrir el PR: el conflicto lo resuelve quien lo genera.
- Cada uno commitea con su propia cuenta.

---

## 3. Las tareas de cada uno

Cada integrante tiene su archivo en [`docs/tareas/`](tareas/README.md), con qué es suyo, qué consume, sus tareas en orden y cuándo cada una está terminada. Se puede pasar tal cual al asistente de IA.

| Integrante | Archivo                                   | Antes de la base (vie 06/11)                                   | Después                                                                  |
| ---------- | ----------------------------------------- | -------------------------------------------------------------- | ------------------------------------------------------------------------ |
| Genaro     | [genaro.md](tareas/genaro.md)             | **La base del frontend**                                       | Catálogo, detalle, productos, marcas, categorías, `/admin/sistema`, ruleset |
| Yasmín     | [yasmin.md](tareas/yasmin.md)             | Revisa el defecto 5; harness de Playwright, sesiones y pruebas de permisos | Ingresar, registro, cuenta, layout del panel, usuarios y roles   |
| Nicolás    | [nicolas.md](tareas/nicolas.md)           | Revisa el defecto 1; `GET /api/tienda/configuracion`; lógica del carrito | Carrito, checkout, retorno de pago, consulta de pedidos, e2e del flujo |
| Agustín    | [agustin.md](tareas/agustin.md)           | Revisa 2, 3 y 4 y los prueba en el sandbox; frontend en producción | CI del frontend y e2e, pedidos del panel, stock y proveedores         |

Los defectos 6, 7 y 8 de la sección 1 se resuelven en el mismo PR que trae este plan.

## 4. Calendario hasta el Parcial II

| Día            | Genaro                              | Yasmín                                   | Nicolás                                     | Agustín                                       |
| -------------- | ----------------------------------- | ---------------------------------------- | ------------------------------------------- | --------------------------------------------- |
| **vie 06/11**  | **G0 — base mergeada**              | Y0 revisión · Y1 harness                 | N0 revisión · N1 configuración · N2 carrito | A0 revisión y sandbox · A1                    |
| sáb 07/11      | G1 encabezado · G2 catálogo         | Y1 harness · Y2 sesiones                 | N2 carrito                                  | A2 frontend en producción                     |
| dom 08/11      | G2 catálogo                         | Y3 permisos                              | N3 `/carrito`                               | A2 frontend en producción                     |
| lun 09/11      | G3 detalle                          | Y3 permisos                              | N4 checkout                                 | A3 job `Verificar frontend`                   |
| **mar 10/11** — **TP N° 1** | — | Y4 ingresar y registro                 | N4 checkout                                 | A4 pedidos del panel                          |
| mié 11/11      | G4 productos                        | Y5 cuenta · Y6 layout del panel          | N4 checkout · N5 retorno                    | A4 · A5 transiciones                          |
| jue 12/11      | G5 marcas y categorías · G6 galería | Y7 usuarios y roles                      | N6 consulta de pedidos                      | A5 · A6 stock y proveedores                   |
| **vie 13/11**  | **G7 ruleset de `main`**            | Pruebas de permisos al día               | N7 e2e del flujo                            | **A3 job `e2e` en verde** · A6                |
| sáb 14/11      | G8 recorrido                        | Y8 recorrido                             | N7 e2e del flujo en verde                   | A7 checkout en un teléfono real               |
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
