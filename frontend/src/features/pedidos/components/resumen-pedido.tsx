import { Card, CardContent } from "@/components/ui/card";
import type { VistaCarrito } from "@/features/carrito/tipos";
import { formatearPrecio } from "@/lib/formato";
import type { EstimacionEnvio } from "../envio";

// Resumen del carrito dentro del checkout. Todo es estimado: el total que vale
// lo calcula el servidor al confirmar (spec, H13), con catálogo y stock vigentes.
export function ResumenPedido({
  vista,
  envio,
}: {
  vista: VistaCarrito;
  /** null = retiro (sin cargo); si no, lo que se pudo estimar del envío. */
  envio: EstimacionEnvio | null;
}) {
  const costoEnvio = envio?.tipo === "costo" ? envio.centavos : 0;
  return (
    <Card>
      <CardContent className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold">Resumen</h2>
        <ul className="flex flex-col gap-3">
          {vista.lineas.map((l) => (
            <li key={l.id} className="flex flex-col gap-1">
              <p className="font-medium break-words">{l.nombre}</p>
              {l.disponible ? (
                <p className="flex justify-between gap-2 text-sm">
                  <span>
                    {l.unidades} × {formatearPrecio(l.precioUnitarioCentavos)}
                  </span>
                  <span>{formatearPrecio(l.subtotalCentavos)}</span>
                </p>
              ) : (
                <p className="text-sm font-medium text-destructive">
                  No disponible: {l.unidades} u. Si confirmás, el pedido va a ser rechazado; quitalo desde el
                  carrito.
                </p>
              )}
            </li>
          ))}
        </ul>
        <dl className="flex flex-col gap-1">
          <div className="flex justify-between gap-2">
            <dt>Productos</dt>
            <dd>{formatearPrecio(vista.totalEstimadoCentavos)}</dd>
          </div>
          <div className="flex justify-between gap-2">
            <dt>Envío</dt>
            <dd>
              {envio === null
                ? "Sin cargo (retiro en el local)"
                : envio.tipo === "costo"
                  ? `${formatearPrecio(envio.centavos)} (estimado)`
                  : "A calcular al confirmar"}
            </dd>
          </div>
          <div className="flex justify-between gap-2 text-lg font-semibold">
            <dt>Total estimado</dt>
            <dd>{formatearPrecio(vista.totalEstimadoCentavos + costoEnvio)}</dd>
          </div>
        </dl>
        <p className="text-sm text-muted-foreground">
          El precio y las unidades se verifican de nuevo al confirmar. Si algo cambió, te lo avisamos antes de
          cobrar.
        </p>
      </CardContent>
    </Card>
  );
}
