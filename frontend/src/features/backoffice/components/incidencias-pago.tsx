import { Formulario } from "@/components/formulario";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { formatearFecha, formatearPrecio } from "@/lib/formato";
import type { EstadoFormulario } from "@/lib/acciones";
import type { IncidenciaPago } from "../tipos";

type Resolver = (id: string, previo: EstadoFormulario, datos: FormData) => Promise<EstadoFormulario>;

const MOTIVOS = {
  MONTO_DISTINTO: "El importe cobrado no coincide con el total del pedido.",
  PEDIDO_NO_PENDIENTE: "El pago llegó cuando el pedido ya no esperaba ese cobro. Puede requerir un reintegro externo.",
};

export function IncidenciasPago({ items, resolver }: { items: IncidenciaPago[]; resolver: Resolver }) {
  if (items.length === 0) return null;
  return (
    <section className="grid gap-4" aria-label="Incidencias de pago">
      <h2 className="font-heading text-xl font-medium">Incidencias de pago</h2>
      <p className="text-sm text-muted-foreground">
        Revisá el cobro en Mercado Pago y registrá cómo lo resolviste. El cierre conserva el historial y no cambia el pago ni las unidades del pedido.
        Se muestran hasta 100 incidencias, con las pendientes primero.
      </p>
      {items.map((i) => (
        <Card key={i.id}>
          <CardHeader className="gap-2">
            <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
              <CardTitle className="min-w-0 [overflow-wrap:anywhere]">Pago {i.referenciaExterna}</CardTitle>
              <Badge variant={i.resueltoEn ? "secondary" : "destructive"}>{i.resueltoEn ? "Resuelta" : "Pendiente de resolución"}</Badge>
            </div>
            <CardDescription>Registrada el {formatearFecha(i.creadoEn)}.</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4">
            <p>{MOTIVOS[i.motivo]}</p>
            <dl className="grid gap-3 bg-muted p-4 sm:grid-cols-2">
              <div className="grid gap-1">
                <dt className="text-sm text-muted-foreground">Importe cobrado</dt>
                <dd className="font-medium tabular-nums">{formatearPrecio(i.montoCentavos)}</dd>
              </div>
              <div className="grid gap-1">
                <dt className="text-sm text-muted-foreground">Total del pedido</dt>
                <dd className="font-medium tabular-nums">{formatearPrecio(i.totalPedidoCentavos)}</dd>
              </div>
            </dl>
            {i.resueltoEn ? (
              <div className="grid gap-1 text-sm">
                <p>Resuelta por {i.resueltoPorNombre || "Administración"} el {formatearFecha(i.resueltoEn)}.</p>
                <p className="whitespace-pre-wrap [overflow-wrap:anywhere]">{i.notaResolucion}</p>
              </div>
            ) : (
              <Formulario accion={resolver.bind(null, i.id)} campos={[{
                nombre: "nota", etiqueta: "Cómo se resolvió la incidencia", tipo: "textarea", requerido: true,
                ayuda: "Dejá constancia de la verificación y, si corresponde, del reintegro realizado fuera del sistema.",
              }]} textoEnviar="Registrar resolución" textoEnviando="Guardando resolución…" />
            )}
          </CardContent>
        </Card>
      ))}
    </section>
  );
}
