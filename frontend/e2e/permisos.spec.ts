import { APIRequestContext, expect, test } from "@playwright/test";

import { cuentas, NombreCuenta, sesion } from "./sesiones";

// Una prueba por fila de docs/api.md, «dada vuelta»: quién NO puede y qué
// tiene que contestar el backend. Los valores esperados salen de la tabla,
// no del código. Si una falla, encontraste un bug: no cambies el número,
// avisale al dueño del endpoint.

type Metodo = "GET" | "POST" | "PATCH" | "PUT" | "DELETE";

const INEXISTENTE = "99999999-9999-9999-9999-999999999999";
const VARIANTE_CON_STOCK = "11111111-1111-1111-1111-111111111111";

const COMPRADOR = { nombre: "Ana", apellido: "Pérez", correo: "ana@e2e.test", telefono: "3415550000", dni: "30111222" };

async function llamar(request: APIRequestContext, metodo: Metodo, ruta: string, quien: NombreCuenta | "anonimo", cuerpo: unknown = {}, extra: Record<string, string> = {}) {
  return request.fetch(ruta, {
    method: metodo,
    headers: { ...sesion(quien), ...extra },
    data: metodo === "GET" || metodo === "DELETE" ? undefined : cuerpo,
  });
}

// «público»: nadie recibe 401 ni 403.
async function esPublico(request: APIRequestContext, metodo: Metodo, ruta: string, cuerpo: unknown = {}) {
  const r = await llamar(request, metodo, ruta, "anonimo", cuerpo);
  expect(r.status(), `${metodo} ${ruta} sin sesión`).not.toBe(401);
  expect(r.status(), `${metodo} ${ruta} sin sesión`).not.toBe(403);
}

// Sin sesión → 401; con una sesión válida sin el permiso → 403.
async function pidePermiso(request: APIRequestContext, metodos: Metodo[], ruta: string, sinPermiso: NombreCuenta[]) {
  for (const metodo of metodos) {
    expect((await llamar(request, metodo, ruta, "anonimo")).status(), `${metodo} ${ruta} sin sesión`).toBe(401);
    for (const quien of sinPermiso) {
      expect((await llamar(request, metodo, ruta, quien)).status(), `${metodo} ${ruta} como ${quien}`).toBe(403);
    }
  }
}

// Sin sesión → 401 (cualquier sesión puede).
async function pideSesion(request: APIRequestContext, metodos: Metodo[], ruta: string) {
  for (const metodo of metodos) {
    expect((await llamar(request, metodo, ruta, "anonimo")).status(), `${metodo} ${ruta} sin sesión`).toBe(401);
  }
}

async function crearPedido(request: APIRequestContext, quien: NombreCuenta | "anonimo", medioPago = "MERCADO_PAGO") {
  const r = await llamar(request, "POST", "/api/pedidos", quien, {
    items: [{ varianteId: VARIANTE_CON_STOCK, unidades: 1 }],
    comprador: COMPRADOR,
    modoEntrega: "RETIRO",
    medioPago,
  });
  expect(r.status(), "preparación: crear el pedido").toBe(201);
  return (await r.json()) as { id: string; tokenAcceso?: string };
}

// Las cuatro cuentas de la tabla: anónimo, cliente, administrativo con un
// solo permiso (stock.gestionar) y dueño.
const SIN_CATALOGO: NombreCuenta[] = ["cliente", "administrativo"];
const SIN_STOCK: NombreCuenta[] = ["cliente"];
const SIN_PEDIDOS: NombreCuenta[] = ["cliente", "administrativo"];
const SIN_USUARIOS: NombreCuenta[] = ["cliente", "administrativo"];

test.describe("Salud", () => {
  test("GET /api/salud es público", async ({ request }) => {
    await esPublico(request, "GET", "/api/salud");
  });
});

