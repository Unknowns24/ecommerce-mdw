// Tipos de los pedidos del panel, a partir de docs/api/pedidos.http
// (GET /api/admin/pedidos y /api/admin/pedidos/{id}).

export type EstadoPedido = "PENDIENTE_DE_PAGO" | "CONFIRMADO" | "DESPACHADO" | "COMPLETADO" | "CANCELADO";
export type EstadoPago = "PENDIENTE" | "APROBADO" | "RECHAZADO";
export type MedioPago = "MERCADO_PAGO" | "EFECTIVO";
export type ModoEntrega = "RETIRO" | "ENVIO";

export type PedidoResumen = {
  id: string;
  numero: number;
  modoEntrega: ModoEntrega;
  medioPago: MedioPago;
  estadoPedido: EstadoPedido;
  estadoPago: EstadoPago;
  totalCentavos: number;
  creadoEn: string;
  compradorNombre: string;
  compradorApellido: string;
  incidenciasPendientes: number;
};

export type PaginaPedidos = { items: PedidoResumen[]; total: number; pagina: number; tamano: number };

export type PedidoAdmin = Omit<PedidoResumen, "compradorNombre" | "compradorApellido" | "incidenciasPendientes"> & {
  domicilio?: string;
  distanciaMetros?: number;
  envioCentavos: number;
  subtotalCentavos: number;
  venceEn?: string;
  detalles: { varianteId: string; codigo: string; nombre: string; unidades: number; precioUnitarioCentavos: number; subtotalCentavos: number }[];
  comprador: { nombre: string; apellido: string; correo: string; telefono: string; dni: string };
  usuarioId?: string;
  canceladoEn?: string;
  motivoCancelacion?: string;
};

export type IncidenciaPago = {
  id: string;
  referenciaExterna: string;
  motivo: "MONTO_DISTINTO" | "PEDIDO_NO_PENDIENTE";
  montoCentavos: number;
  totalPedidoCentavos: number;
  creadoEn: string;
  resueltoEn?: string;
  resueltoPorNombre?: string;
  notaResolucion?: string;
};

export const ESTADOS_PEDIDO: Record<EstadoPedido, string> = {
  PENDIENTE_DE_PAGO: "Pendiente de pago",
  CONFIRMADO: "Confirmado",
  DESPACHADO: "Despachado",
  COMPLETADO: "Completado",
  CANCELADO: "Cancelado",
};

export function textoPago(estado: EstadoPago, medio: MedioPago): string {
  if (estado === "APROBADO") return medio === "EFECTIVO" ? "Cobrado en efectivo" : "Pagado con Mercado Pago";
  if (estado === "RECHAZADO") return "Pago rechazado";
  return medio === "EFECTIVO" ? "Efectivo, sin cobrar" : "Pago pendiente en Mercado Pago";
}

export const TEXTO_ENTREGA: Record<ModoEntrega, string> = { RETIRO: "Retiro en el local", ENVIO: "Envío a domicilio" };
