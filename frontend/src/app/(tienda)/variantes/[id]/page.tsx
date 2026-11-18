import type { Metadata } from "next";
import Link from "next/link";
import { notFound, unstable_rethrow } from "next/navigation";
import { Suspense, type ReactNode } from "react";

import { BarraAccionFija } from "@/components/barra-accion-fija";
import { BotonAgregarAlCarrito } from "@/features/carrito/components/boton-agregar";
import { listarVariantes, obtenerCondicionesEntrega, obtenerVariante, type CondicionesEntrega } from "@/features/catalogo/api";
import { SinStock } from "@/features/catalogo/components/estado-stock";
import { GaleriaVariante } from "@/features/catalogo/components/galeria-variante";
import { Migas } from "@/features/catalogo/components/migas";
import { TarjetaVariante } from "@/features/catalogo/components/tarjeta-variante";
import { ErrorApi } from "@/lib/api";
import { formatearPrecio } from "@/lib/formato";

export const metadata: Metadata = { title: "Producto" };

// Sección plegable nativa: funciona con teclado y lector de pantalla sin JS.
function Seccion({ titulo, abierta, children }: { titulo: string; abierta?: boolean; children: ReactNode }) {
  return (
    <details open={abierta} className="group border-t py-4">
      <summary className="flex cursor-pointer list-none items-center justify-between">
        {/* El título sigue siendo un encabezado: se navega por encabezados con lector de pantalla. */}
        <h2 className="font-heading text-lg font-semibold">{titulo}</h2>
        <span aria-hidden="true" className="text-2xl leading-none font-light text-foreground transition-transform group-open:rotate-45">
          +
        </span>
      </summary>
      <div className="pt-3">{children}</div>
    </details>
  );
}

// Solo datos de la configuración real de la tienda (spec, H13 y H15).
function EntregaYPago({ condiciones }: { condiciones: CondicionesEntrega }) {
  const tarifas = [...condiciones.tarifas].sort((a, b) => a.costoCentavos - b.costoCentavos);
  const km = (m: number) => `${(m / 1000).toLocaleString("es-AR", { maximumFractionDigits: 1 })} km`;
  return (
    <ul className="grid max-w-prose gap-2 text-sm">
      <li>Retiro en el local, sin cargo.</li>
      {tarifas.length > 0 ? (
        <li>
          Envío a domicilio hasta {km(condiciones.distanciaMaximaMetros)} del local, desde {formatearPrecio(tarifas[0].costoCentavos)} según la distancia.
          El costo final se calcula en el checkout.
        </li>
      ) : null}
      <li>
        Pagás con Mercado Pago{condiciones.efectivoHabilitado ? ", o en efectivo si retirás en el local" : ""}.
      </li>
      <li>Las unidades se reservan recién al confirmar el pedido.</li>
    </ul>
  );
}

// Más productos de la misma categoría (sin la variante que se está viendo).
// Llega después de la ficha (Suspense): no la demora.
async function Relacionados({ categoria, excluir }: { categoria: { id: string; nombre: string }; excluir: string[] }) {
  const pagina = await listarVariantes({ q: "", categoria: categoria.id, marca: "", orden: "nombre_asc", pagina: 1 }).catch((e: unknown) => {
    unstable_rethrow(e);
    console.error("productos relacionados", e);
    return null;
  });
  const items = (pagina?.items ?? []).filter((v) => !excluir.includes(v.id)).slice(0, 4);
  if (items.length === 0) return null;
  return (
    <section aria-labelledby="titulo-relacionados" className="grid gap-5 pt-16">
      <div className="flex flex-wrap items-end justify-between gap-2">
        <h2 id="titulo-relacionados" className="font-heading text-[1.75rem] leading-tight font-medium sm:text-4xl">
          Más de {categoria.nombre}
        </h2>
        <Link href={`/?categoria=${categoria.id}`} className="text-sm font-medium underline underline-offset-4">
          Ver todo
        </Link>
      </div>
      <ul className="grid grid-cols-2 gap-x-3 gap-y-8 sm:grid-cols-4 sm:gap-x-5">
        {items.map((v) => (
          <li key={v.id}>
            <TarjetaVariante variante={v} titulo="h3" />
          </li>
        ))}
      </ul>
    </section>
  );
}

