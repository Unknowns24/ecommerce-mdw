import { execFileSync } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";

import { API_URL, ARCHIVO_CUENTAS, BINARIOS, entornoApi, env, exigirBaseLocal } from "./entorno";
import { CuentaDePrueba, firmarJWT, NombreCuenta, PERMISOS_DUENO } from "./sesiones";

// globalSetup: corre DESPUÉS de que los webServer respondieron (así ordena
// Playwright las tareas), con la API ya compilada en .e2e/bin.
// 1. Resetea la base local (el cinturón vuelve a mirar el DSN).
// 2. Migraciones y seeds, en este orden: seed-catalogo, seed, seed-pedidos.
// 3. Crea las cuentas de prueba por la API pública y guarda sus ids.

const CONTRASENA = "clave-de-prueba-123";

function correr(binario: string, args: string[] = []) {
  execFileSync(path.join(BINARIOS, binario), args, {
    env: { ...process.env, ...entornoApi() },
    stdio: ["ignore", "ignore", "inherit"],
  });
}

async function pedir<T>(metodo: string, ruta: string, cuerpo: unknown, token?: string): Promise<T> {
  const r = await fetch(`${API_URL}${ruta}`, {
    method: metodo,
    headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: JSON.stringify(cuerpo),
  });
  if (!r.ok) {
    throw new Error(`${metodo} ${ruta} respondió ${r.status}: ${await r.text()}`);
  }
  return (await r.json()) as T;
}

export default async function preparar() {
  exigirBaseLocal(env.DATABASE_DSN);

  // La API ya está corriendo y su tarea periódica (vencimiento de reservas)
  // puede tener la base tomada justo ahora: lock_timeout y hasta 3 intentos.
  for (let intento = 1; ; intento++) {
    try {
      execFileSync(
        "psql",
        [env.DATABASE_DSN, "-q", "-v", "ON_ERROR_STOP=1", "-c", "SET lock_timeout = '5s'; DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"],
        { stdio: ["ignore", "ignore", "pipe"] },
      );
      break;
    } catch (e) {
      if (intento === 3) throw e;
      await new Promise((r) => setTimeout(r, 1000));
    }
  }
  correr("api", ["migrate"]);
  correr("seed-catalogo");
  correr("seed");
  correr("seed-pedidos");

  type Usuario = { id: string };
  const dueno = await pedir<Usuario>("POST", "/api/auth/bootstrap", {
    nombre: "Dueña",
    apellido: "Prueba",
    correo: "duena@e2e.test",
    contraseña: CONTRASENA,
    clave_bootstrap: env.BOOTSTRAP_MASTER_KEY,
  });
  const tokenDueno = firmarJWT({ sub: dueno.id, correo: "duena@e2e.test", roles: ["dueno"], permisos: PERMISOS_DUENO });

  const cliente = await pedir<Usuario>("POST", "/api/auth/registro", {
    nombre: "Clara",
    apellido: "Cliente",
    correo: "cliente@e2e.test",
    contraseña: CONTRASENA,
  });
  const otroCliente = await pedir<Usuario>("POST", "/api/auth/registro", {
    nombre: "Otro",
    apellido: "Cliente",
    correo: "otro@e2e.test",
    contraseña: CONTRASENA,
  });

  // Un administrativo con un solo permiso, para probar el 403 de los demás.
  await pedir("POST", "/api/admin/roles", { nombre: "solo_stock", permisos: ["stock.gestionar"] }, tokenDueno);
  const administrativo = await pedir<Usuario>(
    "POST",
    "/api/admin/usuarios",
    { nombre: "Ana", apellido: "Stock", correo: "stock@e2e.test", contraseña: CONTRASENA },
    tokenDueno,
  );
  await pedir("PUT", `/api/admin/usuarios/${administrativo.id}/roles`, { roles: ["solo_stock"] }, tokenDueno);

  const resultado: Record<NombreCuenta, CuentaDePrueba> = {
    dueno: { id: dueno.id, correo: "duena@e2e.test", contrasena: CONTRASENA, roles: ["dueno"], permisos: PERMISOS_DUENO },
    cliente: { id: cliente.id, correo: "cliente@e2e.test", contrasena: CONTRASENA, roles: ["cliente"], permisos: [] },
    otroCliente: { id: otroCliente.id, correo: "otro@e2e.test", contrasena: CONTRASENA, roles: ["cliente"], permisos: [] },
    administrativo: {
      id: administrativo.id,
      correo: "stock@e2e.test",
      contrasena: CONTRASENA,
      roles: ["solo_stock"],
      permisos: ["stock.gestionar"],
    },
  };
  mkdirSync(path.dirname(ARCHIVO_CUENTAS), { recursive: true });
  writeFileSync(ARCHIVO_CUENTAS, JSON.stringify(resultado, null, 2));
}
