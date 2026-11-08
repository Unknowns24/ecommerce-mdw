"use server";

import { revalidatePath } from "next/cache";
import { traducirError, valoresDe } from "@/lib/acciones";
import { ErrorApi } from "@/lib/api";
import { obtenerCuenta, requerirSesion } from "@/features/identidad/sesion";
import {
  agregarItem,
  cambiarUnidades,
  leerVariante,
  quitarItem,
  vaciarCarrito,
  valorizarInvitado,
} from "./api";
import type { EstadoCarrito, ItemInvitado, VistaCarrito } from "./tipos";

// Acciones compartidas por el botón del detalle y la pantalla /carrito (N3).
// Las del cliente exigen sesión; el invitado no tiene carrito en el servidor:
// su lista vive en localStorage y acá solo se valida y se valoriza.

function refrescar() {
  // El carrito cambió: el contador del encabezado y /carrito quedaron viejos.
  revalidatePath("/", "layout");
}

/** Agregar desde el detalle. `varianteId` sale de la URL (.bind), no del form. */
export async function agregarAction(
  varianteId: string,
  _prev: EstadoCarrito,
  formData: FormData,
): Promise<EstadoCarrito> {
  const valores = valoresDe(formData);
  const unidades = Number(valores.unidades);
  const cuenta = await obtenerCuenta();

  if (cuenta) {
    // Cliente: la API valida (400) y compara con el disponible (409).
    try {
      await agregarItem(varianteId, unidades);
    } catch (e) {
      return traducirError(e, valores);
    }
    refrescar();
    return { enviado: true, modo: "cliente", unidades };
  }

  // Invitado: no hay API que compare unidades con el disponible (ver informe,
  // pregunta abierta). Solo se exige una cantidad entera positiva y que la
  // variante exista y no figure sin stock.
  if (!Number.isInteger(unidades) || unidades < 1) {
    return {
      enviado: true,
      errores: { unidades: "Ingresá una cantidad entera de 1 o más." },
      valores,
    };
  }
  try {
    const v = await leerVariante(varianteId);
    if (!v.disponible) {
      return { enviado: true, mensaje: "Este producto está sin stock.", valores };
    }
  } catch (e) {
    return traducirError(e, valores);
  }
  return { enviado: true, modo: "invitado", unidades };
}

/** Cliente: cambiar las unidades de una línea. `itemId` sale de la URL/bind. */
export async function cambiarUnidadesAction(
  itemId: string,
  _prev: EstadoCarrito,
  formData: FormData,
): Promise<EstadoCarrito> {
  await requerirSesion();
  const valores = valoresDe(formData);
  try {
    await cambiarUnidades(itemId, Number(valores.unidades));
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar();
  return { enviado: true, modo: "cliente" };
}

/** Cliente: quitar una línea. */
export async function quitarItemAction(itemId: string): Promise<EstadoCarrito> {
  await requerirSesion();
  try {
    await quitarItem(itemId);
  } catch (e) {
    return traducirError(e, {});
  }
  refrescar();
  return { enviado: true, modo: "cliente" };
}

/** Cliente: vaciar el carrito. */
export async function vaciarCarritoAction(): Promise<EstadoCarrito> {
  await requerirSesion();
  try {
    await vaciarCarrito();
  } catch (e) {
    return traducirError(e, {});
  }
  refrescar();
  return { enviado: true, modo: "cliente" };
}

/**
 * Invitado: recibe lo que había en localStorage y devuelve las líneas con
 * nombre, precio y disponibilidad leídos de la API. Lo que llega del navegador
 * no se confía: se descarta todo lo que no sea {varianteId, unidades enteras >= 1}.
 */
export async function valorizarInvitadoAction(items: ItemInvitado[]): Promise<VistaCarrito> {
  const limpios: ItemInvitado[] = [];
  if (Array.isArray(items)) {
    for (const it of items) {
      if (
        it &&
        typeof it.varianteId === "string" &&
        Number.isInteger(it.unidades) &&
        it.unidades >= 1
      ) {
        limpios.push({ varianteId: it.varianteId, unidades: it.unidades });
      }
    }
  }
  try {
    return await valorizarInvitado(limpios);
  } catch (e) {
    if (e instanceof ErrorApi) console.error("valorizarInvitado", e.status, e.cuerpo);
    throw e;
  }
}

