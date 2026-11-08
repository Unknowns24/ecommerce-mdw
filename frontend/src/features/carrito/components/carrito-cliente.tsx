import type { VistaCarrito } from "../tipos";
import { CarritoVacio } from "./carrito-vacio";
import { ListaCliente } from "./lista-cliente";
import { AvisoCarrito, ResumenCarrito } from "./resumen-carrito";

// Server Component: dibuja el carrito que devolvió la API para quien tiene sesión.
export function CarritoCliente({ vista }: { vista: VistaCarrito }) {
  if (vista.lineas.length === 0) return <CarritoVacio />;
  return (
    <div className="flex flex-col gap-4">
      <AvisoCarrito invitado={false} />
      <ListaCliente lineas={vista.lineas} />
      <ResumenCarrito totalCentavos={vista.totalEstimadoCentavos} />
    </div>
  );
}
