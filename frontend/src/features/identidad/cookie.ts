import "server-only";

import { cookies } from "next/headers";

// La cookie de sesión con los mismos atributos que usa el backend (ADR-006 y
// ADR-007): HttpOnly, Secure en producción, SameSite=Lax y 24 h.
export async function guardarSesion(token: string): Promise<void> {
  (await cookies()).set("sesion", token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 24 * 60 * 60,
  });
}

export async function borrarSesion(): Promise<void> {
  (await cookies()).delete("sesion");
}

// Solo direcciones internas: nada de redirigir afuera con un ?siguiente=.
export function destinoSeguro(siguiente: string | undefined, porDefecto: string): string {
  if (siguiente && siguiente.startsWith("/") && !siguiente.startsWith("//")) return siguiente;
  return porDefecto;
}
