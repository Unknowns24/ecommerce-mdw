import { Badge } from "@/components/ui/badge";

import { ESTADOS_PEDIDO, textoPago, type EstadoPago, type EstadoPedido, type MedioPago } from "../tipos";

// Estado del pedido y del pago, siempre con texto (sección 7).
export function EstadoDelPedido({ estado }: { estado: EstadoPedido }) {
  const variante = estado === "CANCELADO" ? "destructive" : estado === "COMPLETADO" ? "default" : estado === "PENDIENTE_DE_PAGO" ? "secondary" : "outline";
  return <Badge variant={variante}>{ESTADOS_PEDIDO[estado]}</Badge>;
}

export function EstadoDelPago({ estado, medio }: { estado: EstadoPago; medio: MedioPago }) {
  const variante = estado === "RECHAZADO" ? "destructive" : estado === "APROBADO" ? "default" : "secondary";
  return <Badge variant={variante}>{textoPago(estado, medio)}</Badge>;
}
