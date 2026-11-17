"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requerirPermiso } from "@/features/identidad/sesion";
import { traducirError, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Crear un rol (H03). Crear un rol no crea funcionalidades: solo agrupa
// permisos que ya existen.
export async function crearRolAction(_prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("usuarios.gestionar");
  const nombre = String(formData.get("nombre") ?? "");
  const permisos = formData.getAll("permisos").map(String);
  const valores = { nombre, permisos: permisos.join(",") };
  try {
    await api("/api/admin/roles", { metodo: "POST", cuerpo: { nombre, permisos } });
  } catch (e) {
    return traducirError(e, valores);
  }
  revalidatePath("/admin/roles");
  revalidatePath("/admin/usuarios");
  redirect("/admin/roles");
}
