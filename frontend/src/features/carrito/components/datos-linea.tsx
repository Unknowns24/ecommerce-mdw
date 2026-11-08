import { formatearPrecio } from "@/lib/formato";
import type { LineaCarrito } from "../tipos";

// Lo que se ve de una línea, igual para cliente e invitado: nombre, código,
// precio unitario y, si la variante ya no se puede comprar, un aviso con texto.
export function DatosLinea({ linea }: { linea: LineaCarrito }) {
  return (
    <div className="flex flex-col gap-1">
      <p className="font-medium break-words">{linea.nombre}</p>
      {linea.codigo ? (
        <p className="text-sm text-muted-foreground">Código: {linea.codigo}</p>
      ) : null}
      {linea.disponible ? (
        <p className="text-sm">Precio por unidad: {formatearPrecio(linea.precioUnitarioCentavos)}</p>
      ) : (
        <p className="text-sm font-medium text-destructive">
          No disponible: este producto ya no se puede comprar. Quitalo del carrito.
        </p>
      )}
    </div>
  );
}

export function SubtotalLinea({ linea }: { linea: LineaCarrito }) {
  return linea.disponible ? (
    <p className="font-medium">Subtotal: {formatearPrecio(linea.subtotalCentavos)}</p>
  ) : (
    <p className="text-sm text-muted-foreground">Sin subtotal ({linea.unidades} u.)</p>
  );
}
