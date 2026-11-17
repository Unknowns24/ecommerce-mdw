import { EsqueletoPedido } from "@/features/pedidos/components/esqueleto-pedido";

// Mismo contenedor que page.tsx para que no salte el diseño cuando llegan los datos.
export default function CargandoPedido() {
  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
      <EsqueletoPedido />
    </div>
  );
}
