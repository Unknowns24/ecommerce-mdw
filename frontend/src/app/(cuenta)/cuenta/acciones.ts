"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requerirSesion } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Editar los datos propios (H02). La identidad sale de la sesión: el cuerpo
// no lleva id. El correo no se cambia desde acá (spec, H02: «Pendiente»).
export async function actualizarCuentaAction(_prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirSesion();
  const valores = valoresDe(formData);
  const telefono = valores.telefono?.trim();
  try {
    await api("/api/cuenta", {
      metodo: "PATCH",
      cuerpo: { nombre: valores.nombre ?? "", apellido: valores.apellido ?? "", telefono: telefono ? telefono : null },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  revalidatePath("/", "layout");
  redirect("/cuenta?guardado=1");
}