test.describe("Identidad", () => {
  test("POST /api/auth/bootstrap es público, y con un dueño ya existente se rechaza", async ({ request }) => {
    const r = await llamar(request, "POST", "/api/auth/bootstrap", "anonimo", {
      nombre: "Otra",
      apellido: "Dueña",
      correo: "otra-duena@e2e.test",
      contraseña: "clave-segura-123",
      clave_bootstrap: "cualquiera",
    });
    expect([401, 403]).not.toContain(r.status());
    expect(r.ok()).toBe(false);
  });

  test("POST /api/auth/registro es público y nunca crea una cuenta con permisos", async ({ request }) => {
    const r = await llamar(request, "POST", "/api/auth/registro", "anonimo", {
      nombre: "Quiere",
      apellido: "Serdueño",
      correo: "quiere-ser-dueno@e2e.test",
      contraseña: "clave-segura-123",
      rol: "dueno",
      roles: ["dueno"],
      permisos: ["usuarios.gestionar"],
    });
    expect(r.status()).toBe(201);
    const login = await llamar(request, "POST", "/api/auth/login", "anonimo", {
      correo: "quiere-ser-dueno@e2e.test",
      contraseña: "clave-segura-123",
    });
    expect(login.status()).toBe(200);
    const { token } = await login.json();
    const cuenta = await request.get("/api/cuenta", { headers: { Authorization: `Bearer ${token}` } });
    expect((await cuenta.json()).permisos).toEqual([]);
  });

  test("POST /api/auth/login es público: sin sesión previa, las credenciales válidas entran", async ({ request }) => {
    const c = cuentas().cliente;
    const r = await llamar(request, "POST", "/api/auth/login", "anonimo", { correo: c.correo, contraseña: c.contrasena });
    expect(r.status()).toBe(200);
  });

  test("POST /api/auth/logout pide sesión", async ({ request }) => {
    await pideSesion(request, ["POST"], "/api/auth/logout");
  });

  test("GET /api/cuenta pide sesión", async ({ request }) => {
    await pideSesion(request, ["GET"], "/api/cuenta");
  });

  test("PATCH /api/cuenta pide sesión", async ({ request }) => {
    await pideSesion(request, ["PATCH"], "/api/cuenta");
  });

  test("GET /api/admin/usuarios: dueño con usuarios.gestionar", async ({ request }) => {
    await pidePermiso(request, ["GET"], "/api/admin/usuarios", SIN_USUARIOS);
  });

  test("POST /api/admin/usuarios: dueño con usuarios.gestionar", async ({ request }) => {
    await pidePermiso(request, ["POST"], "/api/admin/usuarios", SIN_USUARIOS);
  });

  test("PATCH /api/admin/usuarios/{id}/estado: dueño con usuarios.gestionar", async ({ request }) => {
    await pidePermiso(request, ["PATCH"], `/api/admin/usuarios/${INEXISTENTE}/estado`, SIN_USUARIOS);
  });

  test("PUT /api/admin/usuarios/{id}/roles: dueño con usuarios.gestionar", async ({ request }) => {
    await pidePermiso(request, ["PUT"], `/api/admin/usuarios/${INEXISTENTE}/roles`, SIN_USUARIOS);
  });

  test("GET /api/admin/roles: dueño con usuarios.gestionar", async ({ request }) => {
    await pidePermiso(request, ["GET"], "/api/admin/roles", SIN_USUARIOS);
  });

  test("POST /api/admin/roles: dueño con usuarios.gestionar", async ({ request }) => {
    await pidePermiso(request, ["POST"], "/api/admin/roles", SIN_USUARIOS);
  });
});

