"use client";

import { useActionState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Campo } from "@/components/campo";
import { cambiarUnidadesAction, quitarItemAction } from "../acciones";
import { ESTADO_INICIAL } from "../estado-inicial";
import type { ConflictoCarrito, EstadoCarrito, LineaCarrito } from "../tipos";
import { DatosLinea, SubtotalLinea } from "./datos-linea";

// Línea del carrito de quien tiene sesión. El id de la línea sale de la lista
// del servidor y se ata con .bind: el formulario no manda ids ni precios.
// Sin redirect: la acción revalida y la página se actualiza en el lugar.
// El 409 no se dibuja acá sino arriba de la lista (ListaCliente): esta línea
// se lo avisa con `alConflicto`. El 400 sí queda debajo del campo.
export function LineaCliente({
  linea,
  alConflicto,
}: {
  linea: LineaCarrito;
  alConflicto: (conflicto: ConflictoCarrito | null) => void;
}) {
  const [estado, cambiar, cambiando] = useActionState<EstadoCarrito, FormData>(
    async (prev, formData) => {
      const nuevo = await cambiarUnidadesAction(linea.id, prev, formData);
      alConflicto(nuevo.mensaje ? { mensaje: nuevo.mensaje, detalles: nuevo.detalles } : null);
      return nuevo;
    },
    ESTADO_INICIAL,
  );
  const [, quitar, quitando] = useActionState<EstadoCarrito, FormData>(
    async () => {
      const nuevo = await quitarItemAction(linea.id);
      alConflicto(null);
      return nuevo;
    },
    ESTADO_INICIAL,
  );
  const ocupado = cambiando || quitando;

  return (
    <li>
      <Card>
        <CardContent className="flex flex-col gap-4">
          <DatosLinea linea={linea} />

          <div className="flex flex-wrap items-end gap-3">
            {linea.disponible ? (
              <form
                action={cambiar}
                aria-label={`Unidades de ${linea.nombre}`}
                className="flex items-end gap-2"
              >
                <div className="w-24">
                  <Campo
                    etiqueta="Unidades"
                    nombre="unidades"
                    tipo="number"
                    inputMode="numeric"
                    requerido
                    valorInicial={estado.valores?.unidades ?? String(linea.unidades)}
                    error={estado.errores?.unidades}
                  />
                </div>
                <Button type="submit" variant="outline" disabled={ocupado}>
                  {cambiando ? (
                    "Actualizando…"
                  ) : (
                    <>
                      Actualizar<span className="sr-only"> unidades de {linea.nombre}</span>
                    </>
                  )}
                </Button>
              </form>
            ) : (
              <p className="text-sm">Unidades: {linea.unidades}</p>
            )}
            <form action={quitar}>
              <Button type="submit" variant="ghost" disabled={ocupado}>
                {quitando ? (
                  "Quitando…"
                ) : (
                  <>
                    Quitar<span className="sr-only"> {linea.nombre} del carrito</span>
                  </>
                )}
              </Button>
            </form>
          </div>

          <SubtotalLinea linea={linea} />
        </CardContent>
      </Card>
    </li>
  );
}
