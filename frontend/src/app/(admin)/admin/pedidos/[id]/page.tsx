import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { BotonAccion } from "@/components/formulario";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { listarIncidencias, obtenerPedido } from "@/features/backoffice/api";
import { IncidenciasPago } from "@/features/backoffice/components/incidencias-pago";
import { EstadoDelPago, EstadoDelPedido } from "@/features/backoffice/components/estados";
import { FormularioCancelar } from "@/features/backoffice/components/formulario-cancelar";
import { TEXTO_ENTREGA, type PedidoAdmin } from "@/features/backoffice/tipos";
import { requerirPermiso } from "@/features/identidad/sesion";
import { ErrorApi } from "@/lib/api";
import { formatearFecha, formatearPrecio } from "@/lib/formato";

import { cancelarAction, completarEfectivoAction, despacharAction, entregarAction, resolverIncidenciaAction } from "./acciones";

export const metadata: Metadata = { title: "Pedido" };

// Qué acción ofrece cada estado (spec, sección 6). Es cortesía: el backend
// valida cada transición igual.
function acciones(p: PedidoAdmin) {
  const confirmado = p.estadoPedido === "CONFIRMADO";
  const pagado = p.estadoPago === "APROBADO";
  return {
    despachar: p.modoEntrega === "ENVIO" && confirmado && pagado,
    entregar: (p.modoEntrega === "ENVIO" && p.estadoPedido === "DESPACHADO") || (p.modoEntrega === "RETIRO" && p.medioPago === "MERCADO_PAGO" && confirmado && pagado),
    completarEfectivo: p.modoEntrega === "RETIRO" && p.medioPago === "EFECTIVO" && confirmado && !pagado,
    cancelar: p.estadoPedido === "PENDIENTE_DE_PAGO" || confirmado,
  };
}

// Detalle de un pedido del panel (H15, H17, H18).
export default async function PedidoPanel({ params }: PageProps<"/admin/pedidos/[id]">) {
  await requerirPermiso("pedidos.gestionar");
  const { id } = await params;
  const [pedido, incidencias] = await Promise.all([obtenerPedido(id), listarIncidencias(id)]).catch((e: unknown) => {
    if (e instanceof ErrorApi && (e.status === 400 || e.status === 404)) notFound();
    throw e;
  });
  const puede = acciones(pedido);
  const hayAcciones = puede.despachar || puede.entregar || puede.completarEfectivo;

  return (
    <section className="grid gap-6">
      <div className="grid gap-2">
        <Link href="/admin/pedidos" className="text-sm underline underline-offset-4">
          Volver a pedidos
        </Link>
        <h1 className="text-2xl font-semibold">Pedido #{pedido.numero}</h1>
        <div className="flex flex-wrap items-center gap-2">
          <EstadoDelPedido estado={pedido.estadoPedido} />
          <EstadoDelPago estado={pedido.estadoPago} medio={pedido.medioPago} />
          <span className="text-sm text-muted-foreground">Creado el {formatearFecha(pedido.creadoEn)}</span>
        </div>
      </div>

      <IncidenciasPago items={incidencias} resolver={resolverIncidenciaAction.bind(null, id)} />

      {hayAcciones ? (
        <Card>
          <CardHeader>
            <CardTitle>Próximo paso</CardTitle>
          </CardHeader>
          <CardContent className="grid max-w-sm gap-3">
            {puede.despachar ? <BotonAccion accion={despacharAction.bind(null, pedido.id)} texto="Marcar despachado" textoEnviando="Registrando el despacho…" variante="default" /> : null}
            {puede.entregar ? <BotonAccion accion={entregarAction.bind(null, pedido.id)} texto="Marcar entregado" textoEnviando="Registrando la entrega…" variante="default" /> : null}
            {puede.completarEfectivo ? (
              <>
                <BotonAccion accion={completarEfectivoAction.bind(null, pedido.id)} texto="Completar pedido" textoEnviando="Registrando cobro y entrega…" variante="default" />
                <p className="text-sm text-muted-foreground">Registra juntos el cobro en efectivo y la entrega en el local.</p>
              </>
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Comprador</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-1 text-sm">
            <p className="font-medium">
              {pedido.comprador.nombre} {pedido.comprador.apellido}
            </p>
            <p>{pedido.comprador.correo}</p>
            <p>Teléfono {pedido.comprador.telefono}</p>
            <p>DNI {pedido.comprador.dni}</p>
            <p className="text-muted-foreground">{pedido.usuarioId ? "Cliente registrado" : "Compró como invitado"}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Entrega y pago</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-1 text-sm">
            <p className="font-medium">{TEXTO_ENTREGA[pedido.modoEntrega]}</p>
            {pedido.domicilio ? <p>{pedido.domicilio}</p> : null}
            {pedido.distanciaMetros !== undefined ? <p>A {pedido.distanciaMetros} metros del local</p> : null}
            <p>{pedido.medioPago === "EFECTIVO" ? "Efectivo al retirar" : "Mercado Pago"}</p>
            {pedido.venceEn && pedido.estadoPedido === "PENDIENTE_DE_PAGO" ? <p>La reserva vence el {formatearFecha(pedido.venceEn)}</p> : null}
            {pedido.canceladoEn ? (
              <p>
                Cancelado el {formatearFecha(pedido.canceladoEn)}. Motivo: {pedido.motivoCancelacion}
              </p>
            ) : null}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Productos</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Producto</TableHead>
                  <TableHead className="text-right">Unidades</TableHead>
                  <TableHead className="text-right">Precio</TableHead>
                  <TableHead className="text-right">Subtotal</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {pedido.detalles.map((d) => (
                  <TableRow key={d.varianteId}>
                    <TableCell>
                      {d.nombre} <span className="text-muted-foreground">({d.codigo})</span>
                    </TableCell>
                    <TableCell className="text-right">{d.unidades}</TableCell>
                    <TableCell className="text-right">{formatearPrecio(d.precioUnitarioCentavos)}</TableCell>
                    <TableCell className="text-right">{formatearPrecio(d.subtotalCentavos)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
          <dl className="grid justify-end gap-1 text-sm">
            <div className="flex justify-between gap-8">
              <dt>Subtotal</dt>
              <dd>{formatearPrecio(pedido.subtotalCentavos)}</dd>
            </div>
            <div className="flex justify-between gap-8">
              <dt>Envío</dt>
              <dd>{formatearPrecio(pedido.envioCentavos)}</dd>
            </div>
            <div className="flex justify-between gap-8 font-semibold">
              <dt>Total</dt>
              <dd>{formatearPrecio(pedido.totalCentavos)}</dd>
            </div>
          </dl>
        </CardContent>
      </Card>

      {/* Queda montada después de cancelar, para mostrar el aviso que devolvió la API. */}
      {puede.cancelar || pedido.estadoPedido === "CANCELADO" ? (
        <Card>
          <CardHeader>
            <CardTitle>{pedido.estadoPedido === "CANCELADO" ? "Cancelación" : "Cancelar el pedido"}</CardTitle>
          </CardHeader>
          <CardContent className="max-w-lg">
            <FormularioCancelar
              accion={cancelarAction.bind(null, pedido.id)}
              pagado={pedido.estadoPago === "APROBADO"}
              cancelado={pedido.estadoPedido === "CANCELADO"}
            />
          </CardContent>
        </Card>
      ) : null}
    </section>
  );
}
