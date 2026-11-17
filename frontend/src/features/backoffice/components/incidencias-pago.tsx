import { Formulario } from "@/components/formulario";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
      <h2 className="text-xl font-semibold">Incidencias de pago</h2>
      <p className="text-sm text-muted-foreground">
        Revisá el cobro en Mercado Pago y registrá cómo lo resolviste. El cierre conserva el historial y no cambia el pago ni las unidades del pedido.
        Se muestran hasta 100 incidencias, con las pendientes primero.
      </p>
      {items.map((i) => (
        <Card key={i.id}>
          <CardHeader>
            <CardTitle>Pago {i.referenciaExterna}</CardTitle>
            <Badge variant={i.resueltoEn ? "secondary" : "destructive"}>{i.resueltoEn ? "Resuelta" : "Pendiente de resolución"}</Badge>
          </CardHeader>
          <CardContent className="grid gap-4">
            <p>{MOTIVOS[i.motivo]}</p>
            <p className="text-sm">Cobrado: {formatearPrecio(i.montoCentavos)}. Total del pedido: {formatearPrecio(i.totalPedidoCentavos)}.</p>
            <p className="text-sm text-muted-foreground">Registrada el {formatearFecha(i.creadoEn)}.</p>
            {i.resueltoEn ? (
              <div className="grid gap-1 text-sm">
                <p>Resuelta por {i.resueltoPorNombre || "Administración"} el {formatearFecha(i.resueltoEn)}.</p>
                <p className="whitespace-pre-wrap">{i.notaResolucion}</p>
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
