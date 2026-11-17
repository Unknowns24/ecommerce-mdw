import type { Metadata } from "next";

import { SinDatos } from "@/components/sin-datos";
import { listarCategorias, listarMarcas, listarVariantes, POR_PAGINA } from "@/features/catalogo/api";
import { AccesosCategorias } from "@/features/catalogo/components/accesos-categorias";
import { FiltrosActivos } from "@/features/catalogo/components/filtros-activos";
import { FiltrosCatalogoForm } from "@/features/catalogo/components/filtros-catalogo";
import { Paginacion } from "@/features/catalogo/components/paginacion";
import { TarjetaVariante } from "@/features/catalogo/components/tarjeta-variante";
import { filtrosDe, hayFiltros } from "@/features/catalogo/filtros";
import { ErrorApi } from "@/lib/api";

export const metadata: Metadata = { title: "Catálogo" };

// Catálogo (H04): tarjetas por variante, búsqueda, filtros por categoría y
// marca, orden por precio y paginación, todo en la query string.
export default async function Catalogo({ searchParams }: PageProps<"/">) {
  const filtros = filtrosDe(await searchParams);

  // Lecturas independientes en paralelo (ADR-007).
  const [pagina, categorias, marcas] = await Promise.all([
    listarVariantes(filtros).catch((e: unknown) => {
      // Un filtro mal armado a mano en la URL (400) no es un error del sistema.
      if (e instanceof ErrorApi && e.status === 400) return null;
      throw e;
    }),
    listarCategorias(),
    listarMarcas(),
  ]);

  const nombreBuscado = filtros.q ? ` con «${filtros.q}»` : "";

  return (
    <div className="grid gap-5">
      <h1 className="font-heading text-3xl leading-tight font-semibold">{filtros.q ? `Resultados para «${filtros.q}»` : "Catálogo"}</h1>
      <AccesosCategorias filtros={filtros} categorias={categorias} />
      <div className="grid gap-6 lg:grid-cols-[15rem_1fr] lg:gap-10">
      <aside className="lg:sticky lg:top-6 lg:self-start">
        <FiltrosCatalogoForm filtros={filtros} marcas={marcas} />
      </aside>

      <section aria-label="Productos" className="grid content-start gap-4">
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
            <p className="text-sm text-muted-foreground" role="status">
              {pagina.total === 1 ? "1 producto" : `${pagina.total} productos`}
            </p>
            <ul className="grid grid-cols-2 gap-x-3 gap-y-8 sm:grid-cols-3 sm:gap-x-5 xl:grid-cols-4">
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
    </div>
  );
}
