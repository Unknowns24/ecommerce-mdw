import { notFound, redirect } from "next/navigation";
import { ErrorApi } from "./error-api";

export type EstadoFormulario = {
  enviado: boolean;
  errores?: Record<string, string>;
  mensaje?: string;
  detalles?: unknown;
  valores?: Record<string, string>;
};

export const SIN_ENVIAR: EstadoFormulario = { enviado: false };

export function traducirError(e: unknown, valores: Record<string, string>): EstadoFormulario {
  if (!(e instanceof ErrorApi)) {
    throw e;
  }

  switch (e.status) {
    case 400: {
      const detalles = e.cuerpo.detalles as { campos?: Record<string, string> } | undefined;
      return { enviado: true, errores: detalles?.campos, valores };
    }
    case 409:
      return { enviado: true, mensaje: e.message, detalles: e.cuerpo.detalles, valores };
    case 401:
      redirect("/ingresar");
    case 403:
    case 404:
      notFound();
    default:
      throw e;
  }
}

export function valoresDe(formData: FormData): Record<string, string> {
  const campos = Array.from(formData).filter(
    (entrada): entrada is [string, string] => typeof entrada[1] === "string",
  );
  return Object.fromEntries(campos);
}
