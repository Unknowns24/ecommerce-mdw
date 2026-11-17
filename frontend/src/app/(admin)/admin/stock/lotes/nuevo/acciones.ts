"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { aCentavos } from "@/features/catalogo/admin/importe";
import { requerirPermiso } from "@/features/identidad/sesion";
import { fechaDeIngreso } from "@/features/stock/api";
import { traducirError, valoresDe, type EstadoFormulario } from "@/lib/acciones";
import { api } from "@/lib/api";

// Ingreso de lote (H09): variante, proveedor, unidades, costo unitario y fecha.
export async function ingresarLoteAction(_prev: EstadoFormulario, formData: FormData): Promise<EstadoFormulario> {
  await requerirPermiso("stock.gestionar");
  const valores = valoresDe(formData);
  const costo = aCentavos(valores.costoUnitarioCentavos);
  if (costo === null) {
    return { enviado: true, errores: { costoUnitarioCentavos: "Escribí el costo en pesos, por ejemplo 1500,00" }, valores };
  }
  const unidades = Number.parseInt(valores.unidades ?? "", 10);
  const varianteId = valores.varianteId ?? "";
  try {
    await api("/api/admin/lotes", {
      metodo: "POST",
      cuerpo: {
        varianteId,
        proveedorId: valores.proveedorId ?? "",
        unidades: Number.isFinite(unidades) ? unidades : 0,
        costoUnitarioCentavos: costo,
        fecha: fechaDeIngreso(valores.fecha),
      },
    });
  } catch (e) {
    return traducirError(e, valores);
  }
  revalidatePath(`/admin/stock/${varianteId}`);
  revalidatePath(`/variantes/${varianteId}`);
  revalidatePath("/", "layout");
  redirect(`/admin/stock/${varianteId}`);
}
