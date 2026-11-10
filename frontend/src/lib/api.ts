import "server-only";

import { cookies } from "next/headers";
import { ErrorApi, type CuerpoError } from "./error-api";

export { ErrorApi, type CuerpoError } from "./error-api";

type Opciones = {
  metodo?: "GET" | "POST" | "PATCH" | "PUT" | "DELETE";
  cuerpo?: unknown;
  headers?: Record<string, string>;
};

export async function api<T>(ruta: string, opciones?: Opciones): Promise<T> {
  const base = process.env.API_BASE_URL;
  if (!base) {
    throw new Error("Falta configurar API_BASE_URL");
  }

  const url = `${base.replace(/\/+$/, "")}/${ruta.replace(/^\/+/, "")}`;
  const headers = new Headers(opciones?.headers);
  const sesion = (await cookies()).get("sesion")?.value;
  if (sesion) {
    headers.set("Authorization", `Bearer ${sesion}`);
  }
  if (opciones?.cuerpo !== undefined) {
    headers.set("Content-Type", "application/json");
  }

  const respuesta = await fetch(url, {
    method: opciones?.metodo ?? "GET",
    headers,
    body: JSON.stringify(opciones?.cuerpo),
    cache: "no-store",
    redirect: "manual",
  });

  if (!respuesta.ok) {
    const cuerpo: CuerpoError | null = await respuesta.json().catch(() => null);
    throw new ErrorApi(
      respuesta.status,
      typeof cuerpo?.error === "string"
        ? cuerpo
        : { error: respuesta.statusText || "Error de la API" },
    );
  }

  const texto = await respuesta.text();
  return (texto === "" ? undefined : JSON.parse(texto)) as T;
}
