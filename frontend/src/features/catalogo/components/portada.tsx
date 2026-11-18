import Link from "next/link";

import { buttonVariants } from "@/components/ui/button";
import { formatearPrecio } from "@/lib/formato";
import { cn } from "@/lib/utils";

import { listarVariantes, type CondicionesEntrega } from "../api";
import { FOTO_PORTADA, FOTOS_CATEGORIA } from "../campanas";
import type { CategoriaPublica, FiltrosCatalogo, MarcaPublica, VarianteResumen } from "../tipos";
import { CampoCampana, tono } from "./campo-campana";
import { TarjetaVariante } from "./tarjeta-variante";

type Grupo = { id: string; nombre: string; total: number; desde?: number; href: string };

const BASE: FiltrosCatalogo = { q: "", categoria: "", marca: "", orden: "precio_asc", pagina: 1 };

// Cuántos productos tiene un grupo y desde qué precio: la primera página
// ordenada por precio ya trae el total y el más barato.
async function resumir(f: Partial<FiltrosCatalogo>) {
  const p = await listarVariantes({ ...BASE, ...f });
  return { total: p.total, desde: p.items[0]?.precioMinoristaCentavos };
}

const cantidad = (n: number) => (n === 1 ? "1 producto" : `${n} productos`);

// Botón con borde sobre un campo de color: hereda el color del texto.
const botonCampana = cn(buttonVariants({ variant: "outline", size: "lg" }), "border-current bg-transparent text-current hover:bg-current/10 hover:text-current");

