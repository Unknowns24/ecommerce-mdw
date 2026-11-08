import { Alert } from "@/components/ui/alert";
import { Card, CardContent } from "@/components/ui/card";
import { formatearFecha, formatearPrecio } from "@/lib/formato";
import { TEXTO_ENTREGA, TEXTO_MEDIO_PAGO } from "../etiquetas";
import type { PedidoVisible } from "../tipos";
import { InsigniaEstadoPago, InsigniaEstadoPedido } from "./estado-pedido";

// Detalle de un pedido (H16), igual para el cliente (/cuenta/pedidos/[id]) y
// para el invitado (/seguimiento/[token]). Server Component: solo dibuja lo que
// devolvió la API, con los precios de cuando se hizo el pedido.
export function DetallePedido({ pedido }: { pedido: PedidoVisible }) {
  const esperaPagoOnline = pedido.estadoPedido === "PENDIENTE_DE_PAGO" && pedido.medioPago === "MERCADO_PAGO";

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">Pedido #{pedido.numero}</h1>
        <p className="text-sm text-muted-foreground">Hecho el {formatearFecha(pedido.creadoEn)}</p>
      </div>

      <Card>
        <CardContent className="flex flex-col gap-3">
          <h2 className="text-lg font-semibold">Estado</h2>
          <dl className="flex flex-col gap-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <dt>Estado del pedido</dt>
              <dd>
                <InsigniaEstadoPedido estado={pedido.estadoPedido} />
              </dd>
            </div>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <dt>Estado del pago</dt>
              <dd>
                <InsigniaEstadoPago estado={pedido.estadoPago} medio={pedido.medioPago} />
              </dd>
            </div>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <dt>Medio de pago</dt>
              <dd>{TEXTO_MEDIO_PAGO[pedido.medioPago]}</dd>
            </div>
          </dl>
          {esperaPagoOnline ? (
            <Alert role="status">
              <p className="font-medium">Tu pago todavía no está acreditado.</p>
              <p>
                Reservamos tus productos
                {pedido.venceEn ? ` hasta el ${formatearFecha(pedido.venceEn)}` : " por 15 minutos"}. Si el pago no
                se acredita en ese plazo, el pedido se cancela.
              </p>
            </Alert>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardContent className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold">Productos</h2>
          <ul className="flex flex-col gap-3">
            {pedido.detalles.map((d) => (
              <li key={d.codigo} className="flex flex-col gap-1">
                <p className="font-medium break-words">{d.nombre}</p>
                <p className="text-sm text-muted-foreground">Código: {d.codigo}</p>
                <p className="flex justify-between gap-2 text-sm">
                  <span>
                    {d.unidades} × {formatearPrecio(d.precioUnitarioCentavos)}
                  </span>
                  <span>{formatearPrecio(d.subtotalCentavos)}</span>
                </p>
              </li>
            ))}
          </ul>
          <dl className="flex flex-col gap-1">
            <div className="flex justify-between gap-2">
              <dt>Productos</dt>
              <dd>{formatearPrecio(pedido.subtotalCentavos)}</dd>
            </div>
            <div className="flex justify-between gap-2">
              <dt>Envío</dt>
              <dd>{pedido.envioCentavos > 0 ? formatearPrecio(pedido.envioCentavos) : "Sin cargo"}</dd>
            </div>
            <div className="flex justify-between gap-2 text-lg font-semibold">
              <dt>Total</dt>
              <dd>{formatearPrecio(pedido.totalCentavos)}</dd>
            </div>
          </dl>
          <p className="text-sm text-muted-foreground">
            Son los precios de cuando hiciste el pedido; no cambian si el catálogo se actualiza.
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="flex flex-col gap-2">
          <h2 className="text-lg font-semibold">Entrega</h2>
          <p>{TEXTO_ENTREGA[pedido.modoEntrega]}</p>
          {pedido.modoEntrega === "ENVIO" ? (
            <>
              {pedido.domicilio ? <p className="break-words">Domicilio: {pedido.domicilio}</p> : null}
              {pedido.distanciaMetros !== undefined ? <p>Distancia: {pedido.distanciaMetros} m</p> : null}
            </>
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}
