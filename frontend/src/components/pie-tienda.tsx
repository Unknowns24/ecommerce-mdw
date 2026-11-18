import { StoreIcon, TruckIcon, UserRoundIcon, WalletIcon } from "lucide-react";
import Link from "next/link";
import { unstable_rethrow } from "next/navigation";
import type { ReactNode } from "react";

import { listarCategorias, listarMarcas, obtenerCondicionesEntrega } from "@/features/catalogo/api";

// Pie de la tienda: solo información que la spec y la API respaldan
// (condiciones de entrega y pago, categorías, marcas, cuenta y cómo comprar).
// Sin dirección, teléfono, redes, newsletter ni políticas: no hay datos
// cargados y no se inventan.
export async function PieTienda() {
  // Si la API falla, el pie queda sin esa parte pero la página no se cae.
  function sinDato<T>(vacio: T) {
    return (e: unknown) => {
      unstable_rethrow(e);
      console.error("pie de la tienda", e);
      return vacio;
    };
  }
  const [categorias, marcas, condiciones] = await Promise.all([
    listarCategorias().catch(sinDato([])),
    listarMarcas().catch(sinDato([])),
    obtenerCondicionesEntrega().catch(sinDato(null)),
  ]);
  const todas = categorias.flatMap((c) => [c, ...(c.hijas ?? [])]);

  // La franja de condiciones: cada una sale de la configuración de la tienda
  // o de la spec (la compra como invitado es el caso principal).
  const km = condiciones ? (condiciones.distanciaMaximaMetros / 1000).toLocaleString("es-AR", { maximumFractionDigits: 1 }) : "";
  const condicionesTienda = condiciones
    ? [
        { icono: StoreIcon, texto: "Retiro sin cargo en el local" },
        ...(condiciones.tarifas.length > 0 ? [{ icono: TruckIcon, texto: `Envío a domicilio hasta ${km} km` }] : []),
        { icono: WalletIcon, texto: condiciones.efectivoHabilitado ? "Mercado Pago o efectivo" : "Pagás con Mercado Pago" },
        { icono: UserRoundIcon, texto: "Comprás sin crear una cuenta" },
      ]
    : [];

  return (
    <footer className="mt-24 border-t bg-background">
      {condicionesTienda.length > 0 ? (
        <ul
          aria-label="Condiciones de compra"
          className="mx-auto grid max-w-6xl grid-cols-2 gap-x-4 gap-y-6 border-b px-4 py-8 md:flex md:justify-center md:gap-14"
        >
          {condicionesTienda.map(({ icono: Icono, texto }) => (
            <li key={texto} className="flex items-center gap-3 text-[0.8125rem] text-muted-foreground">
              <Icono className="size-6 shrink-0 stroke-[1.25] text-foreground" aria-hidden="true" />
              {texto}
            </li>
          ))}
        </ul>
      ) : null}

      <div className="mx-auto grid max-w-6xl gap-12 px-4 pt-12 pb-8">
        <nav aria-label="Pie de página" className="grid grid-cols-2 gap-x-6 gap-y-10 md:grid-cols-4">
          {todas.length > 0 ? (
            <Columna titulo="Categorías">
              {todas.map((c) => (
                <Enlace key={c.id} href={`/?categoria=${c.id}`}>
                  {c.nombre}
                </Enlace>
              ))}
            </Columna>
          ) : null}
          {marcas.length > 0 ? (
            <Columna titulo="Marcas">
              {marcas.map((m) => (
                <Enlace key={m.id} href={`/?marca=${m.id}`}>
                  {m.nombre}
                </Enlace>
              ))}
            </Columna>
          ) : null}
          <Columna titulo="Tu cuenta">
            <Enlace href="/carrito">Carrito</Enlace>
            <Enlace href="/cuenta/pedidos">Mis pedidos</Enlace>
            <Enlace href="/registro">Crear cuenta</Enlace>
          </Columna>
          <div className="col-span-2 grid content-start gap-4 text-sm md:col-span-1">
            <h2 className={claseTitulo}>Cómo comprar</h2>
            <p className="text-muted-foreground">Podés comprar sin cuenta. Retirás en el local o te lo enviamos dentro de la zona de envío.</p>
            <p className="text-muted-foreground">Pagás con Mercado Pago. Las opciones de pago y entrega se confirman en el checkout.</p>
          </div>
        </nav>

        <div className="flex flex-col gap-2 border-t pt-6 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-base font-light tracking-[0.28em] uppercase">BC Importados</p>
          <p className="text-xs text-muted-foreground">© BC Importados. Perfumes, cremas, body splash y electrónica importados.</p>
        </div>
      </div>
    </footer>
  );
}

const claseTitulo = "text-xs font-medium tracking-[0.18em] uppercase";

function Columna({ titulo, children }: { titulo: string; children: ReactNode }) {
  return (
    <div className="grid content-start gap-4 text-sm">
      <h2 className={claseTitulo}>{titulo}</h2>
      <ul className="grid gap-2.5">{children}</ul>
    </div>
  );
}

function Enlace({ href, children }: { href: string; children: ReactNode }) {
  return (
    <li>
      <Link href={href} className="text-muted-foreground underline-offset-4 hover:text-foreground hover:underline">
        {children}
      </Link>
    </li>
  );
}
