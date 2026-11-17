import Link from "next/link";
import { notFound } from "next/navigation";
import type { ReactNode } from "react";

import { buttonVariants } from "@/components/ui/button";
import { BotonSalir } from "@/features/identidad/components/boton-salir";
import { modulosDe } from "@/features/identidad/modulos-panel";
import { requerirSesion } from "@/features/identidad/sesion";

// Barra del panel (spec, sección 5): nombre del sistema, quién sos, tus
// roles, los módulos que tus permisos habilitan y «Salir». Sin sesión,
// /ingresar; con sesión pero sin ningún permiso, el mismo 404 que lo que no
// existe. Cada página igual verifica su permiso y la API decide.
export default async function LayoutAdmin({ children }: { children: ReactNode }) {
  const cuenta = await requerirSesion();
  if (cuenta.permisos.length === 0) {
    notFound();
  }
  const modulos = modulosDe(cuenta);
  // El panel es una herramienta: solo Public Sans, también en los títulos de
  // las tarjetas (DESIGN.md, sección 4).
  return (
    <div className="[--font-bodoni:var(--font-public-sans)]">
      <header className="bg-muted">
        <div className="mx-auto grid max-w-6xl gap-2 px-4 py-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <Link href="/admin" className="font-semibold text-primary underline-offset-4 hover:underline">
              BC Importados · Panel
            </Link>
            <div className="flex flex-wrap items-center gap-2">
              <p className="text-sm">
                <span className="font-medium">
                  {cuenta.nombre} {cuenta.apellido}
                </span>{" "}
                <span className="text-muted-foreground">({cuenta.roles.join(", ")})</span>
              </p>
              <Link href="/" className={buttonVariants({ variant: "ghost" })}>
                Ir a la tienda
              </Link>
              <BotonSalir />
            </div>
          </div>
          <nav aria-label="Módulos del panel" className="-mx-1 flex flex-wrap gap-1">
            {modulos.map((m) => (
              <Link key={m.href} href={m.href} className={buttonVariants({ variant: "ghost", size: "sm" })}>
                {m.texto}
              </Link>
            ))}
          </nav>
        </div>
      </header>
      <main className="mx-auto w-full max-w-6xl px-4 py-6">{children}</main>
    </div>
  );
}
