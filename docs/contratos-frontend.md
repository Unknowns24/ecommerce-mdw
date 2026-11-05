# Contrato del frontend — lo que nadie cambia sin avisar

Este archivo hace posible que los cuatro escriban pantallas en paralelo sin esperarse. Fija tres cosas: **quién es dueño de cada carpeta**, **las firmas de las piezas compartidas** y **los puntos de encuentro entre módulos**. Si una firma de acá tiene que cambiar, se abre un Issue y se avisa en el grupo antes de tocarla: los demás ya escribieron código contra ella.

Las direcciones de las pantallas salen del mapa de pantallas (`docs/specs.md`, sección 5); los endpoints, de `docs/api.md` y `docs/api/*.http`; cómo se consume la API, de ADR-007.

## 1. Quién es dueño de qué

Todo bajo `frontend/src/`. Nadie edita una carpeta ajena: si hace falta un cambio ahí, Issue para su dueño.

| Carpeta o archivo                                                                 | Dueño   |
| --------------------------------------------------------------------------------- | ------- |
| `app/layout.tsx`, `app/globals.css`, `components/ui/`, `components.json`, `next.config.ts` | Genaro  |
| `lib/api.ts`, `lib/acciones.ts`, `lib/formato.ts`, `components/campo.tsx`, `components/sin-datos.tsx`, `components/error-de-carga.tsx` | Genaro |
| `app/(tienda)/layout.tsx`, `app/(tienda)/not-found.tsx`, `app/(tienda)/page.tsx`, `app/(tienda)/variantes/` | Genaro |
| `app/(admin)/admin/productos/`, `marcas/`, `categorias/`, `sistema/`              | Genaro  |
| `features/catalogo/`                                                              | Genaro  |
| `app/(tienda)/ingresar/`, `app/(tienda)/registro/`                                 | Yasmín  |
| `app/(cuenta)/layout.tsx`, `app/(cuenta)/cuenta/page.tsx`                          | Yasmín  |
| `app/(admin)/layout.tsx`, `app/(admin)/not-found.tsx`, `app/(admin)/admin/page.tsx`, `admin/usuarios/`, `admin/roles/` | Yasmín |
| `features/identidad/`                                                             | Yasmín  |
| `playwright.config.ts`, `e2e/preparar.ts`, `e2e/sesiones.ts`, `e2e/permisos.spec.ts`, `.env.e2e` | Yasmín |
| `app/(tienda)/carrito/`, `app/(tienda)/checkout/`, `app/(tienda)/pago/`, `app/(tienda)/seguimiento/` | Nicolás |
| `app/(cuenta)/cuenta/pedidos/`                                                    | Nicolás |
| `features/carrito/`, `features/pedidos/`                                          | Nicolás |
| `e2e/flujo-principal.spec.ts`                                                     | Nicolás |
| `app/(admin)/admin/pedidos/`, `admin/proveedores/`, `admin/stock/`                | Agustín |
| `features/backoffice/`, `features/stock/`                                         | Agustín |
| `.github/workflows/`, `docker-compose.yml`, `deploy/`, `frontend/Dockerfile`      | Agustín |

Los nombres de los jobs del CI también son contrato, porque el ruleset de `main` los exige: `Verificar backend` (ya existe), `Verificar frontend` y `e2e` (Agustín).

`frontend/package.json`: cualquiera puede agregar una dependencia, **en un commit aparte y avisando en el grupo**. Los scripts los fija Genaro en la base (sección 4); el `test:e2e` lo completa Yasmín.

## 2. Las piezas compartidas (las escribe Genaro en la base, día 0)

### `lib/api.ts` — la única puerta a la API de Go

Solo se importa desde el servidor (Server Components, Server Actions, `route.ts`). Empieza con `import "server-only"`.

