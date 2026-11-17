"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requerirPermiso } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Alta de producto (H08). Al crearlo se va a su edición, donde se cargan
// variantes e imágenes.
export async function crearProductoAction(_prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  const categorias = formData.getAll("categorias").map(String);
  const valores: Record<string, string> = { ...valoresDe(formData), categorias: categorias.join(",") };
  let id: string;
  try {
    ({ id } = await api<{ id: string }>("/api/admin/productos", {
      metodo: "POST",
      cuerpo: {
        nombre: valores.nombre ?? "",
        marcaId: valores.marcaId || null,
        descripcion: valores.descripcion ?? "",
        metodologiaRotacion: valores.metodologiaRotacion ?? "FIFO",
        categorias,
      },
    }));
  } catch (e) {
    return traducirError(e, valores);
  }
  revalidatePath("/admin/productos");
  revalidatePath("/", "layout");
  redirect(`/admin/productos/${id}`);
}