// Portada de la tienda: solo cuando el catálogo se abre sin filtros. Todo
// sale de la API (categorías, marcas, cantidades, precios y condiciones de
// entrega); las fotos de campaña, de campanas.ts.
export async function Portada({
  categorias,
  marcas,
  productos,
  condiciones,
}: {
  categorias: CategoriaPublica[];
  marcas: MarcaPublica[];
  productos: VarianteResumen[];
  condiciones: CondicionesEntrega | null;
}) {
  const todas = categorias.flatMap((c) => [c, ...(c.hijas ?? [])]);
  const [resCategorias, resMarcas] = await Promise.all([
    Promise.all(todas.map((c) => resumir({ categoria: c.id }))),
    Promise.all(marcas.map((m) => resumir({ marca: m.id }))),
  ]);

  // Las categorías con más productos van primero; las vacías no se muestran.
  const grupos: Grupo[] = todas
    .map((c, i) => ({ id: c.id, nombre: c.nombre, ...resCategorias[i], href: `/?categoria=${c.id}` }))
    .filter((g) => g.total > 0)
    .sort((a, b) => b.total - a.total);
  const gruposMarca: Grupo[] = marcas
    .map((m, i) => ({ id: m.id, nombre: m.nombre, ...resMarcas[i], href: `/?marca=${m.id}` }))
    .filter((g) => g.total > 0)
    .sort((a, b) => b.total - a.total);

  const conStock = productos.filter((p) => p.disponible).slice(0, 4);
  const [primera, segunda, ...resto] = grupos;

  return (
    <>
      <Hero principal={primera} condiciones={condiciones} />

      {primera ? (
        <section data-ancho="completo" aria-labelledby="titulo-categorias" className="grid gap-5 pt-14 sm:pt-20">
          <Encabezado id="titulo-categorias" titulo="Comprá por categoría" bajada="Fragancias, cuidado de la piel y audio." />
          <ul className="grid gap-1 md:grid-cols-2">
            {[primera, segunda].filter(Boolean).map((g, i) => (
              <li key={g!.id}>
                <Bloque grupo={g!} color={tono(i)} />
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {conStock.length >= 2 ? (
        <section aria-labelledby="titulo-stock" className="grid gap-5 pt-14 sm:pt-20 lg:grid-cols-[1fr_2fr] lg:gap-12">
          <div className="flex flex-wrap items-end justify-between gap-2 lg:flex-col lg:items-start lg:justify-start lg:gap-4">
            <div className="grid gap-1">
              <h2 id="titulo-stock" className="font-heading text-[1.75rem] leading-tight font-medium sm:text-4xl">
                Con stock para entregar
              </h2>
              <p className="text-sm text-muted-foreground">Se reservan al confirmar el pedido.</p>
            </div>
            <Link href="#catalogo" className="text-sm font-medium underline underline-offset-4">
              Ver todo el catálogo
            </Link>
          </div>
          {/* En el teléfono, un carril que se desliza; desde sm, grilla. */}
          <ul
            className={cn(
              "-mx-4 flex snap-x snap-mandatory gap-3 overflow-x-auto px-4 pb-2 sm:mx-0 sm:grid sm:gap-5 sm:overflow-visible sm:px-0",
              conStock.length === 2 ? "sm:grid-cols-2" : conStock.length === 3 ? "sm:grid-cols-3" : "sm:grid-cols-4",
            )}
          >
            {conStock.map((v) => (
              <li key={v.id} className="w-[44%] shrink-0 snap-start sm:w-auto">
                <TarjetaVariante variante={v} titulo="h3" />
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {resto.length > 0 ? (
        <section data-ancho="completo" aria-label="Más categorías" className="grid gap-1 pt-14 sm:pt-20">
          {resto.map((g, i) => (
            <Banner key={g.id} grupo={g} color={tono(i + 2)} invertido={i % 2 === 1} />
          ))}
        </section>
      ) : null}

      {gruposMarca.length > 0 ? (
        <section aria-labelledby="titulo-marcas" className="grid gap-6 pt-14 sm:pt-20">
          <h2 id="titulo-marcas" className="text-center font-heading text-[1.75rem] leading-tight font-medium sm:text-4xl">
            Marcas
          </h2>
          <ul className="grid border-t sm:grid-cols-3 sm:border-t-0">
            {gruposMarca.map((g) => (
              <li key={g.id} className="border-b sm:border-b-0 sm:not-last:border-r">
                <Link href={g.href} className="group grid gap-1 py-5 sm:place-items-center sm:px-4 sm:py-10 sm:text-center">
                  <span className="font-heading text-2xl leading-tight font-medium group-hover:underline group-hover:underline-offset-4 sm:text-3xl">
                    {g.nombre}
                  </span>
                  <span className="text-sm text-muted-foreground">
                    {cantidad(g.total)}
                    {g.desde !== undefined ? `, desde ${formatearPrecio(g.desde)}` : ""}
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </>
  );
}

function Encabezado({ id, titulo, bajada }: { id: string; titulo: string; bajada: string }) {
  return (
    <div className="mx-auto grid w-full max-w-6xl gap-1 px-4">
      <h2 id={id} className="font-heading text-[1.75rem] leading-tight font-medium sm:text-4xl">
        {titulo}
      </h2>
      <p className="text-sm text-muted-foreground">{bajada}</p>
    </div>
  );
}

// Hero: el titular de la tienda y lo que hace falta saber antes de comprar,
// con las condiciones reales de la configuración.
function Hero({ principal, condiciones }: { principal?: Grupo; condiciones: CondicionesEntrega | null }) {
  const tarifas = [...(condiciones?.tarifas ?? [])].sort((a, b) => a.costoCentavos - b.costoCentavos);
  const km = condiciones ? `${(condiciones.distanciaMaximaMetros / 1000).toLocaleString("es-AR", { maximumFractionDigits: 1 })} km` : "";
  const datos = [
    { titulo: "Retiro", texto: "Sin cargo, en el local." },
    ...(condiciones && tarifas.length > 0
      ? [{ titulo: "Envío", texto: `Hasta ${km} del local, desde ${formatearPrecio(tarifas[0].costoCentavos)}.` }]
      : []),
    { titulo: "Pago", texto: condiciones?.efectivoHabilitado ? "Mercado Pago, o efectivo al retirar." : "Mercado Pago." },
  ];
  return (
    <section data-ancho="completo" aria-labelledby="titulo-portada" className="-mt-6">
      <CampoCampana foto={FOTO_PORTADA} color="bg-primary text-primary-foreground">
        <div className="mx-auto grid max-w-6xl gap-12 px-4 pt-16 pb-12 sm:pt-24 sm:pb-16 lg:min-h-[34rem] lg:grid-cols-[1.4fr_1fr] lg:items-end lg:gap-16 lg:pt-32">
          <div className="grid gap-6">
            <p className="text-sm opacity-80">Perfumería, cosmética y electrónica importadas</p>
            <h1 id="titulo-portada" className="font-heading text-[2.75rem] leading-[1.02] font-medium sm:text-6xl lg:text-7xl">
              Fragancias, cremas y body splash
            </h1>
            <p className="max-w-md text-base/relaxed opacity-85">Comprás sin crear una cuenta y elegís cómo recibirlo.</p>
            <div className="flex flex-col gap-3 sm:flex-row">
              {principal ? (
                <Link href={principal.href} className={cn(botonCampana, "w-full sm:w-auto")}>
                  Ver {principal.nombre.toLocaleLowerCase("es-AR")}
                </Link>
              ) : null}
              <Link href="#catalogo" className={cn(botonCampana, "w-full sm:w-auto")}>
                Todo el catálogo
              </Link>
            </div>
          </div>
          <dl className="grid gap-4 border-t border-current/30 pt-6 text-sm sm:grid-cols-3 lg:grid-cols-1 lg:border-t-0 lg:border-l lg:pt-0 lg:pl-8">
            {datos.map((d) => (
              <div key={d.titulo} className="grid gap-0.5">
                <dt className="font-heading text-lg font-medium">{d.titulo}</dt>
                <dd className="opacity-85">{d.texto}</dd>
              </div>
            ))}
          </dl>
        </div>
      </CampoCampana>
    </section>
  );
}

// Bloque grande de una categoría (la grilla de dos de la portada).
function Bloque({ grupo, color }: { grupo: Grupo; color: string }) {
  return (
    <Link href={grupo.href} className="group block">
      <CampoCampana foto={FOTOS_CATEGORIA[grupo.id]} color={color} className="aspect-[5/4] md:aspect-[4/5] lg:aspect-[4/3]">
        <div className="flex h-full flex-col justify-between p-6 sm:p-10">
          <p className="text-sm opacity-80">{cantidad(grupo.total)}</p>
          <div className="grid justify-items-start gap-4">
            <h3 className="font-heading text-5xl leading-none font-medium sm:text-6xl lg:text-7xl">{grupo.nombre}</h3>
            {grupo.desde !== undefined ? <p className="text-sm opacity-85">Desde {formatearPrecio(grupo.desde)}</p> : null}
            <span className={cn(botonCampana, "transition-colors group-hover:bg-current/10")}>
              Ver {grupo.nombre.toLocaleLowerCase("es-AR")}
            </span>
          </div>
        </div>
      </CampoCampana>
    </Link>
  );
}

// Banner horizontal: el campo con el nombre de un lado y el texto del otro;
// se alternan para que la portada no repita la misma forma.
function Banner({ grupo, color, invertido }: { grupo: Grupo; color: string; invertido: boolean }) {
  return (
    <Link href={grupo.href} className="group grid md:grid-cols-2">
      <CampoCampana
        foto={FOTOS_CATEGORIA[grupo.id]}
        color={color}
        className={cn("flex min-h-56 items-end p-6 sm:min-h-80 sm:p-10", invertido && "md:order-2")}
      >
        <h3 className="font-heading text-5xl leading-none font-medium break-words sm:text-6xl md:text-5xl lg:text-7xl">{grupo.nombre}</h3>
      </CampoCampana>
      <div className="grid content-center justify-items-start gap-2 bg-background px-4 py-8 sm:p-10 lg:px-16">
        <p className="text-sm text-muted-foreground">{cantidad(grupo.total)}</p>
        {grupo.desde !== undefined ? (
          <p className="font-heading text-3xl leading-tight font-medium lg:text-4xl">Desde {formatearPrecio(grupo.desde)}</p>
        ) : null}
        <span className="pt-2 text-base font-medium underline decoration-1 underline-offset-4 group-hover:decoration-2">
          Ver {grupo.nombre.toLocaleLowerCase("es-AR")}
        </span>
      </div>
    </Link>
  );
}
