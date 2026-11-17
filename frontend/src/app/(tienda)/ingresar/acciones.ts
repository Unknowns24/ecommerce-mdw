"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { guardarSesion, destinoSeguro } from "@/features/identidad/cookie";
import type { Cuenta } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api, ErrorApi } from "@/lib/api";

// Ingreso (H02). Es público: no se exige sesión. La contraseña nunca vuelve
// en `valores`.
export async function ingresarAction(siguiente: string | undefined, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  const { contrasena, ...valores } = valoresDe(formData);

  let token: string;
  try {
    ({ token } = await api<{ token: string }>("/api/auth/login", {
      metodo: "POST",
      cuerpo: { correo: valores.correo ?? "", contraseña: contrasena ?? "" },
    }));
  } catch (e) {
    // Credenciales incorrectas: arriba, sin decir si el correo existe.
    if (e instanceof ErrorApi && e.status === 401) {
      return { enviado: true, mensaje: "El correo o la contraseña no son correctos.", valores };
    }
    return traducirError(e, valores);
  }

  await guardarSesion(token);
  // Quien administra va al panel; el resto, a donde venía o al catálogo. Los
  // permisos salen de la API, nunca de decodificar el JWT (ADR-007).
  const cuenta = await api<Cuenta>("/api/cuenta", { headers: { Authorization: `Bearer ${token}` } });
  revalidatePath("/", "layout");
  redirect(destinoSeguro(siguiente, cuenta.permisos.length > 0 ? "/admin" : "/"));
}
