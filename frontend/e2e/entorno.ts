import { readFileSync } from "node:fs";
import path from "node:path";

// Lee frontend/.env.e2e. Lo que ya está en el entorno gana (el CI reemplaza
// DATABASE_DSN por la de su service de Postgres).
export function entornoE2E(): Record<string, string> {
  const archivo = readFileSync(path.join(__dirname, "..", ".env.e2e"), "utf8");
  const valores: Record<string, string> = {};
  for (const linea of archivo.split("\n")) {
    const limpia = linea.trim();
    if (!limpia || limpia.startsWith("#")) continue;
    const i = limpia.indexOf("=");
    if (i < 0) continue;
    valores[limpia.slice(0, i)] = limpia.slice(i + 1);
  }
  for (const clave of Object.keys(valores)) {
    const delEntorno = process.env[clave];
    if (delEntorno) valores[clave] = delEntorno;
  }
  return valores;
}

// El cinturón: las pruebas escriben y borran la base. Si DATABASE_DSN no es
// local, no se arranca nada (AGENTS.md, «Pruebas»).
export function exigirBaseLocal(dsn: string | undefined): void {
  let host = "";
  try {
    host = new URL(dsn ?? "").hostname;
  } catch {
    host = "";
  }
  if (host !== "localhost" && host !== "127.0.0.1") {
    throw new Error(
      `Las pruebas e2e solo corren contra una base local (localhost o 127.0.0.1). DATABASE_DSN apunta a «${host || "?"}».`,
    );
  }
}

export const env = entornoE2E();
export const API_URL = `http://localhost:${env.API_PORT}`;
export const WEB_URL = `http://localhost:${env.WEB_PORT}`;
export const BINARIOS = path.join(__dirname, "..", ".e2e", "bin");
export const ARCHIVO_CUENTAS = path.join(__dirname, "..", ".e2e", "cuentas.json");

// Variables con las que corren la API y sus comandos (migrate y seeds).
export function entornoApi(): Record<string, string> {
  return {
    ENVIRONMENT: env.ENVIRONMENT,
    APP_NAME: env.APP_NAME,
    APP_PORT: env.API_PORT,
    APP_SECRET_KEY: env.APP_SECRET_KEY,
    DATABASE_DSN: env.DATABASE_DSN,
    BOOTSTRAP_MASTER_KEY: env.BOOTSTRAP_MASTER_KEY,
    ORDER_ACCESS_TOKEN_SECRET: env.ORDER_ACCESS_TOKEN_SECRET,
    PAYMENT_RESERVATION_TTL: env.PAYMENT_RESERVATION_TTL,
    PUBLIC_BASE_URL: WEB_URL,
    MERCADOPAGO_WEBHOOK_SECRET: env.MERCADOPAGO_WEBHOOK_SECRET,
    MERCADOPAGO_ACCESS_TOKEN: process.env.MERCADOPAGO_ACCESS_TOKEN ?? "",
  };
}
