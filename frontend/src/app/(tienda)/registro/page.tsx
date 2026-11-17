import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import { FormularioRegistro } from "@/features/identidad/components/formulario-registro";
import { obtenerCuenta } from "@/features/identidad/sesion";

import { registrarseAction } from "./acciones";

export const metadata: Metadata = { title: "Crear cuenta" };

export default async function Registro({ searchParams }: PageProps<"/registro">) {
  if (await obtenerCuenta()) redirect("/cuenta");
  const { siguiente } = await searchParams;
  const destino = typeof siguiente === "string" ? siguiente : undefined;
  return (
    <section className="mx-auto grid max-w-md gap-6">
      <div className="grid gap-1">
        <h1 className="font-heading text-3xl leading-tight font-semibold">Crear cuenta</h1>
        <p className="text-muted-foreground">Es opcional: con una cuenta conservás tu carrito y ves tus pedidos.</p>
      </div>
      <FormularioRegistro accion={registrarseAction.bind(null, destino)} />
      <p className="text-sm">
        ¿Ya tenés cuenta?{" "}
        <Link href="/ingresar" className="underline underline-offset-4">
          Ingresá
        </Link>
      </p>
    </section>
  );
}
