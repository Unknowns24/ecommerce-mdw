import { expect, test } from "@playwright/test";

import { sesion } from "./sesiones";

// Y2 — la sesión fabricada vale para la API como una de verdad.
test("la sesión de cliente obtiene 200 en GET /api/cuenta y 403 en GET /api/admin/pedidos", async ({ request }) => {
  const cuenta = await request.get("/api/cuenta", { headers: sesion("cliente") });
  expect(cuenta.status()).toBe(200);
  expect((await cuenta.json()).roles).toEqual(["cliente"]);

  const admin = await request.get("/api/admin/pedidos", { headers: sesion("cliente") });
  expect(admin.status()).toBe(403);
});
