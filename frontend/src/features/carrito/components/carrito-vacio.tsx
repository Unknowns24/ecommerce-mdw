import { SinDatos } from "@/components/sin-datos";

// Estado «vacío»: dice qué pasa y ofrece la acción que lo resuelve.
export function CarritoVacio() {
  return (
    <SinDatos
      titulo="Tu carrito está vacío"
      descripcion="Todavía no agregaste productos. Mirá el catálogo y elegí lo que necesites."
      accion={{ texto: "Ir al catálogo", href: "/" }}
    />
  );
}
