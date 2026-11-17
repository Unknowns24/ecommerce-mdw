import type { Metadata } from "next";
import Link from "next/link";

import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { modulosDe } from "@/features/identidad/modulos-panel";
import { requerirSesion } from "@/features/identidad/sesion";

export const metadata: Metadata = { title: "Panel" };

// Inicio del panel: accesos según permisos.
export default async function InicioPanel() {
  const cuenta = await requerirSesion();
  const modulos = modulosDe(cuenta);
  return (
    <section className="grid gap-6">
      <h1 className="text-2xl font-semibold">Hola, {cuenta.nombre}</h1>
      <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {modulos.map((m) => (
          <li key={m.href}>
            <Card className="h-full">
              <CardHeader>
                <CardTitle>
                  <Link href={m.href} className="underline underline-offset-4">
                    {m.texto}
                  </Link>
                </CardTitle>
                <CardDescription>{m.descripcion}</CardDescription>
              </CardHeader>
            </Card>
          </li>
        ))}
      </ul>
      <p className="text-sm">
        <Link href="/admin/sistema" className="underline underline-offset-4">
          Galería del sistema visual
        </Link>
      </p>
    </section>
  );
}
