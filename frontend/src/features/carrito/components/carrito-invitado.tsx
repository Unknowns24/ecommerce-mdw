"use client";

import { useEffect, useState, useSyncExternalStore } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Campo } from "@/components/campo";
import { ErrorDeCarga } from "@/components/error-de-carga";
import { valorizarInvitadoAction } from "../acciones";
import {
  cambiarUnidadesInvitado,
  instantaneaCarritoInvitado,
  instantaneaServidor,
  leerCarritoInvitado,
  quitarDelCarritoInvitado,
  suscribirseAlCarritoInvitado,
  unidadesDeInstantanea,
} from "../almacen-invitado";
import type { LineaCarrito, VistaCarrito } from "../tipos";
import { CarritoVacio } from "./carrito-vacio";
import { DatosLinea, SubtotalLinea } from "./datos-linea";
import { EsqueletoCarrito } from "./esqueleto-carrito";
import { AvisoCarrito, ResumenCarrito } from "./resumen-carrito";

// Resultado de valorizar: guarda el texto de localStorage con el que se pidió,
// para saber si ya quedó viejo (el comprador cambió algo y se está recalculando).
type Resultado = { crudo: string; vista: VistaCarrito } | { crudo: string; fallo: true };

const sinSuscripcion = () => () => {};

// Carrito de quien no tiene sesión. En localStorage solo hay {varianteId, unidades};
// nombres, precios y disponibilidad los devuelve el servidor (valorizarInvitadoAction).
// Sin stock validado acá: la disponibilidad se confirma en el checkout (decisión del equipo).
export function CarritoInvitado() {
  const crudo = useSyncExternalStore(
    suscribirseAlCarritoInvitado,
    instantaneaCarritoInvitado,
    instantaneaServidor,
  );
  // false en el servidor y en la hidratación: así no se muestra «vacío» antes de leer el navegador.
  const montado = useSyncExternalStore(sinSuscripcion, () => true, () => false);
  const hayItems = unidadesDeInstantanea(crudo) > 0;

  const [resultado, setResultado] = useState<Resultado | null>(null);
  const [intento, setIntento] = useState(0);
  const [enCurso, setEnCurso] = useState<string | null>(null);
  const [errores, setErrores] = useState<Record<string, string>>({});

  useEffect(() => {
    if (!montado || !hayItems) return;
    let vigente = true;
    valorizarInvitadoAction(leerCarritoInvitado())
      .then((vista) => {
        if (vigente) {
          setResultado({ crudo, vista });
          setEnCurso(null);
        }
      })
      .catch((e: unknown) => {
        console.error("valorizar el carrito del invitado", e);
        if (vigente) {
          setResultado({ crudo, fallo: true });
          setEnCurso(null);
        }
      });
    return () => {
      vigente = false;
    };
  }, [crudo, montado, hayItems, intento]);

  if (!montado) return <EsqueletoCarrito />;
  if (!hayItems) return <CarritoVacio />;
  if (resultado === null) return <EsqueletoCarrito />;
  if ("fallo" in resultado) {
    return (
      <ErrorDeCarga
        reintentar={() => {
          setResultado(null);
          setIntento((n) => n + 1);
        }}
      />
    );
  }

  const actualizando = resultado.crudo !== crudo;

  function cambiar(linea: LineaCarrito, texto: string) {
    const unidades = Number(texto);
    if (texto.trim() === "" || !Number.isInteger(unidades) || unidades < 1) {
      setErrores((e) => ({ ...e, [linea.id]: "Ingresá una cantidad entera de 1 o más." }));
      return;
    }
    setErrores((e) => {
      const { [linea.id]: _quitado, ...resto } = e;
      void _quitado;
      return resto;
    });
    setEnCurso(linea.id);
    cambiarUnidadesInvitado(linea.varianteId, unidades);
  }

  function quitar(linea: LineaCarrito) {
    setEnCurso(linea.id);
    quitarDelCarritoInvitado(linea.varianteId);
  }

  return (
    <div className="flex flex-col gap-4">
      <AvisoCarrito invitado />
      <ul className="flex flex-col gap-3" aria-busy={actualizando}>
        {resultado.vista.lineas.map((linea) => {
          const estaEnCurso = actualizando && enCurso === linea.id;
          return (
            <li key={linea.id}>
              <Card>
                <CardContent className="flex flex-col gap-4">
                  <DatosLinea linea={linea} />
                  <div className="flex flex-wrap items-end gap-3">
                    {linea.disponible ? (
                      <form
                        aria-label={`Unidades de ${linea.nombre}`}
                        className="flex items-end gap-2"
                        onSubmit={(ev) => {
                          ev.preventDefault();
                          const texto = new FormData(ev.currentTarget).get("unidades");
                          cambiar(linea, typeof texto === "string" ? texto : "");
                        }}
                      >
                        <div className="w-24">
                          <Campo
                            etiqueta="Unidades"
                            nombre="unidades"
                            tipo="number"
                            inputMode="numeric"
                            requerido
                            valorInicial={String(linea.unidades)}
                            error={errores[linea.id]}
                          />
                        </div>
                        <Button type="submit" variant="outline" disabled={actualizando}>
                          {estaEnCurso ? (
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
                    <Button
                      type="button"
                      variant="ghost"
                      disabled={actualizando}
                      onClick={() => quitar(linea)}
                    >
                      Quitar<span className="sr-only"> {linea.nombre} del carrito</span>
                    </Button>
                  </div>
                  <SubtotalLinea linea={linea} />
                </CardContent>
              </Card>
            </li>
          );
        })}
      </ul>
      <ResumenCarrito totalCentavos={resultado.vista.totalEstimadoCentavos} />
    </div>
  );
}
