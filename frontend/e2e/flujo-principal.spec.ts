import { expect, test, type Browser, type Page } from "@playwright/test";
import { execFileSync } from "node:child_process";
import { randomUUID } from "node:crypto";

import { API_URL, env, exigirBaseLocal, WEB_URL } from "./entorno";
import { sesion, tokenDe, type NombreCuenta } from "./sesiones";

// El flujo principal (spec, sección 5), un test.step por fila del mapa de
// pantallas, buscando como una persona: por rol, etiqueta y texto. Sigue la
// propuesta de ADR-007 («Pendiente»): la compra entera por efectivo con
// retiro, y la rama de Mercado Pago solo hasta la redirección al init_point.
// El caso de error del medio es pedir más unidades que las disponibles y ver
// el rechazo que las enumera (H13).

const LOVE_SPELL = "Love Spell crema 236 ml"; // 50 unidades (seed)
const PURE_SEDUCTION_ID = "22222222-2222-2222-2222-222222222222"; // 1 unidad (seed)
const PURE_SEDUCTION = "Pure Seduction splash 250 ml";

async function conSesion(browser: Browser, cuenta: NombreCuenta): Promise<Page> {
  const contexto = await browser.newContext({ baseURL: WEB_URL });
  await contexto.addCookies([{ name: "sesion", value: tokenDe(cuenta), url: WEB_URL, httpOnly: true, sameSite: "Lax" }]);
  return contexto.newPage();
}

async function completarDatos(page: Page) {
  await page.getByLabel("Nombre").fill("Ana");
  await page.getByLabel("Apellido").fill("Pérez");
  await page.getByLabel("Correo").fill("ana@e2e.test");
  await page.getByLabel("Teléfono").fill("3415550000");
  await page.getByLabel("DNI").fill("30111222");
  await page.getByLabel("¿Cómo recibís tu pedido?").selectOption("RETIRO");
}

