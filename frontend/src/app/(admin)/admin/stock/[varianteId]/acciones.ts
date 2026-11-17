"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requerirPermiso } from "@/features/identidad/sesion";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Ajuste de stock (H09): lote, unidades con sentido y motivo. Si deja stock
// negativo o toca unidades reservadas, la API responde 409 y va arriba.
export async function ajustarStockAction(varianteId: string, _prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("stock.gestionar");
  const valores = valoresDe(formData);
  const cantidad = Number.parseInt(valores.unidades ?? "", 10);
  const unidades = Number.isFinite(cantidad) ? (valores.sentido === "restar" ? -cantidad : cantidad) : 0;
  try {
    await api(`/api/admin/variantes/${encodeURIComponent(varianteId)}/ajustes`, {
      metodo: "POST",
      cuerpo: { loteId: valores.loteId ?? "", unidades, motivo: valores.motivo ?? "" },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  revalidatePath(`/admin/stock/${varianteId}`);
  revalidatePath(`/variantes/${varianteId}`);
  revalidatePath("/", "layout");
  redirect(`/admin/stock/${varianteId}?ajustado=1`);
}
