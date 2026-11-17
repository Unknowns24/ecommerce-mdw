"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { obtenerPedido, resolverIncidencia } from "@/features/backoffice/api";
import { requerirPermiso } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Transiciones administrativas (H15, H17, H18). El id sale de .bind; la
// regla de cada transición la decide el backend: repetir una acción vuelve
// con su 409 arriba. Después de cada una se invalidan el detalle y el listado.

function refrescar(id: string) {
  revalidatePath(`/admin/pedidos/${id}`);
  revalidatePath("/admin/pedidos");
}

async function transicion(id: string, ruta: "despacho" | "entrega" | "cobro-efectivo"): Promise<EstadoFormulario> {
  await requerirPermiso("pedidos.gestionar");
  try {
    await api(`/api/admin/pedidos/${encodeURIComponent(id)}/${ruta}`, { metodo: "POST" });
  } catch (e) {
    return traducirError(e, {});
  }
  refrescar(id);
  redirect(`/admin/pedidos/${id}`);
}

export async function despacharAction(id: string): Promise<EstadoFormulario> {
  return transicion(id, "despacho");
}

export async function entregarAction(id: string): Promise<EstadoFormulario> {
  return transicion(id, "entrega");
}

export async function completarEfectivoAction(id: string): Promise<EstadoFormulario> {
  return transicion(id, "cobro-efectivo");
}

export async function resolverIncidenciaAction(pedidoId: string, incidenciaId: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("pedidos.gestionar");
  const valores = valoresDe(formData);
  try {
    await resolverIncidencia(pedidoId, incidenciaId, valores.nota ?? "");
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar(pedidoId);
  redirect(`/admin/pedidos/${pedidoId}`);
}

// Cancelar exige motivo (lo valida la API). Si el pedido estaba pagado,
// además la confirmación explícita de que el reintegro se hizo fuera del
// sistema (spec, sección 6): sin la casilla no se llama a la API. La
// respuesta trae el aviso de que no se devuelve dinero en Mercado Pago, y
// se muestra tal cual.
export async function cancelarAction(id: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("pedidos.gestionar");
  const valores = valoresDe(formData);
  const pedido = await obtenerPedido(id);
  if (pedido.estadoPago === "APROBADO" && valores.reintegro !== "si") {
    return {
      enviado: true,
      errores: { reintegro: "Confirmá que el reintegro ya se hizo fuera del sistema para poder cancelar un pedido pagado." },
      valores,
    };
  }
  let aviso: string | undefined;
  try {
    ({ aviso } = await api<{ aviso?: string }>(`/api/admin/pedidos/${encodeURIComponent(id)}/cancelacion`, {
      metodo: "POST",
      cuerpo: { motivo: valores.motivo ?? "" },
    }));
  } catch (e) {
    return traducirError(e, valores);
  }
  refrescar(id);
  return { enviado: true, detalles: { aviso: aviso ?? null, cancelado: true } };
}
