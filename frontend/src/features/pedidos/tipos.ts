import type { EstadoFormulario } from "@/lib/acciones";

// Formas derivadas de docs/api/pedidos.http, docs/api/pagos.http y de los DTO
// de Go (backend/internal/adapters/rest/dtos/pedidos). Los importes viajan en
// centavos (ADR-004).

export type ModoEntrega = "RETIRO" | "ENVIO";
export type MedioPago = "MERCADO_PAGO" | "EFECTIVO";
export type EstadoPedido =
  | "PENDIENTE_DE_PAGO"
  | "CONFIRMADO"
  | "DESPACHADO"
  | "COMPLETADO"
  | "CANCELADO";
export type EstadoPago = "PENDIENTE" | "APROBADO" | "RECHAZADO";

// --- GET /api/tienda/configuracion (público) -------------------------------

/** Rango semiabierto [desdeMetros, hastaMetros) (spec, sección 9). */
export type Tarifa = { desdeMetros: number; hastaMetros: number; costoCentavos: number };

export type ConfiguracionTienda = {
  efectivoHabilitado: boolean;
  distanciaMaximaMetros: number;
  tarifas: Tarifa[];
};

// --- POST /api/pedidos ------------------------------------------------------

/** Lo único que viaja de cada línea: nunca precios (los pone el servidor). */
export type ItemPedido = { varianteId: string; unidades: number };

export type CuerpoCrearPedido = {
  items: ItemPedido[];
  comprador: { nombre: string; apellido: string; correo: string; telefono: string; dni: string };
  modoEntrega: ModoEntrega;
  domicilio?: string;
  distanciaMetros?: number;
  medioPago: MedioPago;
};

/** Respuesta 201 de POST /api/pedidos: solo los campos que el checkout usa. */
export type PedidoCreado = {
  id: string;
  numero: number;
  estadoPedido: EstadoPedido;
  estadoPago: EstadoPago;
  totalCentavos: number;
  tokenAcceso: string;
};

// --- POST /api/pedidos/{id}/pago -------------------------------------------

export type RespuestaPago = { initPoint: string };

// --- Estado del formulario de checkout -------------------------------------

/**
 * Pedido que ya quedó creado aunque el pago no haya podido iniciarse. Mientras
 * esté presente la pantalla no vuelve a ofrecer «Confirmar» (crearía otro
 * pedido): ofrece reintentar el pago de ESTE pedido y verlo.
 * `enlace` es la ruta de consulta: /seguimiento/[token] (invitado, el enlace
 * privado) o /cuenta/pedidos/[id] (cliente).
 */
export type PedidoSinPago = { id: string; numero: number; enlace: string; privado: boolean };

export type EstadoCheckout = EstadoFormulario & { pedido?: PedidoSinPago };

// --- Consulta: GET /api/pedidos, /api/pedidos/{id}, /api/pedidos/publico/{token} ---

export type LineaPedido = {
  codigo: string;
  nombre: string;
  unidades: number;
  /** Precio histórico: el de cuando se creó el pedido, no el del catálogo de hoy. */
  precioUnitarioCentavos: number;
  subtotalCentavos: number;
};

/** Lo que muestran las dos vistas (propia y pública, que es la reducida). */
export type PedidoVisible = {
  numero: number;
  modoEntrega: ModoEntrega;
  domicilio?: string;
  distanciaMetros?: number;
  envioCentavos: number;
  subtotalCentavos: number;
  totalCentavos: number;
  estadoPedido: EstadoPedido;
  estadoPago: EstadoPago;
  medioPago: MedioPago;
  /** Vencimiento de la reserva de Mercado Pago (ISO). Efectivo no vence. */
  venceEn?: string;
  creadoEn: string;
  detalles: LineaPedido[];
};

/** GET /api/pedidos/{id}: la vista propia (la pública no trae id). */
export type PedidoPropio = PedidoVisible & { id: string };

/** Fila del listado «Mis pedidos». */
export type PedidoResumen = {
  id: string;
  numero: number;
  modoEntrega: ModoEntrega;
  medioPago: MedioPago;
  estadoPedido: EstadoPedido;
  estadoPago: EstadoPago;
  totalCentavos: number;
  creadoEn: string;
};

export type PaginaPedidos = {
  items: PedidoResumen[];
  total: number;
  pagina: number;
  tamano: number;
};
