import "server-only";
import { api, ErrorApi } from "@/lib/api";
import type {
  ConfiguracionTienda,
  CuerpoCrearPedido,
  PedidoCreado,
  RespuestaPago,
  Tarifa,
} from "./tipos";

// Llamadas a la API de Go de pedidos y pago. Solo desde el servidor (Server
// Components y Server Actions de este módulo).

/**
 * GET /api/tienda/configuracion (público). 404 = sin configuración cargada:
 * el checkout lo entiende como «el envío no está disponible» (nicolas.md, N1).
 * Cualquier otro error sube a error.tsx.
 */
export async function leerConfiguracion(): Promise<ConfiguracionTienda | null> {
  try {
    const c = await api<Partial<ConfiguracionTienda>>("/api/tienda/configuracion");
    const tarifas: Tarifa[] = Array.isArray(c.tarifas) ? c.tarifas : [];
    return {
      efectivoHabilitado: c.efectivoHabilitado === true,
      distanciaMaximaMetros: typeof c.distanciaMaximaMetros === "number" ? c.distanciaMaximaMetros : 0,
      tarifas,
    };
  } catch (e) {
    if (e instanceof ErrorApi && e.status === 404) return null;
    throw e;
  }
}

/** POST /api/pedidos. La identidad sale de la sesión (la API la lee del contexto). */
export function crearPedido(cuerpo: CuerpoCrearPedido): Promise<PedidoCreado> {
  return api<PedidoCreado>("/api/pedidos", { metodo: "POST", cuerpo });
}

/**
 * POST /api/pedidos/{id}/pago. Paga su comprador: la sesión (que `api` reenvía)
 * o el invitado con su token en el header `X-Token-Acceso` (nunca en la URL).
 */
export function iniciarPago(pedidoId: string, tokenAcceso?: string): Promise<RespuestaPago> {
  return api<RespuestaPago>(`/api/pedidos/${encodeURIComponent(pedidoId)}/pago`, {
    metodo: "POST",
    headers: tokenAcceso ? { "X-Token-Acceso": tokenAcceso } : undefined,
  });
}