test("compra como invitado con efectivo y retiro, y la administración la completa", async ({ page, browser }) => {
  // Recorre más de diez pantallas con tres sesiones distintas: en el CI tarda
  // más que el límite por defecto de 30 s (visto: 35,5 s).
  test.setTimeout(90_000);
  let numeroPedido = "";

  await test.step("1 · Catálogo: búsqueda, filtros y orden (/)", async () => {
    await page.goto("/");
    await page.getByLabel("Buscar productos").fill("love spell");
    await page.getByRole("button", { name: "Buscar", exact: true }).click();
    await expect(page).toHaveURL(/q=love\+spell/);
    await expect(page.getByRole("link", { name: LOVE_SPELL })).toBeVisible();
  });

  await test.step("1 · Detalle de una variante y sus alternativas (/variantes/[id])", async () => {
    await page.getByRole("link", { name: LOVE_SPELL }).click();
    await expect(page.getByRole("heading", { level: 1, name: LOVE_SPELL })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Otras presentaciones" })).toBeVisible();
    await page.getByRole("button", { name: "Agregar al carrito" }).click();
    await expect(page.getByText("Agregado al carrito.")).toBeVisible();

    // Más unidades de las que hay: el carrito del invitado no reserva ni
    // compara; el rechazo llega en el checkout.
    await page.goto(`/variantes/${PURE_SEDUCTION_ID}`);
    await page.getByLabel("Unidades").fill("3");
    await page.getByRole("button", { name: "Agregar al carrito" }).click();
    await expect(page.getByText("Agregado al carrito.")).toBeVisible();
  });

  await test.step("2 · Carrito (/carrito)", async () => {
    await page.goto("/carrito");
    await expect(page.getByText("El carrito no reserva unidades ni congela precios.")).toBeVisible();
    await expect(page.getByText(LOVE_SPELL).first()).toBeVisible();
    await expect(page.getByText(PURE_SEDUCTION).first()).toBeVisible();
  });

  await test.step("3 · Ingresar, opcional antes de comprar (/ingresar)", async () => {
    await page.goto("/ingresar");
    await expect(page.getByRole("heading", { name: "Ingresar" })).toBeVisible();
    await expect(page.getByText("podés seguir como invitado")).toBeVisible();
  });

  await test.step("3 · Registrarse, opcional (/registro)", async () => {
    await page.goto("/registro");
    await expect(page.getByRole("heading", { name: "Crear cuenta" })).toBeVisible();
  });

  await test.step("3–5 · Checkout con más unidades que las disponibles: el rechazo las enumera", async () => {
    await page.goto("/checkout");
    await completarDatos(page);
    await page.getByLabel("Medio de pago").selectOption("EFECTIVO");
    await page.getByRole("button", { name: "Confirmar pedido" }).click();
    const rechazo = page.getByRole("alert").filter({ hasText: PURE_SEDUCTION });
    await expect(rechazo).toContainText("Hay 1 unidad disponible.");
    // El formulario no se vació.
    await expect(page.getByLabel("Nombre")).toHaveValue("Ana");
  });

  await test.step("3–5 · Checkout: datos, entrega y resumen (/checkout); 6 · el servidor crea el pedido", async () => {
    await page.goto("/carrito");
    await page.getByRole("button", { name: `Quitar ${PURE_SEDUCTION} del carrito` }).click();
    await expect(page.getByText(PURE_SEDUCTION)).toHaveCount(0);
    await page.getByRole("link", { name: "Continuar al checkout" }).click();
    await completarDatos(page);
    await page.getByLabel("Medio de pago").selectOption("EFECTIVO");
    await page.getByRole("button", { name: "Confirmar pedido" }).click();
    // Efectivo: nace confirmado y la pantalla siguiente es el detalle del pedido.
    await page.waitForURL(/\/seguimiento\/.+/);
  });

  await test.step("9 · Pedido del invitado, por enlace privado (/seguimiento/[token])", async () => {
    const titulo = page.getByRole("heading", { level: 1, name: /Pedido #\d+/ });
    await expect(titulo).toBeVisible();
    numeroPedido = ((await titulo.textContent()) ?? "").replace(/\D/g, "");
    await expect(page.getByText("Confirmado").first()).toBeVisible();
    await expect(page.getByText(LOVE_SPELL).first()).toBeVisible();
  });

  await test.step("7 · Vuelta de Mercado Pago (/pago/retorno?estado=…): nunca dice «pagado» por el parámetro", async () => {
    await page.goto("/pago/retorno?estado=aprobado");
    await expect(page.getByRole("heading", { name: "Estamos verificando tu pago" })).toBeVisible();
  });

  await test.step("9 · Mis pedidos y su detalle (/cuenta/pedidos, /cuenta/pedidos/[id])", async () => {
    const cliente = await conSesion(browser, "cliente");
    await cliente.goto("/");
    await cliente.getByLabel("Buscar productos").fill("love spell");
    await cliente.getByRole("button", { name: "Buscar", exact: true }).click();
    await cliente.getByRole("link", { name: LOVE_SPELL }).click();
    await cliente.getByRole("button", { name: "Agregar al carrito" }).click();
    await expect(cliente.getByText("Agregado al carrito.")).toBeVisible();
    await cliente.goto("/checkout");
    await completarDatos(cliente);
    await cliente.getByLabel("Medio de pago").selectOption("EFECTIVO");
    await cliente.getByRole("button", { name: "Confirmar pedido" }).click();
    await cliente.waitForURL(/\/cuenta\/pedidos\/[0-9a-f-]{36}/);
    await expect(cliente.getByRole("heading", { level: 1, name: /Pedido #\d+/ })).toBeVisible();
    await cliente.goto("/cuenta/pedidos");
    await expect(cliente.getByRole("heading", { name: "Mis pedidos" })).toBeVisible();
    await expect(cliente.getByText(/Pedido #\d+/).first()).toBeVisible();
    await cliente.context().close();
  });

  await test.step("10 · Pedidos del panel (/admin/pedidos) y «Completar pedido» (/admin/pedidos/[id])", async () => {
    const admin = await conSesion(browser, "dueno");
    await admin.goto("/admin/pedidos");
    await admin.getByRole("link", { name: `Pedido #${numeroPedido}`, exact: true }).click();
    await expect(admin.getByRole("heading", { level: 1, name: `Pedido #${numeroPedido}` })).toBeVisible();
    await admin.getByRole("button", { name: "Completar pedido" }).click();
    await expect(admin.getByText("Cobrado en efectivo")).toBeVisible();
    await expect(admin.getByText("Completado").first()).toBeVisible();
    await admin.context().close();
  });
});

// Rama de Mercado Pago hasta la redirección al init_point (ADR-007). Necesita
// un token del sandbox (A0): sin MERCADOPAGO_ACCESS_TOKEN en el entorno no hay
// forma de crear la preferencia, y la app no tiene modo prueba.
test("compra con Mercado Pago hasta la redirección al init_point", async ({ page }) => {
  test.skip(!process.env.MERCADOPAGO_ACCESS_TOKEN, "Falta MERCADOPAGO_ACCESS_TOKEN del sandbox (A0)");

  await test.step("1–2 · Catálogo, detalle y carrito", async () => {
    await page.goto("/");
    await page.getByLabel("Buscar productos").fill("love spell");
    await page.getByRole("button", { name: "Buscar", exact: true }).click();
    await page.getByRole("link", { name: LOVE_SPELL }).click();
    await page.getByRole("button", { name: "Agregar al carrito" }).click();
    await expect(page.getByText("Agregado al carrito.")).toBeVisible();
  });

  await test.step("3–7 · Checkout con Mercado Pago redirige al init_point", async () => {
    await page.goto("/checkout");
    await completarDatos(page);
    await page.getByLabel("Medio de pago").selectOption("MERCADO_PAGO");
    await page.getByRole("button", { name: "Confirmar pedido" }).click();
    await page.waitForURL(/mercadopago\./, { timeout: 30_000 });
  });
});

test("el panel abre la edición de un producto con variantes", async ({ browser }) => {
  const admin = await conSesion(browser, "dueno");
  try {
    await admin.goto("/admin/productos");
    await admin.getByRole("link", { name: "Love Spell", exact: true }).click();
    await expect(admin.getByLabel("Nombre", { exact: true }).first()).toHaveValue("Love Spell");
    await expect(admin.getByRole("button", { name: "Guardar producto", exact: true })).toBeVisible();
  } finally {
    await admin.context().close();
  }
});

test("la administración resuelve una incidencia con nota sin alterar el pedido", async ({ browser, request }) => {
  exigirBaseLocal(env.DATABASE_DSN);
  const r = await request.post(`${API_URL}/api/pedidos`, { data: {
    items: [{ varianteId: "11111111-1111-1111-1111-111111111111", unidades: 1 }],
    comprador: { nombre: "Ana", apellido: "Pérez", correo: "ana@e2e.test", telefono: "3415550000", dni: "30111222" },
    modoEntrega: "RETIRO", medioPago: "MERCADO_PAGO",
  } });
  expect(r.status()).toBe(201);
  const pedido = await r.json() as { id: string; totalCentavos: number };
  expect(pedido.id).toMatch(/^[0-9a-f-]{36}$/);
  const incidencia = randomUUID();
  // Fixture externa, sin puertas ni modo prueba en la aplicación.
  execFileSync("psql", [env.DATABASE_DSN, "-v", "ON_ERROR_STOP=1", "-c",
    `INSERT INTO incidencia_pago (id,pedido_id,referencia_externa,motivo,monto_centavos,total_pedido_centavos) VALUES ('${incidencia}','${pedido.id}','cobro-e2e','MONTO_DISTINTO',100,${pedido.totalCentavos})`,
  ]);
  const ruta = `${API_URL}/api/admin/pedidos/${pedido.id}/incidencias`;
  expect((await request.post(`${API_URL}/api/admin/pedidos/99999999-9999-9999-9999-999999999999/incidencias/${incidencia}/resolucion`, {
    headers: sesion("dueno"), data: { nota: "Verificado" },
  })).status()).toBe(404);
  const admin = await conSesion(browser, "dueno");
  try {
    await admin.goto("/admin/pedidos");
    await expect(admin.getByText("1 incidencia pendiente", { exact: true })).toBeVisible();
    await admin.goto(`/admin/pedidos/${pedido.id}`);
    await expect(admin.getByRole("heading", { name: "Incidencias de pago" })).toBeVisible();
    await admin.getByRole("button", { name: "Registrar resolución" }).click();
    await expect(admin.getByLabel("Cómo se resolvió la incidencia")).toHaveAttribute("aria-invalid", "true");
    const nota = "Se verificó el cobro y se reintegró manualmente en Mercado Pago.";
    await admin.getByLabel("Cómo se resolvió la incidencia").fill(nota);
    await admin.getByRole("button", { name: "Registrar resolución" }).click();
    await expect(admin.getByText(nota, { exact: true })).toBeVisible();
    await expect(admin.getByText(/Resuelta por Dueña Prueba/)).toBeVisible();
    const items = await (await request.get(ruta, { headers: sesion("dueno") })).json();
    expect(items[0].notaResolucion).toBe(nota);
    const actual = await (await request.get(`${API_URL}/api/admin/pedidos/${pedido.id}`, { headers: sesion("dueno") })).json();
    expect(actual.estadoPedido).toBe("PENDIENTE_DE_PAGO");
    expect(actual.estadoPago).toBe("PENDIENTE");
    await admin.goto("/admin/pedidos");
    await expect(admin.getByText("1 incidencia pendiente", { exact: true })).toHaveCount(0);
  } finally { await admin.context().close(); }
});
