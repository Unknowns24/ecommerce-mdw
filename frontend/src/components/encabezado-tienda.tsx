import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { EnlaceCarrito } from "@/features/carrito/components/enlace-carrito";
import { MenuSesion } from "@/features/identidad/components/menu-sesion";

// Encabezado de (tienda) y (cuenta) (spec, sección 5): nombre con link al
// inicio, búsqueda, carrito e «Ingresar» o el nombre de quien tiene sesión.
export function EncabezadoTienda() {
  return (
    <header className="bg-muted">
      <div className="mx-auto flex max-w-6xl flex-wrap items-center gap-x-4 gap-y-3 px-4 py-3">
        <Link href="/" className="order-1 font-heading text-xl font-semibold leading-none sm:text-2xl text-primary underline-offset-4 hover:underline">
          BC Importados
        </Link>
        <div className="order-2 ml-auto sm:order-3 sm:ml-0">
          <EnlaceCarrito />
        </div>
        <nav aria-label="Tu cuenta" className="order-3 sm:order-4 has-data-con-sesion:-ml-2 has-data-con-sesion:w-full sm:has-data-con-sesion:ml-0 sm:has-data-con-sesion:w-auto">
          <MenuSesion />
        </nav>
        <form action="/" role="search" className="order-4 flex w-full gap-2 sm:order-2 sm:w-auto sm:flex-1">
          <label htmlFor="busqueda-encabezado" className="sr-only">
            Buscar productos
          </label>
          <Input id="busqueda-encabezado" name="q" type="search" placeholder="Perfumes, cremas, auriculares…" className="bg-background" />
          <Button type="submit" variant="outline" className="bg-background">
            Buscar
          </Button>
        </form>
      </div>
    </header>
  );
}
