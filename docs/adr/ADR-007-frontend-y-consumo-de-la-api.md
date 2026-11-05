# ADR-007 — Frontend en Next.js que consume la API de Go desde el servidor

- Estado: propuesto
- Fecha: 2026-11-05

## Contexto

El backend es una API HTTP en Go (ADR-003) con la sesión en un JWT dentro de la cookie `sesion`, `HttpOnly` y `SameSite=Lax` (ADR-006). Las reglas de negocio, la validación y la autorización ya viven ahí y están probadas.

El stack sugerido por la materia es un único proyecto Next.js donde las pantallas leen la base directamente (`lib/db/`) y los formularios escriben con Server Actions que llaman a los casos de uso. Nuestro proyecto tiene esa misma separación, pero el caso de uso está del otro lado de una llamada HTTP. Las clases 10 a 12 piden pantallas que resuelvan sus cinco estados con `loading.tsx` y `error.tsx`, formularios con Server Actions y `useActionState`, errores por campo y de regla en lugares distintos, e invalidación de las pantallas que quedaron viejas después de escribir.

## Decisión

**Quién llama a la API.** Solo el servidor de Next: Server Components para leer y Server Actions para escribir. Cada módulo tiene sus llamadas en `frontend/src/features/<modulo>/api.ts`, sobre un único cliente compartido (`frontend/src/lib/api.ts`) que:

- arma la URL con `API_BASE_URL` (variable de servidor, sin `NEXT_PUBLIC_`): en producción la dirección interna del contenedor de la API, sin salir a internet;
- reenvía la cookie `sesion` que recibió del navegador;
- devuelve el cuerpo si la respuesta es 2xx y, si no, un error tipado con el status y el `Cuerpo` del contrato de errores (`error`, `detalles`);
- deja subir como excepción todo lo que no sea 400, 404 o 409, para que lo atrape el `error.tsx` de la sección.

Ningún Client Component hace `fetch` a la API.

**La sesión.** La Server Action de ingreso llama a `POST /api/auth/login`, toma el token de la respuesta y escribe la cookie `sesion` en la respuesta de Next con los mismos atributos que usa el backend (`HttpOnly`, `Secure` en producción, `SameSite=Lax`, 24 h). Salir llama a `POST /api/auth/logout` y borra la cookie. Los roles y permisos que se muestran (el rol en la barra, qué módulos ve cada uno en el panel) salen de `GET /api/cuenta`: el frontend nunca decodifica el JWT para decidir nada. Hoy esa respuesta no los incluye; agregarlos es requisito de esta decisión (ver `docs/api.md`, «Diferencias»).

**La validación.** La del backend es la única. El frontend no replica las reglas en Zod: una Server Action traduce la respuesta:

| Respuesta de la API | Qué hace la acción                                                     |
| ------------------- | ---------------------------------------------------------------------- |
| 2xx                 | `revalidatePath` de las pantallas viejas y `redirect` según el mapa    |
| 400                 | Devuelve `detalles.campos` → un error debajo de cada campo             |
| 409                 | Devuelve `error` y `detalles` → un mensaje arriba del formulario       |
| 401                 | Redirige a `/ingresar`                                                 |
| 403, 404            | `notFound()` (mismo texto para «no existe» y «no es tuyo»)             |
| otro                | Excepción → `error.tsx`                                                |

**El despliegue.** Un contenedor más en `docker-compose.yml` (`web`, `next start`) y una regla más en Nginx: `/api/` va al contenedor de la API y todo lo demás al de Next, **bajo el mismo dominio** (`mdw.unkcode.com`). Así la cookie `sesion` sirve para los dos, el navegador nunca hace una llamada cruzada de origen y Mercado Pago vuelve a `/pago/retorno` en el mismo sitio. El CI publica la imagen del frontend junto con la del backend.

**El invitado.** Su carrito vive en `localStorage` (spec, H07) y lo lee un Client Component; el checkout lo manda en el cuerpo de la Server Action, como ya espera `POST /api/pedidos`. El precio nunca sale de ahí: lo vuelve a calcular el backend.

## Consecuencias

**Ganamos:** las reglas, la validación y la autorización siguen en un solo lugar —el backend en Go, ya probado y defendido en el Parcial I—, y las pantallas solo traducen. `loading.tsx` y `error.tsx` funcionan como en el stack sugerido, porque los datos llegan en el servidor. Ningún token llega al JavaScript del navegador. Cada integrante escribe en `features/<modulo>/` y en sus rutas del mapa sin tocar archivos de otro.

**Resignamos:** cada pantalla agrega un salto de red (Next → API) que en el stack sugerido no existe; dentro de la misma máquina es despreciable, pero dos lecturas independientes en una misma página se piden en paralelo (`Promise.all`), nunca una detrás de la otra. No hay validación instantánea en el cliente mientras se escribe: el error aparece al enviar. Y hay dos procesos para desplegar y monitorear en lugar de uno.

**Pendiente:** cómo prueba el flujo e2e el paso de Mercado Pago sin depender del sandbox en el CI. Propuesta: el flujo de punta a punta recorre la alternativa de efectivo con retiro entera, y la rama de Mercado Pago se prueba hasta la redirección al `init_point`; el webhook ya tiene sus pruebas en Go. No se agrega ningún modo prueba a la aplicación.
