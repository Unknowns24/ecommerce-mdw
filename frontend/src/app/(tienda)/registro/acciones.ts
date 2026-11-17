"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { guardarSesion, destinoSeguro } from "@/features/identidad/cookie";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Registro público (H02): crea una cuenta de cliente sin privilegios. El
// formulario no manda roles: aunque los mandara, la API los ignora.
export async function registrarseAction(siguiente: string | undefined, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  const { contrasena, ...valores } = valoresDe(formData);
  const credenciales = { correo: valores.correo ?? "", contraseña: contrasena ?? "" };

  try {
    await api("/api/auth/registro", {
      metodo: "POST",
      cuerpo: { nombre: valores.nombre ?? "", apellido: valores.apellido ?? "", ...credenciales },
    });
  } catch (e) {
    return traducirError(e, valores);
  }

  const { token } = await api<{ token: string }>("/api/auth/login", { metodo: "POST", cuerpo: credenciales });
  await guardarSesion(token);
  revalidatePath("/", "layout");
  redirect(destinoSeguro(siguiente, "/"));
}
