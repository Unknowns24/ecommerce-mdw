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
  // Leer la cookie primero marca la pantalla como dinámica (por request): si
  // no, un error de configuración en el build haría que Next intente
  // prerenderizarla.
  const sesion = (await cookies()).get("sesion")?.value;
  const base = process.env.API_BASE_URL;
  if (!base) {
    throw new Error("Falta configurar API_BASE_URL");
  }

  const url = `${base.replace(/\/+$/, "")}/${ruta.replace(/^\/+/, "")}`;
  const headers = new Headers(opciones?.headers);
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
    // Una API colgada no deja colgada la pantalla: a los 25 s se corta y lo
    // atrapa el error.tsx de la sección. Es más que el peor caso de iniciar
    // un pago (dos llamadas de 10 s a Mercado Pago) y menos que el
    // WriteTimeout de la API (30 s).
    signal: AbortSignal.timeout(25_000),
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
