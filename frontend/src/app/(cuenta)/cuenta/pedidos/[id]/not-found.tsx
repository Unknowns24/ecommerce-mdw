import { PedidoNoEncontrado } from "@/features/pedidos/components/pedido-no-encontrado";

export default function PedidoPropioNoEncontrado() {
  return <PedidoNoEncontrado accion={{ texto: "Ver mis pedidos", href: "/cuenta/pedidos" }} />;
}
