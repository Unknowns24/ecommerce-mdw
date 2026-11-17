"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { preciosDe } from "@/features/catalogo/admin/formularios";
import { requerirPermiso } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Edición de un producto, sus variantes e imágenes (H08). Los ids salen de
// .bind en la página, nunca de un campo oculto. Después de escribir, se
// invalidan el panel y la tienda (el catálogo público cambió).

const enc = encodeURIComponent;

function refrescar(productoId: string) {
  revalidatePath(`/admin/productos/${productoId}`);
  revalidatePath("/admin/productos");
  revalidatePath("/", "layout");
}

export async function editarProductoAction(id: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  const categorias = formData.getAll("categorias").map(String);
  const valores: Record<string, string> = { ...valoresDe(formData), categorias: categorias.join(",") };
  try {
    await api(`/api/admin/productos/${enc(id)}`, {
      metodo: "PATCH",
      cuerpo: {
        nombre: valores.nombre ?? "",
        marcaId: valores.marcaId || null,
        descripcion: valores.descripcion ?? "",
        metodologiaRotacion: valores.metodologiaRotacion,
        categorias,
      },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar(id);
  redirect(`/admin/productos/${id}?guardado=producto`);
}

export async function desactivarProductoAction(id: string): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  try {
    await api(`/api/admin/productos/${enc(id)}`, { metodo: "DELETE" });
  } catch (e) {
    return traducirError(e, {});
  }
  refrescar(id);
  redirect("/admin/productos");
}

export async function crearVarianteAction(productoId: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  const valores = valoresDe(formData);
  const { minorista, mayorista, errores } = preciosDe(valores);
  if (minorista === null || mayorista === null) return { enviado: true, errores, valores };
  try {
    await api("/api/admin/variantes", {
      metodo: "POST",
      cuerpo: {
        productoId,
        codigo: valores.codigo ?? "",
        nombre: valores.nombre ?? "",
        descripcion: valores.descripcion ?? "",
        precioMinoristaCentavos: minorista,
        precioMayoristaCentavos: mayorista,
      },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar(productoId);
  redirect(`/admin/productos/${productoId}?guardado=variante`);
}

export async function editarVarianteAction(productoId: string, varianteId: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  const valores = valoresDe(formData);
  const { minorista, mayorista, errores } = preciosDe(valores);
  if (minorista === null || mayorista === null) return { enviado: true, errores, valores };
  try {
    await api(`/api/admin/variantes/${enc(varianteId)}`, {
      metodo: "PATCH",
      cuerpo: {
        codigo: valores.codigo ?? "",
        nombre: valores.nombre ?? "",
        descripcion: valores.descripcion ?? "",
        precioMinoristaCentavos: minorista,
        precioMayoristaCentavos: mayorista,
      },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar(productoId);
  revalidatePath(`/variantes/${varianteId}`);
  redirect(`/admin/productos/${productoId}?guardado=variante`);
}

export async function desactivarVarianteAction(productoId: string, varianteId: string): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  try {
    await api(`/api/admin/variantes/${enc(varianteId)}`, { metodo: "DELETE" });
  } catch (e) {
    return traducirError(e, {});
  }
  refrescar(productoId);
  revalidatePath(`/variantes/${varianteId}`);
  redirect(`/admin/productos/${productoId}`);
}

export async function agregarImagenAction(productoId: string, varianteId: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  const valores = valoresDe(formData);
  const orden = Number.parseInt(valores.orden || "0", 10);
  try {
    await api(`/api/admin/variantes/${enc(varianteId)}/imagenes`, {
      metodo: "POST",
      cuerpo: { referenciaImagen: valores.referenciaImagen ?? "", orden: Number.isFinite(orden) ? orden : -1 },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar(productoId);
  revalidatePath(`/variantes/${varianteId}`);
  redirect(`/admin/productos/${productoId}`);
}

export async function quitarImagenAction(productoId: string, varianteId: string, imagenId: string): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  try {
    await api(`/api/admin/imagenes/${enc(imagenId)}`, { metodo: "DELETE" });
  } catch (e) {
    return traducirError(e, {});
  }
  refrescar(productoId);
  revalidatePath(`/variantes/${varianteId}`);
  redirect(`/admin/productos/${productoId}`);
}

export async function imagenPrincipalAction(productoId: string, varianteId: string, imagenId: string): Promise<EstadoFormulario> {
  await requerirPermiso("catalogo.escribir");
  try {
    await api(`/api/admin/variantes/${enc(varianteId)}/imagen-principal`, { metodo: "PUT", cuerpo: { imagenId } });
  } catch (e) {
    return traducirError(e, {});
  }
  refrescar(productoId);
  revalidatePath(`/variantes/${varianteId}`);
  redirect(`/admin/productos/${productoId}`);
}
