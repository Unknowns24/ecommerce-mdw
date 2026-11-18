import type { Metadata } from "next";

import { SinDatos } from "@/components/sin-datos";
import { Separator } from "@/components/ui/separator";
import { listarCategorias, listarMarcas, listarVariantes, obtenerCondicionesEntrega, POR_PAGINA } from "@/features/catalogo/api";
import { FOTOS_CATEGORIA } from "@/features/catalogo/campanas";
import { AccesosCategorias } from "@/features/catalogo/components/accesos-categorias";
import { BannerCategoria } from "@/features/catalogo/components/banner-categoria";
import { FiltrosActivos } from "@/features/catalogo/components/filtros-activos";
import { FiltrosCatalogoForm } from "@/features/catalogo/components/filtros-catalogo";
import { Paginacion } from "@/features/catalogo/components/paginacion";
import { Portada } from "@/features/catalogo/components/portada";
import { TarjetaVariante } from "@/features/catalogo/components/tarjeta-variante";
import { filtrosDe, hayFiltros } from "@/features/catalogo/filtros";
import { ErrorApi } from "@/lib/api";

export const metadata: Metadata = { title: "Catálogo" };

// Catálogo (H04): tarjetas por variante, búsqueda, filtros por categoría y
// marca, orden por precio y paginación, todo en la query string. Sin filtros,
// arriba va la portada editorial (DESIGN.md, sección 12); con filtros, un
// banner con el título de lo que se está viendo.
export default async function Catalogo({ searchParams }: PageProps<"/">) {
  const filtros = filtrosDe(await searchParams);

  // Lecturas independientes en paralelo (ADR-007).
  const portada = !hayFiltros(filtros) && filtros.pagina === 1 && filtros.orden === "nombre_asc";
  const [pagina, categorias, marcas, condiciones] = await Promise.all([
    listarVariantes(filtros).catch((e: unknown) => {
      // Un filtro mal armado a mano en la URL (400) no es un error del sistema.
      if (e instanceof ErrorApi && e.status === 400) return null;
      throw e;
    }),
    listarCategorias(),
    listarMarcas(),
    portada ? obtenerCondicionesEntrega() : null,
  ]);

  const nombreBuscado = filtros.q ? ` con «${filtros.q}»` : "";
  const categoria = categorias.flatMap((c) => [c, ...(c.hijas ?? [])]).find((c) => c.id === filtros.categoria);
  const marca = marcas.find((m) => m.id === filtros.marca);
  const titulo = filtros.q ? `Resultados para «${filtros.q}»` : (categoria?.nombre ?? marca?.nombre ?? "Catálogo");
  const resumen = pagina && pagina.items.length > 0 ? (pagina.total === 1 ? "1 producto" : `${pagina.total} productos`) : null;

  return (
    <>
      {portada ? (
        <Portada categorias={categorias} marcas={marcas} productos={pagina?.items ?? []} condiciones={condiciones} />
      ) : (
        <BannerCategoria titulo={titulo} foto={categoria ? FOTOS_CATEGORIA[categoria.id] : undefined} />
      )}
      <div id="catalogo" className={portada ? "grid scroll-mt-32 gap-5 pt-14 sm:pt-20" : "grid gap-5 pt-5 sm:pt-8"}>
        {/* Con portada, su hero es el h1 y el catálogo pasa a ser una sección. */}
        {portada ? (
          <h2 className="text-center text-2xl leading-tight font-light tracking-[0.2em] uppercase sm:text-3xl">Todo el catálogo</h2>
        ) : null}
        <AccesosCategorias filtros={filtros} categorias={categorias} />

        <div className="grid gap-4">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            {resumen ? (
              <p className="text-xs tracking-[0.14em] text-muted-foreground uppercase" role="status">
                {resumen}
              </p>
            ) : (
              <span />
            )}
            <FiltrosCatalogoForm filtros={filtros} marcas={marcas} />
          </div>
          <Separator />
        </div>

        <section aria-label="Productos" className="grid content-start gap-6">
          <FiltrosActivos filtros={filtros} categorias={categorias} marcas={marcas} />
          {pagina === null ? (
            <SinDatos
              titulo="No pudimos aplicar esos filtros"
              descripcion="La dirección tiene filtros que no reconocemos."
              accion={{ texto: "Limpiar filtros", href: "/" }}
            />
          ) : pagina.total === 0 && !hayFiltros(filtros) ? (
            <SinDatos titulo="Todavía no hay productos" descripcion="El catálogo está vacío por ahora. Volvé a visitarnos pronto." />
          ) : pagina.items.length === 0 ? (
            <SinDatos
              titulo="Ningún producto coincide"
              descripcion={`No encontramos productos${nombreBuscado} con los filtros elegidos.`}
              accion={{ texto: "Limpiar filtros", href: "/" }}
            />
          ) : (
            <>
              <ul className="grid grid-cols-2 gap-x-2 gap-y-10 sm:grid-cols-3 sm:gap-x-3 sm:gap-y-14">
                {pagina.items.map((v) => (
                  <li key={v.id}>
                    <TarjetaVariante variante={v} />
                  </li>
                ))}
              </ul>
              <Paginacion filtros={filtros} total={pagina.total} porPagina={POR_PAGINA} />
            </>
          )}
        </section>
      </div>
    </>
  );
}
