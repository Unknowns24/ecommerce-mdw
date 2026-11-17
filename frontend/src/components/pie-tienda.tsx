import Link from "next/link";

// Pie de la tienda: solo información que la spec respalda (local físico,
// retiro, envío por distancia y medios de pago). Sin dirección, teléfono,
// redes ni políticas: no hay datos cargados y no se inventan.
export function PieTienda() {
  return (
    <footer className="mt-16 bg-muted">
      <div className="mx-auto grid max-w-6xl gap-8 px-4 py-10 sm:grid-cols-3">
        <div className="grid content-start gap-2">
          <p className="font-heading text-xl font-semibold text-primary">BC Importados</p>
          <p className="max-w-xs text-sm text-muted-foreground">Perfumes, cremas, body splash y electrónica importados.</p>
        </div>
        <nav aria-label="Pie de página" className="grid content-start gap-2 text-sm">
          <p className="font-semibold">Tienda</p>
          <Link href="/" className="w-fit underline-offset-4 hover:underline">
            Catálogo
          </Link>
          <Link href="/carrito" className="w-fit underline-offset-4 hover:underline">
            Carrito
          </Link>
          <Link href="/cuenta/pedidos" className="w-fit underline-offset-4 hover:underline">
            Mis pedidos
          </Link>
        </nav>
        <div className="grid content-start gap-2 text-sm">
          <p className="font-semibold">Cómo comprar</p>
          <p className="text-muted-foreground">Podés comprar sin cuenta. Retirás en el local o te lo enviamos dentro de la zona de envío.</p>
          <p className="text-muted-foreground">Pagás con Mercado Pago. Las opciones de pago y entrega disponibles se confirman en el checkout.</p>
        </div>
      </div>
    </footer>
  );
}
