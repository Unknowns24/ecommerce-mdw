"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requerirPermiso } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

const opcional = (v: string | undefined) => (v && v.trim() ? v.trim() : null);

// Alta de proveedor (H08): nombre obligatorio; correo, teléfono y web opcionales.
export async function crearProveedorAction(_prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("stock.gestionar");
  const valores = valoresDe(formData);
  try {
    await api("/api/admin/proveedores", {
      metodo: "POST",
      cuerpo: { nombre: valores.nombre ?? "", correo: opcional(valores.correo), telefono: opcional(valores.telefono), web: opcional(valores.web) },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  revalidatePath("/admin/proveedores");
  revalidatePath("/admin/stock/lotes/nuevo");
  redirect("/admin/proveedores");
}