test.describe("Catálogo", () => {
  test("GET /api/catalogo/variantes es público", async ({ request }) => {
    await esPublico(request, "GET", "/api/catalogo/variantes");
  });

  test("GET /api/catalogo/variantes/{id} es público", async ({ request }) => {
    await esPublico(request, "GET", `/api/catalogo/variantes/${VARIANTE_CON_STOCK}`);
  });

  test("GET /api/catalogo/categorias es público", async ({ request }) => {
    await esPublico(request, "GET", "/api/catalogo/categorias");
  });

  test("GET /api/catalogo/marcas es público", async ({ request }) => {
    await esPublico(request, "GET", "/api/catalogo/marcas");
  });

  test("GET, POST /api/admin/productos: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["GET", "POST"], "/api/admin/productos", SIN_CATALOGO);
  });

  test("GET, PATCH, DELETE /api/admin/productos/{id}: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["GET", "PATCH", "DELETE"], `/api/admin/productos/${INEXISTENTE}`, SIN_CATALOGO);
  });

  test("GET, POST /api/admin/variantes: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["GET", "POST"], "/api/admin/variantes", SIN_CATALOGO);
  });

  test("GET, PATCH, DELETE /api/admin/variantes/{id}: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["GET", "PATCH", "DELETE"], `/api/admin/variantes/${INEXISTENTE}`, SIN_CATALOGO);
  });

  test("GET /api/admin/variantes/{id}/imagenes: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["GET"], `/api/admin/variantes/${VARIANTE_CON_STOCK}/imagenes`, SIN_CATALOGO);
  });

  test("POST /api/admin/variantes/{id}/imagenes: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["POST"], `/api/admin/variantes/${INEXISTENTE}/imagenes`, SIN_CATALOGO);
  });

  test("PUT /api/admin/variantes/{id}/imagen-principal: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["PUT"], `/api/admin/variantes/${INEXISTENTE}/imagen-principal`, SIN_CATALOGO);
  });

  test("DELETE /api/admin/imagenes/{id}: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["DELETE"], `/api/admin/imagenes/${INEXISTENTE}`, SIN_CATALOGO);
  });

  test("GET, POST /api/admin/marcas: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["GET", "POST"], "/api/admin/marcas", SIN_CATALOGO);
  });

  test("GET, PATCH, DELETE /api/admin/marcas/{id}: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["GET", "PATCH", "DELETE"], `/api/admin/marcas/${INEXISTENTE}`, SIN_CATALOGO);
  });

  test("GET, POST /api/admin/categorias: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["GET", "POST"], "/api/admin/categorias", SIN_CATALOGO);
  });

  test("GET, PATCH, DELETE /api/admin/categorias/{id}: catalogo.escribir", async ({ request }) => {
    await pidePermiso(request, ["GET", "PATCH", "DELETE"], `/api/admin/categorias/${INEXISTENTE}`, SIN_CATALOGO);
  });
});

test.describe("Inventario", () => {
  test("GET, POST /api/admin/proveedores: stock.gestionar", async ({ request }) => {
    await pidePermiso(request, ["GET", "POST"], "/api/admin/proveedores", SIN_STOCK);
  });

  test("POST /api/admin/lotes: stock.gestionar", async ({ request }) => {
    await pidePermiso(request, ["POST"], "/api/admin/lotes", SIN_STOCK);
  });

  test("GET /api/admin/variantes/{id}/stock: stock.gestionar", async ({ request }) => {
    await pidePermiso(request, ["GET"], `/api/admin/variantes/${VARIANTE_CON_STOCK}/stock`, SIN_STOCK);
  });

  test("GET /api/admin/variantes/{id}/movimientos: stock.gestionar", async ({ request }) => {
    await pidePermiso(request, ["GET"], `/api/admin/variantes/${VARIANTE_CON_STOCK}/movimientos`, SIN_STOCK);
  });

  test("POST /api/admin/variantes/{id}/ajustes: stock.gestionar", async ({ request }) => {
    await pidePermiso(request, ["POST"], `/api/admin/variantes/${VARIANTE_CON_STOCK}/ajustes`, SIN_STOCK);
  });
});

