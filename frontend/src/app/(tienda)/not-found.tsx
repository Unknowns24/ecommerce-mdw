import { SinDatos } from "@/components/sin-datos";

// 404 de la tienda, dentro de su layout: quien llegó por un enlace viejo
// tiene a dónde ir. «No existe» y «no es tuyo» dicen lo mismo.
export default function NoEncontrado() {
  return (
    <SinDatos
      titulo="No encontramos esta página"
      descripcion="Puede que el enlace esté mal escrito o que el producto ya no esté a la venta."
      accion={{ texto: "Ir al catálogo", href: "/" }}
    />
  );
}