```ts
export type CuerpoError = { error: string; detalles?: unknown };

export class ErrorApi extends Error {
  constructor(public status: number, public cuerpo: CuerpoError) {
    super(cuerpo.error);
  }
}

type Opciones = {
  metodo?: "GET" | "POST" | "PATCH" | "PUT" | "DELETE";
  cuerpo?: unknown;                 // se manda como JSON
  headers?: Record<string, string>; // p. ej. { "X-Token-Acceso": token }
};

// Arma la URL con API_BASE_URL (variable de servidor), reenvía la cookie
// `sesion` como `Authorization: Bearer <token>` si existe, y devuelve el JSON.
// Respuesta 2xx → el cuerpo (o undefined si viene vacío). Si no → lanza ErrorApi.
export async function api<T>(ruta: string, opciones?: Opciones): Promise<T>;
```

### `lib/acciones.ts` — cómo responde toda Server Action

```ts
export type EstadoFormulario = {
  enviado: boolean;
  errores?: Record<string, string>; // 400: detalles.campos → debajo de cada campo
  mensaje?: string;                 // 409: arriba del formulario
  detalles?: unknown;               // 409: lo que la historia pide enumerar
  valores?: Record<string, string>; // lo que la persona escribió: el form no se vacía
};

export const SIN_ENVIAR: EstadoFormulario = { enviado: false };

// Traduce un error de la API al estado del formulario, según ADR-007:
// 400 → errores, 409 → mensaje + detalles, 401 → redirect("/ingresar"),
// 403 y 404 → notFound(), cualquier otro → vuelve a lanzar (lo atrapa error.tsx).
export function traducirError(e: unknown, valores: Record<string, string>): EstadoFormulario;

// Lee los campos de texto de un FormData para devolverlos en `valores`.
export function valoresDe(formData: FormData): Record<string, string>;
```

Forma de toda acción (la de cada módulo, en `acciones.ts` al lado de su `page.tsx`):

```ts
"use server";
export async function crearAlgoAction(prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");          // 1. sesión y permiso
  const valores = valoresDe(formData);
  try {
    await api("/api/admin/...", { metodo: "POST", cuerpo: { ... } }); // 2. la API valida
  } catch (e) {
    return traducirError(e, valores);                   // 3. 400 / 409 al formulario
  }
  revalidatePath("/admin/...");                         // 4. pantallas viejas
  redirect("/admin/...");                               // 5. siempre lo último
}
```

### `lib/formato.ts`

```ts
export function formatearPrecio(centavos: number): string; // 1650000 → "$ 16.500,00" (es-AR, ARS)
export function formatearFecha(iso: string): string;       // fecha y hora en America/Argentina/Buenos_Aires
```

### `components/campo.tsx`

Conecta etiqueta, control, ayuda y error (`htmlFor`/`id` con `useId()`, `aria-describedby`, `aria-invalid`). Por defecto dibuja un `Input`; para un `<select>` o `<textarea>` se pasa una función.

```tsx
type PropsControl = { id: string; name: string; "aria-describedby"?: string; "aria-invalid"?: boolean; defaultValue?: string };

export function Campo(props: {
  etiqueta: string;
  nombre: string;          // name del control y clave en errores / valores
  error?: string;
  ayuda?: string;
  tipo?: string;           // type del input: "text" | "email" | "password" | "tel" ...
  inputMode?: "numeric" | "text" | "email" | "tel";
  requerido?: boolean;
  valorInicial?: string;
  children?: (control: PropsControl) => React.ReactNode;
}): React.ReactElement;
```

### `components/sin-datos.tsx` y `components/error-de-carga.tsx`

```tsx
export function SinDatos(props: { titulo: string; descripcion: string; accion?: { texto: string; href: string } }): React.ReactElement;

// Para usar adentro de cada error.tsx ("use client"): dice que el problema es del
// sistema, que no se perdió nada, y ofrece reintentar. Nunca muestra error.message.
export function ErrorDeCarga(props: { reintentar: () => void }): React.ReactElement;
```

## 3. Los puntos de encuentro entre módulos

Son los únicos lugares donde la pantalla de una persona usa algo de otra. En la base del día 0, Genaro crea cada archivo con la firma de abajo y un cuerpo provisorio. A partir de ahí el archivo pasa a ser de su dueño, que lo completa **sin cambiar la firma**.

