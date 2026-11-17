"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requerirPermiso } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Marcas (H08). Cambian los filtros del catálogo público: se invalida también.
function refrescar() {
  revalidatePath("/admin/marcas");
  revalidatePath("/admin/productos", "layout");
  revalidatePath("/", "layout");
}

export async function crearMarcaAction(_prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  const valores = valoresDe(formData);
  try {
    await api("/api/admin/marcas", { metodo: "POST", cuerpo: { nombre: valores.nombre ?? "" } });
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar();
  redirect("/admin/marcas");
}

export async function editarMarcaAction(id: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  const valores = valoresDe(formData);
  try {
    await api(`/api/admin/marcas/${encodeURIComponent(id)}`, { metodo: "PATCH", cuerpo: { nombre: valores.nombre ?? "" } });
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar();
  redirect("/admin/marcas");
}

export async function desactivarMarcaAction(id: string): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  try {
    await api(`/api/admin/marcas/${encodeURIComponent(id)}`, { metodo: "DELETE" });
  } catch (e) {
    return traducirError(e, {});
  }
  refrescar();
  redirect("/admin/marcas");
}
