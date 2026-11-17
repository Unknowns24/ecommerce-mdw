"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requerirPermiso } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Cuentas administrativas (H03). Cada acción verifica el permiso en su primera
// línea; la regla de dueño (solo el dueño, y nadie toca al dueño inicial) la
// decide el backend y vuelve como 409 arriba del formulario.

export async function crearUsuarioAction(_prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("usuarios.gestionar");
  const { contrasena, ...valores } = valoresDe(formData);
  try {
    await api("/api/admin/usuarios", {
      metodo: "POST",
      cuerpo: { nombre: valores.nombre ?? "", apellido: valores.apellido ?? "", correo: valores.correo ?? "", contraseña: contrasena ?? "" },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  revalidatePath("/admin/usuarios");
  redirect("/admin/usuarios");
}

// El id sale de .bind en la página, nunca de un campo oculto.
export async function cambiarEstadoAction(id: string, estado: "ACTIVO" | "INACTIVO"): Promise<EstadoFormulario> {
  await requerirPermiso("usuarios.gestionar");
  try {
    await api(`/api/admin/usuarios/${encodeURIComponent(id)}/estado`, { metodo: "PATCH", cuerpo: { estado } });
  } catch (e) {
    return traducirError(e, {});
  }
  revalidatePath("/admin/usuarios");
  redirect("/admin/usuarios");
}

export async function asignarRolesAction(id: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("usuarios.gestionar");
  const roles = formData.getAll("roles").map(String);
  const valores = { roles: roles.join(",") };
  try {
    await api(`/api/admin/usuarios/${encodeURIComponent(id)}/roles`, { metodo: "PUT", cuerpo: { roles } });
  } catch (e) {
    return traducirError(e, valores);
  }
  revalidatePath("/admin/usuarios");
  redirect("/admin/usuarios");
}
