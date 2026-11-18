import { SearchIcon } from "lucide-react";
import Link from "next/link";
import { unstable_rethrow } from "next/navigation";

import { MenuMovil } from "@/components/menu-movil";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { EnlaceCarrito } from "@/features/carrito/components/enlace-carrito";
import { listarCategorias } from "@/features/catalogo/api";
import { MenuSesion } from "@/features/identidad/components/menu-sesion";

// Encabezado de (tienda) y (cuenta) (spec, sección 5): nombre con link al
// inicio, búsqueda, carrito e «Ingresar» o el nombre de quien tiene sesión.
//
// Teléfono: menú y logotipo a la izquierda, carrito a la derecha y, debajo,
// la búsqueda a lo ancho. Escritorio: logotipo a la izquierda; búsqueda,
// cuenta y carrito a la derecha; y debajo, las categorías. Hay un solo
// formulario de búsqueda: cambia de lugar con la grilla.
export async function EncabezadoTienda() {
  const categorias = await listarCategorias().catch((e: unknown) => {
    unstable_rethrow(e); // las señales internas de Next (dinámico, redirect) siguen su camino
    // El encabezado no se cae si la API falla: quedan el logo, la búsqueda y la cuenta.
    console.error("categorías del encabezado", e);
    return [];
  });
  const todas = categorias.flatMap((c) => [c, ...(c.hijas ?? [])]);

  return (
    <header className="relative z-40 border-b bg-background lg:sticky lg:top-0">
      <div className="mx-auto grid max-w-6xl grid-cols-[auto_1fr_auto] items-center gap-x-1 px-4 lg:grid-cols-[1fr_auto_auto] lg:gap-x-4">
        <div className="col-start-1 row-start-1 flex h-16 items-center lg:hidden">
          <MenuMovil>
            <nav aria-label="Secciones" className="py-2">
              <ul className="grid">
                <li>
                  <Link href="/" className="flex h-12 items-center border-b text-sm tracking-[0.15em] uppercase">
                    Todo el catálogo
                  </Link>
                </li>
                {todas.map((c) => (
                  <li key={c.id}>
                    <Link href={`/?categoria=${c.id}`} className="flex h-12 items-center border-b text-sm tracking-[0.15em] uppercase">
                      {c.nombre}
                    </Link>
                  </li>
                ))}
              </ul>
            </nav>
            <div className="pt-4">
              <MenuSesion />
            </div>
          </MenuMovil>
        </div>

        <Link
          href="/"
          className="col-start-2 row-start-1 flex h-16 items-center justify-self-start text-lg leading-none font-light tracking-[0.28em] uppercase sm:text-2xl lg:col-start-1 lg:h-20"
        >
          BC Importados
        </Link>

        <form
          action="/"
          role="search"
          className="col-span-3 row-start-2 flex gap-1 pb-3 lg:col-span-1 lg:col-start-2 lg:row-start-1 lg:w-64 lg:pb-0"
        >
          <label htmlFor="busqueda-encabezado" className="sr-only">
            Buscar productos
          </label>
          <Input id="busqueda-encabezado" name="q" type="search" placeholder="Perfumes, cremas, auriculares…" />
          <Button type="submit" variant="ghost" size="icon" className="shrink-0">
            <SearchIcon className="size-5" aria-hidden="true" />
            <span className="sr-only">Buscar</span>
          </Button>
        </form>

        <div className="col-start-3 row-start-1 flex items-center justify-end gap-1">
          <nav aria-label="Tu cuenta" className="hidden lg:block">
            <MenuSesion />
          </nav>
          <EnlaceCarrito />
        </div>

        {todas.length > 0 ? (
          <nav aria-label="Secciones" className="col-span-3 row-start-2 hidden justify-center lg:flex">
            <ul className="flex gap-10">
              <li>
                <Link href="/" className={claseSeccion}>
                  Todo
                </Link>
              </li>
              {todas.map((c) => (
                <li key={c.id}>
                  <Link href={`/?categoria=${c.id}`} className={claseSeccion}>
                    {c.nombre}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>
        ) : null}
      </div>
    </header>
  );
}

const claseSeccion = "inline-flex h-11 items-center border-b border-transparent text-xs tracking-[0.18em] uppercase hover:border-current";