test.describe("Carrito y pedidos", () => {
  test("GET /api/tienda/configuracion es público", async ({ request }) => {
    await esPublico(request, "GET", "/api/tienda/configuracion");
  });

  test("GET, DELETE /api/carrito pide sesión", async ({ request }) => {
    await pideSesion(request, ["GET", "DELETE"], "/api/carrito");
  });

  test("POST /api/carrito/items pide sesión", async ({ request }) => {
    await pideSesion(request, ["POST"], "/api/carrito/items");
  });

  test("PATCH, DELETE /api/carrito/items/{id}: sin sesión 401, la línea de otro 404", async ({ request }) => {
    await pideSesion(request, ["PATCH", "DELETE"], `/api/carrito/items/${INEXISTENTE}`);

    const agregado = await llamar(request, "POST", "/api/carrito/items", "cliente", { varianteId: VARIANTE_CON_STOCK, unidades: 1 });
    expect(agregado.ok(), "preparación: el cliente agrega una línea").toBe(true);
    const carrito = await (await llamar(request, "GET", "/api/carrito", "cliente")).json();
    const linea = carrito.items[0].id as string;

    expect((await llamar(request, "PATCH", `/api/carrito/items/${linea}`, "otroCliente", { unidades: 1 })).status()).toBe(404);
    expect((await llamar(request, "DELETE", `/api/carrito/items/${linea}`, "otroCliente")).status()).toBe(404);
    await llamar(request, "DELETE", "/api/carrito", "cliente");
  });

  test("POST /api/pedidos es público (sesión opcional)", async ({ request }) => {
    await esPublico(request, "POST", "/api/pedidos", {});
  });

  test("GET /api/pedidos pide sesión", async ({ request }) => {
    await pideSesion(request, ["GET"], "/api/pedidos");
  });

  test("GET /api/pedidos/{id}: sin sesión 401, el pedido de otro 404", async ({ request }) => {
    await pideSesion(request, ["GET"], `/api/pedidos/${INEXISTENTE}`);
    const pedido = await crearPedido(request, "cliente");
    expect((await llamar(request, "GET", `/api/pedidos/${pedido.id}`, "otroCliente")).status()).toBe(404);
  });

  test("GET /api/pedidos/publico/{token}: 404 con un token inválido", async ({ request }) => {
    expect((await llamar(request, "GET", "/api/pedidos/publico/un-token-que-no-existe", "anonimo")).status()).toBe(404);
  });

  test("GET /api/admin/pedidos: pedidos.gestionar", async ({ request }) => {
    await pidePermiso(request, ["GET"], "/api/admin/pedidos", SIN_PEDIDOS);
  });

  test("GET /api/admin/pedidos/{id}: pedidos.gestionar", async ({ request }) => {
    await pidePermiso(request, ["GET"], `/api/admin/pedidos/${INEXISTENTE}`, SIN_PEDIDOS);
  });

  test("POST /api/admin/pedidos/{id}/despacho: pedidos.gestionar", async ({ request }) => {
    await pidePermiso(request, ["POST"], `/api/admin/pedidos/${INEXISTENTE}/despacho`, SIN_PEDIDOS);
  });

  test("POST /api/admin/pedidos/{id}/entrega: pedidos.gestionar", async ({ request }) => {
    await pidePermiso(request, ["POST"], `/api/admin/pedidos/${INEXISTENTE}/entrega`, SIN_PEDIDOS);
  });

  test("POST /api/admin/pedidos/{id}/cobro-efectivo: pedidos.gestionar", async ({ request }) => {
    await pidePermiso(request, ["POST"], `/api/admin/pedidos/${INEXISTENTE}/cobro-efectivo`, SIN_PEDIDOS);
  });

  test("POST /api/admin/pedidos/{id}/cancelacion: pedidos.gestionar", async ({ request }) => {
    await pidePermiso(request, ["POST"], `/api/admin/pedidos/${INEXISTENTE}/cancelacion`, SIN_PEDIDOS);
  });
});

test.describe("Pagos", () => {
  test("POST /api/pedidos/{id}/pago: sin sesión ni token 404, y la sesión de otro 404", async ({ request }) => {
    const pedido = await crearPedido(request, "cliente");
    expect((await llamar(request, "POST", `/api/pedidos/${pedido.id}/pago`, "anonimo")).status()).toBe(404);
    expect((await llamar(request, "POST", `/api/pedidos/${pedido.id}/pago`, "otroCliente")).status()).toBe(404);
    expect(
      (await llamar(request, "POST", `/api/pedidos/${pedido.id}/pago`, "anonimo", {}, { "X-Token-Acceso": "un-token-que-no-es" })).status(),
    ).toBe(404);
  });

  test("GET /api/pedidos/{id}/pago: sin sesión 401, el pedido de otro 404", async ({ request }) => {
    const pedido = await crearPedido(request, "cliente");
    expect((await llamar(request, "GET", `/api/pedidos/${pedido.id}/pago`, "anonimo")).status()).toBe(401);
    expect((await llamar(request, "GET", `/api/pedidos/${pedido.id}/pago`, "otroCliente")).status()).toBe(404);
  });

  test("POST /api/pagos/webhook: 401 con firma inválida", async ({ request }) => {
    const r = await request.post("/api/pagos/webhook?data.id=123456&type=payment", {
      headers: { "x-signature": "ts=1700000000,v1=firma-invalida", "x-request-id": "req-de-prueba" },
      data: { type: "payment", data: { id: "123456" } },
    });
    expect(r.status()).toBe(401);
  });
});
