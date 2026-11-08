import "server-only";
import { cookies } from "next/headers";

// Cookie propia y de corta duración con el token del enlace privado del
// invitado (spec, sección 5, «Propuesta para revisión»): permite que
// /pago/retorno lo lleve a su pedido sin poner el token en las back_urls de
// Mercado Pago ni en ninguna URL fuera de /seguimiento/[token].
//
// El nombre y la duración NO están fijados en la spec: 15 minutos, igual que la
// reserva de Mercado Pago (spec, H14). HttpOnly (el navegador no la lee) y Lax
// (la vuelta de Mercado Pago es una navegación GET de nivel superior, que Lax
// sí envía; Strict no).
export const COOKIE_ACCESO = "acceso-pedido";
const DURACION_SEGUNDOS = 15 * 60;

export async function guardarTokenAcceso(token: string): Promise<void> {
  (await cookies()).set(COOKIE_ACCESO, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: DURACION_SEGUNDOS,
  });
}

export async function leerTokenAcceso(): Promise<string | undefined> {
  return (await cookies()).get(COOKIE_ACCESO)?.value || undefined;
}
