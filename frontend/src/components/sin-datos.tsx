import Link from "next/link";
import type { ReactElement } from "react";
import { Button } from "@/components/ui/button";

export function SinDatos({ titulo, descripcion, accion }: {
  titulo: string;
  descripcion: string;
  accion?: { texto: string; href: string };
}): ReactElement {
  return (
    <section className="grid justify-items-center gap-3 px-4 py-8 text-center">
      <h2 className="text-lg font-semibold">{titulo}</h2>
      <p className="text-muted-foreground">{descripcion}</p>
      {accion && (
        <Button asChild className="w-full sm:w-auto">
          <Link href={accion.href}>{accion.texto}</Link>
        </Button>
      )}
    </section>
  );
}