| Archivo                                             | Firma congelada                                                                                                              | Lo usa                     | Dueño desde el día 1 |
| --------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | -------------------------- | -------------------- |
| `features/identidad/sesion.ts`                      | `type Cuenta = { id: string; nombre: string; apellido: string; correo: string; roles: string[]; permisos: string[] }` · `obtenerCuenta(): Promise<Cuenta \| null>` (null sin sesión) · `requerirSesion(): Promise<Cuenta>` (sin sesión → `redirect("/ingresar")`) · `requerirPermiso(permiso: string): Promise<Cuenta>` (sin sesión → `/ingresar`; sin el permiso → `notFound()`) | todas las páginas con sesión y todas las acciones | Yasmín |
| `features/identidad/components/menu-sesion.tsx`     | `MenuSesion(): Promise<React.ReactElement>` — Server Component: «Ingresar» sin sesión, o el nombre con «Mi cuenta» y «Salir» | encabezado de `(tienda)` (Genaro) | Yasmín |
| `features/carrito/components/boton-agregar.tsx`     | `BotonAgregarAlCarrito(props: { varianteId: string; disponible: boolean })` — `disponible` es el campo del mismo nombre de `GET /api/catalogo/variantes/{id}`; si es `false` dibuja «SIN STOCK» y no agrega. Si se piden más unidades que las que hay, el 409 de la API lo dice | `/variantes/[id]` (Genaro) | Nicolás |
| `features/carrito/components/enlace-carrito.tsx`    | `EnlaceCarrito()` — link a `/carrito` con la cantidad de unidades (o solo el link)                                            | encabezado de `(tienda)` (Genaro) | Nicolás |

Genaro implementa `sesion.ts` de verdad en la base (son pocas líneas sobre `api("/api/cuenta")`), porque todos lo necesitan desde el primer día; Yasmín es su dueña desde ahí. Los otros tres se crean provisorios: `MenuSesion` muestra solo «Ingresar», `BotonAgregarAlCarrito` es un botón deshabilitado y `EnlaceCarrito` es un link sin cantidad.

**La navegación del panel no es un punto de encuentro:** `app/(admin)/layout.tsx` (Yasmín) arma los links desde el mapa de pantallas y los filtra por permiso. Una pantalla nueva del panel que no esté en el mapa primero se agrega al mapa, y después Yasmín le agrega el link.

## 4. Lo que deja la base del día 0 (PR de Genaro)

- Tailwind y shadcn/ui inicializados; `components/ui/` con `button card badge input label skeleton alert separator table select textarea`. Ese commit va solo, sin pantallas.
- Tokens en `app/globals.css` con contraste 4,5:1.
- Las piezas de la sección 2 y los cuatro archivos de la sección 3.
- Los tres grupos con su layout mínimo: `(tienda)` con encabezado (nombre, `EnlaceCarrito`, `MenuSesion`), `(cuenta)` que llama a `requerirSesion()`, `(admin)` que exige alguna sesión con al menos un permiso.
- Scripts en `package.json`: `dev`, `build`, `start`, `lint`, `typecheck` (`tsc --noEmit`) y `test:e2e` (`playwright test`, lo completa Yasmín).
- `.env.example` del frontend con `API_BASE_URL=http://localhost:8080`.
- `output: "standalone"` en `next.config.ts`, para que Agustín arme la imagen del frontend sin tocar ese archivo.

Hasta que esta base esté mergeada, los demás trabajan en lo que no la necesita (ver `docs/tareas/`).

## 5. Convenciones que cada uno sigue solo

- Los tipos de cada módulo se escriben en `features/<modulo>/tipos.ts`, a partir de los `.http` del backend. No se importan tipos de otro módulo; los de `sesion.ts` sí, porque son contrato.
- Las llamadas a la API de cada módulo van en `features/<modulo>/api.ts`, sobre `lib/api.ts`.
- Cada pantalla que lee datos tiene su `loading.tsx` y su `error.tsx` (con `ErrorDeCarga`) y resuelve vacío y sin resultados con textos distintos (spec, sección 7).
- Las pruebas e2e encuentran los elementos por su texto visible (etiquetas, nombres de botones, estados). Quien cambia un texto de una pantalla del flujo principal después de que exista `e2e/flujo-principal.spec.ts` lo avisa en su PR y etiqueta a Nicolás; quien cambia uno de una pantalla con permisos, a Yasmín.
