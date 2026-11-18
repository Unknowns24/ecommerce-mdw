import { unstable_rethrow } from "next/navigation";

import { obtenerCondicionesEntrega } from "@/features/catalogo/api";

// Barra fina negra sobre el encabezado: solo condiciones reales de la tienda
// (GET /api/tienda/configuracion). Sin promociones inventadas: si la tienda
// todavía no cargó su configuración, la barra no aparece.
export async function BarraPromocion() {
  const condiciones = await obtenerCondicionesEntrega().catch((e: unknown) => {
    unstable_rethrow(e);
    console.error("barra de promoción", e);
    return null;
  });
  if (!condiciones) return null;
  const km = (condiciones.distanciaMaximaMetros / 1000).toLocaleString("es-AR", { maximumFractionDigits: 1 });
  return (
    <div className="bg-primary px-4 py-2.5 text-center text-[0.6875rem] font-medium tracking-[0.14em] text-primary-foreground uppercase sm:text-xs">
      <p>
        Retiro sin cargo en el local
        {condiciones.tarifas.length > 0 ? <span className="hidden sm:inline"> · Envío a domicilio hasta {km} km</span> : null}
        {condiciones.efectivoHabilitado ? <span className="hidden md:inline"> · Mercado Pago o efectivo</span> : null}
      </p>
    </div>
  );
}
