import { PedidoNoEncontrado } from "@/features/pedidos/components/pedido-no-encontrado";

// El 404 propio del enlace privado (spec, sección 7): ahí no hay nadie con sesión.
export default function SeguimientoNoEncontrado() {
  return <PedidoNoEncontrado accion={{ texto: "Ir al catálogo", href: "/" }} />;
}
