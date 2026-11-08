import type { Metadata } from "next";
import Link from "next/link";
import { Alert } from "@/components/ui/alert";
import { buttonVariants } from "@/components/ui/button";
import { obtenerCuenta } from "@/features/identidad/sesion";
import { leerTokenAcceso } from "@/features/pedidos/acceso";

export const metadata: Metadata = { title: "Verificando tu pago", robots: { index: false } };

// /pago/retorno (H14): a donde vuelve Mercado Pago. La dirección la arma el
// backend como /pago/retorno?estado=aprobado|rechazado|pendiente y NO lleva
// ni el id del pedido ni el token (spec, sección 5).
//
// Esta pantalla NUNCA muestra el pedido como pagado por ese parámetro: el
// navegador puede volver con cualquier `estado` y la confirmación depende solo
// de la verificación del backend (webhook). Por eso ni siquiera lo lee: dice
// que el pago se está verificando y lleva al detalle, que consulta el estado
// real al servidor.
//
// A dónde lleva:
// - invitado: la cookie propia que dejó el checkout → /seguimiento/[token].
// - cliente: /cuenta/pedidos. El id del pedido no vuelve en las back_urls y la
//   spec no define otra forma de saberlo, así que lleva a «Mis pedidos»
//   y no se inventa un id.
export default async function PaginaRetornoPago() {
  const token = await leerTokenAcceso();
  const cuenta = token ? null : await obtenerCuenta();

  const destino = token
    ? { href: `/seguimiento/${token}`, texto: "Ver mi pedido" }
    : cuenta
      ? { href: "/cuenta/pedidos", texto: "Ver mis pedidos" }
      : null;

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
      <h1 className="text-2xl font-semibold">Estamos verificando tu pago</h1>
      <Alert role="status">
        <p className="font-medium">Todavía no podemos decirte si el pago se acreditó.</p>
        <p>
          Mercado Pago nos avisa el resultado y lo confirmamos nosotros. Puede tardar unos minutos. Hasta
          entonces, el pedido figura como pendiente de pago. Para ver cómo está ahora, abrí el detalle de tu
          pedido.
        </p>
      </Alert>
      {destino ? (
        <Link
          href={destino.href}
          className={buttonVariants({ className: "w-full sm:w-auto sm:self-start" })}
        >
          {destino.texto}
        </Link>
      ) : (
        <>
          <p>
            No encontramos un pedido en este navegador. Si compraste como invitado, abrí el enlace privado que
            te mostramos al confirmar el pedido. Si tenés cuenta, ingresá para ver tus pedidos.
          </p>
          <div className="flex flex-col gap-3 sm:flex-row">
            <Link href="/ingresar" className={buttonVariants({ className: "w-full sm:w-auto" })}>
              Ingresar
            </Link>
            <Link href="/" className={buttonVariants({ variant: "outline", className: "w-full sm:w-auto" })}>
              Ir al catálogo
            </Link>
          </div>
        </>
      )}
    </main>
  );
}
