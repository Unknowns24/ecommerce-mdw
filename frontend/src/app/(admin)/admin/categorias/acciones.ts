"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requerirPermiso } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Categorías y subcategorías (H08). El padre se elige con un <select>
// nativo; «Ninguna» la deja como categoría principal.
function refrescar() {
  revalidatePath("/admin/categorias");
  revalidatePath("/admin/productos", "layout");
  revalidatePath("/", "layout");
}

export async function crearCategoriaAction(_prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  const valores = valoresDe(formData);
  try {
    await api("/api/admin/categorias", { metodo: "POST", cuerpo: { nombre: valores.nombre ?? "", padreId: valores.padreId || null } });
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar();
  redirect("/admin/categorias");
}

export async function editarCategoriaAction(id: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  const valores = valoresDe(formData);
  try {
    await api(`/api/admin/categorias/${encodeURIComponent(id)}`, {
      metodo: "PATCH",
      cuerpo: { nombre: valores.nombre ?? "", padreId: valores.padreId || null },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar();
  redirect("/admin/categorias");
}

export async function desactivarCategoriaAction(id: string): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  try {
    await api(`/api/admin/categorias/${encodeURIComponent(id)}`, { metodo: "DELETE" });
  } catch (e) {
    return traducirError(e, {});
  }
  refrescar();
  redirect("/admin/categorias");
}
