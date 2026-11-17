import { Badge } from "@/components/ui/badge";
import { textoEstadoPago, TEXTO_ESTADO_PEDIDO } from "../etiquetas";
import type { EstadoPago, EstadoPedido, MedioPago } from "../tipos";

// Traducen el valor del backend a texto + estilo. El texto siempre está.

export function InsigniaEstadoPedido({ estado }: { estado: EstadoPedido }) {
  const variante =
    estado === "CANCELADO"
      ? "destructive"
      : estado === "PENDIENTE_DE_PAGO"
        ? "secondary"
        : estado === "CONFIRMADO"
          ? "outline"
          : "default";
  return <Badge variant={variante}>{TEXTO_ESTADO_PEDIDO[estado]}</Badge>;
}

export function InsigniaEstadoPago({ estado, medio }: { estado: EstadoPago; medio: MedioPago }) {
  const variante = estado === "RECHAZADO" ? "destructive" : estado === "APROBADO" ? "default" : "secondary";
  return <Badge variant={variante}>{textoEstadoPago(estado, medio)}</Badge>;
}
