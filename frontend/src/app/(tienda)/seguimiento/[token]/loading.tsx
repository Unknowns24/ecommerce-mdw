import { EsqueletoPedido } from "@/features/pedidos/components/esqueleto-pedido";

// Mismo contenedor que page.tsx para que no salte el diseño cuando llegan los datos.
export default function CargandoSeguimiento() {
  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
      <EsqueletoPedido />
    </main>
  );
}
