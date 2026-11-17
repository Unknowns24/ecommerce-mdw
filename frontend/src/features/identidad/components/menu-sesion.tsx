import Link from "next/link";
import { unstable_rethrow } from "next/navigation";
import type { ReactElement } from "react";

import { buttonVariants } from "@/components/ui/button";

import { obtenerCuenta } from "../sesion";
import { BotonSalir } from "./boton-salir";

// Sin sesión: «Ingresar». Con sesión: el nombre, «Mi cuenta» y «Salir»
// (contrato, sección 3). Quien tiene algún permiso ve además «Panel».
export async function MenuSesion(): Promise<ReactElement> {
  let cuenta: Awaited<ReturnType<typeof obtenerCuenta>>;
  try {
    cuenta = await obtenerCuenta();
  } catch (e) {
    unstable_rethrow(e); // las señales internas de Next (dinámico, redirect) siguen su camino
    // El encabezado no se cae si la API falla: la página muestra su propio
    // error y acá queda un acceso a la cuenta.
    console.error("menú de sesión", e);
    return (
      <Link className={buttonVariants({ variant: "ghost" })} href="/cuenta">
        Mi cuenta
      </Link>
    );
  }
  if (!cuenta) {
    return (
      <Link className={buttonVariants({ variant: "ghost" })} href="/ingresar">
        Ingresar
      </Link>
    );
  }
  return (
    <div data-con-sesion className="flex flex-wrap items-center gap-1">
      <span className="px-2 text-sm font-medium">{cuenta.nombre}</span>
      {cuenta.permisos.length > 0 ? (
        <Link className={buttonVariants({ variant: "ghost" })} href="/admin">
          Panel
        </Link>
      ) : null}
      <Link className={buttonVariants({ variant: "ghost" })} href="/cuenta">
        Mi cuenta
      </Link>
      <BotonSalir />
    </div>
  );
}
