"use client";

import Link from "next/link";
import { useActionState } from "react";
import { reintentarPagoAction } from "@/app/(tienda)/checkout/acciones";
import { Alert } from "@/components/ui/alert";
import { Button, buttonVariants } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import type { EstadoCheckout, PedidoSinPago } from "../tipos";

// El pedido quedó creado pero el pago no se pudo iniciar. NO está pagado.
// Reemplaza al formulario para que nadie vuelva a «Confirmar» y cree otro pedido.
export function PanelPedidoSinPago({ estado }: { estado: EstadoCheckout & { pedido: PedidoSinPago } }) {
  const pedido = estado.pedido;
  const [reintento, reintentar, pendiente] = useActionState<EstadoCheckout, FormData>(
    reintentarPagoAction.bind(null, pedido),
    { enviado: false },
  );
  // Lo más nuevo que se sabe: el resultado del reintento si ya hubo uno.
  const mensaje = reintento.enviado ? reintento.mensaje : estado.mensaje;

  return (
    <Card>
      <CardContent className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold">Tu pedido #{pedido.numero} quedó creado, pero todavía no está pagado</h2>
        {mensaje ? (
          <Alert variant="destructive" role="alert">
            {mensaje}
          </Alert>
        ) : null}
        <p>
          Reservamos tus productos por 15 minutos desde que hiciste el pedido. Si el pago no se acredita en ese
          plazo, el pedido se cancela y las unidades se liberan.
        </p>
        {pedido.privado ? (
          <div className="flex flex-col gap-1">
            <p className="font-medium">Guardá este enlace privado: es la única forma de consultar tu pedido.</p>
            <p className="text-sm break-all">
              <Link href={pedido.enlace} className="underline">
                {pedido.enlace}
              </Link>
            </p>
          </div>
        ) : null}
        <div className="flex flex-col gap-3 sm:flex-row">
          <form action={reintentar} className="sm:order-2">
            <Button type="submit" disabled={pendiente} className="w-full sm:w-auto">
              {pendiente ? "Iniciando el pago…" : "Reintentar el pago"}
            </Button>
          </form>
          <Link
            href={pedido.enlace}
            className={buttonVariants({ variant: "outline", className: "w-full sm:w-auto" })}
          >
            Ver mi pedido
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}
