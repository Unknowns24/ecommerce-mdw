import { createHmac } from "node:crypto";
import { readFileSync } from "node:fs";

import { ARCHIVO_CUENTAS, env } from "./entorno";

// Las sesiones de las pruebas se fabrican afuera de la aplicación: un JWT con
// el mismo formato que backend/internal/adapters/jwt/emisor.go (HS256; claims
// sub, correo, roles, permisos, exp, iat), firmado con el APP_SECRET_KEY de
// .env.e2e. La app no tiene modo prueba ni usuarios mágicos: las cuentas las
// crea e2e/preparar.ts por la API pública, como cualquier instalación.

export type NombreCuenta = "cliente" | "otroCliente" | "administrativo" | "dueno";

export type CuentaDePrueba = { id: string; correo: string; contrasena: string; roles: string[]; permisos: string[] };

export const PERMISOS_DUENO = ["catalogo.escribir", "pedidos.gestionar", "stock.gestionar", "usuarios.gestionar"];

const b64 = (x: string | Buffer) => Buffer.from(x).toString("base64url");

export function firmarJWT(claims: { sub: string; correo: string; roles: string[]; permisos: string[] }, secreto = env.APP_SECRET_KEY): string {
  const ahora = Math.floor(Date.now() / 1000);
  const cuerpo = { ...claims, exp: ahora + 60 * 60, iat: ahora };
  const mensaje = `${b64(JSON.stringify({ alg: "HS256", typ: "JWT" }))}.${b64(JSON.stringify(cuerpo))}`;
  const firma = createHmac("sha256", secreto).update(mensaje).digest("base64url");
  return `${mensaje}.${firma}`;
}

export function cuentas(): Record<NombreCuenta, CuentaDePrueba> {
  return JSON.parse(readFileSync(ARCHIVO_CUENTAS, "utf8"));
}

export function tokenDe(nombre: NombreCuenta): string {
  const c = cuentas()[nombre];
  return firmarJWT({ sub: c.id, correo: c.correo, roles: c.roles, permisos: c.permisos });
}

// Headers para el proyecto `permisos`. «anonimo» no manda nada.
export function sesion(nombre: NombreCuenta | "anonimo"): Record<string, string> {
  return nombre === "anonimo" ? {} : { Authorization: `Bearer ${tokenDe(nombre)}` };
}
