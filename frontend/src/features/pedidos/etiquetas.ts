import type { EstadoPago, EstadoPedido, MedioPago, ModoEntrega } from "./tipos";

// Del valor del backend al texto que ve la persona. Cada estado se dice con
// palabras: el color es un refuerzo, nunca el único mensaje (spec, sección 7).

export const TEXTO_ESTADO_PEDIDO: Record<EstadoPedido, string> = {
  PENDIENTE_DE_PAGO: "Pendiente de pago",
  CONFIRMADO: "Confirmado",
  DESPACHADO: "Despachado",
  COMPLETADO: "Completado",
  CANCELADO: "Cancelado",
};

export const TEXTO_ESTADO_PAGO: Record<EstadoPago, string> = {
  PENDIENTE: "Pendiente",
  APROBADO: "Aprobado",
  RECHAZADO: "Rechazado",
};

/** El pago en efectivo pendiente se cobra en el local (H15): no es un pago «trabado». */
export function textoEstadoPago(estado: EstadoPago, medio: MedioPago): string {
  if (estado === "PENDIENTE" && medio === "EFECTIVO") return "Pendiente de cobro en el local";
  return TEXTO_ESTADO_PAGO[estado];
}

export const TEXTO_MEDIO_PAGO: Record<MedioPago, string> = {
  MERCADO_PAGO: "Mercado Pago",
  EFECTIVO: "Efectivo al retirar",
};

export const TEXTO_ENTREGA: Record<ModoEntrega, string> = {
  RETIRO: "Retiro en el local",
  ENVIO: "Envío a domicilio",
};