// Ficha de una variante (H05). Inactiva, inexistente o con un id mal
// formado: el 404 de la tienda.
export default async function DetalleVariante({ params }: PageProps<"/variantes/[id]">) {
  const { id } = await params;
  // Lecturas independientes en paralelo (ADR-007).
  const [variante, condiciones] = await Promise.all([
    obtenerVariante(id).catch((e: unknown) => {
      if (e instanceof ErrorApi && (e.status === 400 || e.status === 404)) notFound();
      throw e;
    }),
    obtenerCondicionesEntrega(),
  ]);

  const principal = variante.imagenPrincipal || variante.imagenes[0];
  const imagenes = principal ? [principal, ...variante.imagenes.filter((src) => src !== principal)] : [];
  const categoria = variante.categorias[0];

  return (
    <div className="grid gap-4 pb-36 md:pb-0">
      <Migas
        pasos={[
          { texto: "Catálogo", href: "/" },
          ...(categoria ? [{ texto: categoria.nombre, href: `/?categoria=${categoria.id}` }] : []),
          { texto: variante.nombre },
        ]}
      />

      <article className="grid gap-6 md:grid-cols-[1.15fr_1fr] md:gap-10 lg:gap-16">
        <GaleriaVariante imagenes={imagenes} nombre={variante.nombre} rotulo={variante.marca} />

        <div className="grid content-start gap-6 md:sticky md:top-36 md:self-start">
          <div className="grid gap-1">
            <p className="text-sm text-muted-foreground">{variante.marca}</p>
            <h1 className="font-heading text-[2rem] leading-[1.08] font-medium text-balance md:text-[2.75rem]">{variante.nombre}</h1>
            <p className="text-[0.8125rem] text-muted-foreground">Código {variante.codigo}</p>
          </div>

          <div className="grid gap-1">
            <p className="text-2xl font-semibold tabular-nums text-precio">{formatearPrecio(variante.precioMinoristaCentavos)}</p>
            <div>{variante.disponible ? <p className="text-sm font-medium">Disponible</p> : <SinStock />}</div>
          </div>

          {/* En el teléfono, la acción queda fija abajo (DESIGN.md, sección 5). */}
          <BarraAccionFija>
            <BotonAgregarAlCarrito varianteId={variante.id} disponible={variante.disponible} />
          </BarraAccionFija>

          <div>
            {variante.descripcion ? (
              <Seccion titulo="Descripción" abierta>
                <p className="max-w-prose whitespace-pre-line">{variante.descripcion}</p>
              </Seccion>
            ) : null}
            {variante.otrasVariantes.length > 0 ? (
              <Seccion titulo="Otras presentaciones" abierta>
                <ul className="grid gap-2">
                  {variante.otrasVariantes.map((o) => (
                    <li key={o.id} className="flex flex-wrap items-center justify-between gap-2">
                      <Link href={`/variantes/${o.id}`} className="underline underline-offset-4">
                        {o.nombre}
                      </Link>
                      <span className="flex items-center gap-2 tabular-nums">
                        {formatearPrecio(o.precioMinoristaCentavos)}
                        {o.disponible ? null : <SinStock />}
                      </span>
                    </li>
                  ))}
                </ul>
              </Seccion>
            ) : null}
            {condiciones ? (
              <Seccion titulo="Entrega y pago">
                <EntregaYPago condiciones={condiciones} />
              </Seccion>
            ) : null}
            {variante.categorias.length > 0 ? (
              <Seccion titulo="Categorías">
                <ul className="flex flex-wrap gap-3 text-sm">
                  {variante.categorias.map((c) => (
                    <li key={c.id}>
                      <Link href={`/?categoria=${c.id}`} className="underline underline-offset-4">
                        {c.nombre}
                      </Link>
                    </li>
                  ))}
                </ul>
              </Seccion>
            ) : null}
          </div>
        </div>
      </article>

      {categoria ? (
        <Suspense fallback={null}>
          <Relacionados categoria={categoria} excluir={[variante.id, ...variante.otrasVariantes.map((o) => o.id)]} />
        </Suspense>
      ) : null}
    </div>
  );
}
