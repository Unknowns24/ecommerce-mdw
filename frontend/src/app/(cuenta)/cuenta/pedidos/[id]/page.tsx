import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { buttonVariants } from "@/components/ui/button";
import { ErrorApi } from "@/lib/api";
import { requerirSesion } from "@/features/identidad/sesion";
import { leerPedido } from "@/features/pedidos/api";
import { DetallePedido } from "@/features/pedidos/components/detalle-pedido";

export const metadata: Metadata = { title: "Detalle del pedido" };

// /cuenta/pedidos/[id] (H16): el detalle de un pedido propio. La página exige
// la sesión y la API decide: un pedido ajeno y uno inexistente responden el
// mismo 404, y acá se muestran con el mismo not-found.tsx.
export default async function PaginaPedido({ params }: { params: Promise<{ id: string }> }) {
  await requerirSesion();
  const { id } = await params;

  let pedido;
  try {
    pedido = await leerPedido(id);
  } catch (e) {
    if (e instanceof ErrorApi && e.status === 404) notFound();
    throw e;
  }

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
      <DetallePedido pedido={pedido} />
      <Link
        href="/cuenta/pedidos"
        className={buttonVariants({ variant: "outline", className: "w-full sm:w-auto sm:self-start" })}
      >
        Volver a mis pedidos
      </Link>
    </main>
  );
}
