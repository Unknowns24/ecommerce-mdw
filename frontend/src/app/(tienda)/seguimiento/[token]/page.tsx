import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { ErrorApi } from "@/lib/api";
import { leerPedidoPublico } from "@/features/pedidos/api";
import { DetallePedido } from "@/features/pedidos/components/detalle-pedido";

// El token es un secreto: este enlace no se indexa y el navegador no lo manda
// como Referer al abrir otro sitio.
export const metadata: Metadata = {
  title: "Seguimiento de tu pedido",
  robots: { index: false, follow: false },
  referrer: "no-referrer",
};

// /seguimiento/[token] (H16): el pedido del invitado, por su enlace privado.
// Sin sesión: la API decide por el token. Un token inválido y uno inexistente
// son el mismo 404 (not-found.tsx de esta carpeta).
export default async function PaginaSeguimiento({ params }: { params: Promise<{ token: string }> }) {
  const { token } = await params;

  let pedido;
  try {
    pedido = await leerPedidoPublico(token);
  } catch (e) {
    if (e instanceof ErrorApi && e.status === 404) notFound();
    throw e;
  }

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
      <DetallePedido pedido={pedido} />
      <p className="text-sm text-muted-foreground">
        Este enlace es privado: cualquiera que lo tenga puede ver este pedido. Guardalo para volver a
        consultarlo.
      </p>
    </main>
  );
}
