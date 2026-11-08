"use client";

import Link from "next/link";
import { useActionState, useEffect, useRef, useState } from "react";
import { confirmarPedidoAction } from "@/app/(tienda)/checkout/acciones";
import { Campo } from "@/components/campo";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Select } from "@/components/ui/select";
import { vaciarCarritoInvitado } from "@/features/carrito/almacen-invitado";
import { listarDisponibles, textoDisponibles } from "@/features/carrito/disponibles";
import type { VistaCarrito } from "@/features/carrito/tipos";
import { formatearPrecio } from "@/lib/formato";
import { estimarEnvio } from "../envio";
import type { ConfiguracionTienda, EstadoCheckout, ItemPedido, ModoEntrega } from "../tipos";
import { PanelPedidoSinPago } from "./pedido-sin-pago";
import { ResumenPedido } from "./resumen-pedido";

type Props = {
  /** null = no hay configuración cargada (404): sin envío y sin efectivo. */
  config: ConfiguracionTienda | null;
  /** Punto de partida con los datos de la cuenta (vacío para el invitado). */
  inicial: { nombre: string; apellido: string; correo: string };
  vista: VistaCarrito;
  /** Solo el invitado: lo que había en localStorage. Con sesión, null (los ítems los lee el servidor). */
  itemsInvitado: ItemPedido[] | null;
};

const ESTADO_INICIAL: EstadoCheckout = { enviado: false };

function esRedireccion(e: unknown): boolean {
  return (
    typeof e === "object" &&
    e !== null &&
    "digest" in e &&
    typeof e.digest === "string" &&
    e.digest.startsWith("NEXT_REDIRECT")
  );
}

export function FormularioCheckout({ config, inicial, vista, itemsInvitado }: Props) {
  const [estado, accion, pendiente] = useActionState<EstadoCheckout, FormData>(
    async (prev, formData) => {
      try {
        return await confirmarPedidoAction(itemsInvitado ?? [], prev, formData);
      } catch (e) {
        // Una redirección solo ocurre si el pedido se creó: el carrito local del
        // invitado ya cumplió su función. (Con sesión lo vacía el servidor.)
        if (itemsInvitado && esRedireccion(e)) vaciarCarritoInvitado();
        throw e;
      }
    },
    ESTADO_INICIAL,
  );

  // Pedido creado cuyo pago no pudo iniciarse: el carrito del invitado también
  // ya se usó, y la pantalla deja de ofrecer «Confirmar».
  const hayPedidoSinPago = estado.pedido !== undefined;
  useEffect(() => {
    if (hayPedidoSinPago && itemsInvitado) vaciarCarritoInvitado();
  }, [hayPedidoSinPago, itemsInvitado]);

  // Doble envío: el botón se deshabilita mientras corre la acción (`pendiente`),
  // pero un segundo clic puede llegar antes de que React pinte ese estado. Esta
  // guarda lo corta en el mismo evento. Se libera cuando la acción devuelve un
  // estado (un fallo); si salió bien, la redirección se lleva la pantalla.
  const enviando = useRef(false);
  useEffect(() => {
    enviando.current = false;
  }, [estado]);

  // Lo único que se espeja fuera del formulario es lo que cambia qué se
  // muestra (campos de envío, medios de pago, estimación). Los valores de los
  // campos los devuelve la acción, no viven en useState.
  const [modo, setModo] = useState<ModoEntrega>("RETIRO");
  const [distancia, setDistancia] = useState("");

  if (estado.pedido) {
    return <PanelPedidoSinPago estado={{ ...estado, pedido: estado.pedido }} />;
  }

  const valores = estado.valores ?? {};
  const errores = estado.errores ?? {};
  const envioDisponible =
    config !== null && config.distanciaMaximaMetros > 0 && config.tarifas.length > 0;
  const efectivoDisponible = modo === "RETIRO" && config?.efectivoHabilitado === true;
  const metros = /^\d+$/.test(distancia.trim()) ? Number(distancia.trim()) : null;
  const estimacion =
    modo === "ENVIO" && envioDisponible && metros !== null ? estimarEnvio(config, metros) : null;
  const filasConflicto = listarDisponibles(estado.detalles);

  return (
    <div className="flex flex-col gap-6">
      {estado.mensaje ? (
        // 409 (y errores del cuerpo sin campo): arriba del formulario.
        <Alert variant="destructive" role="alert">
          <p className="font-medium">{estado.mensaje}</p>
          {filasConflicto.length > 0 ? (
            <ul className="list-disc pl-5">
              {filasConflicto.map((f) => (
                <li key={f.varianteId}>
                  {vista.lineas.find((l) => l.varianteId === f.varianteId)?.nombre ?? "Producto"}:{" "}
                  {textoDisponibles(f.disponibles)}
                </li>
              ))}
            </ul>
          ) : null}
          {filasConflicto.length > 0 ? (
            <p>
              <Link href="/carrito" className="underline">
                Volver al carrito para ajustar las unidades
              </Link>
            </p>
          ) : null}
        </Alert>
      ) : null}

      <form
        action={accion}
        className="flex flex-col gap-6"
        onSubmit={(ev) => {
          if (enviando.current) {
            ev.preventDefault();
            return;
          }
          enviando.current = true;
        }}
        onChange={(ev) => {
          const datos = new FormData(ev.currentTarget);
          setModo(datos.get("modoEntrega") === "ENVIO" ? "ENVIO" : "RETIRO");
          const d = datos.get("distanciaMetros");
          setDistancia(typeof d === "string" ? d : "");
        }}
      >
        <fieldset className="flex flex-col gap-4">
          <legend className="text-lg font-semibold">Tus datos</legend>
          <Campo
            etiqueta="Nombre"
            nombre="nombre"
            requerido
            valorInicial={valores.nombre ?? inicial.nombre}
            error={errores.nombre}
          />
          <Campo
            etiqueta="Apellido"
            nombre="apellido"
            requerido
            valorInicial={valores.apellido ?? inicial.apellido}
            error={errores.apellido}
          />
          <Campo
            etiqueta="Correo"
            nombre="correo"
            tipo="email"
            inputMode="email"
            requerido
            valorInicial={valores.correo ?? inicial.correo}
            error={errores.correo}
          />
          <Campo
            etiqueta="Teléfono"
            nombre="telefono"
            tipo="tel"
            inputMode="numeric"
            requerido
            valorInicial={valores.telefono ?? ""}
            error={errores.telefono}
          />
          <Campo
            etiqueta="DNI"
            nombre="dni"
            inputMode="numeric"
            requerido
            ayuda="7 u 8 dígitos, sin puntos."
            valorInicial={valores.dni ?? ""}
            error={errores.dni}
          />
        </fieldset>

        <fieldset className="flex flex-col gap-4">
          <legend className="text-lg font-semibold">Entrega</legend>
          <Campo
            etiqueta="¿Cómo recibís tu pedido?"
            nombre="modoEntrega"
            valorInicial={valores.modoEntrega ?? "RETIRO"}
            error={errores.modoEntrega}
            ayuda={
              envioDisponible
                ? undefined
                : "El envío no está disponible por el momento. Podés retirar en el local."
            }
          >
            {(control) => (
              <Select {...control}>
                <option value="RETIRO">Retiro en el local</option>
                <option value="ENVIO" disabled={!envioDisponible}>
                  {envioDisponible ? "Envío a domicilio" : "Envío a domicilio (no disponible)"}
                </option>
              </Select>
            )}
          </Campo>

          {modo === "ENVIO" && envioDisponible ? (
            <>
              <Campo
                etiqueta="Domicilio de entrega"
                nombre="domicilio"
                requerido
                valorInicial={valores.domicilio ?? ""}
                error={errores.domicilio}
              />
              <Campo
                etiqueta="Distancia hasta el local (en metros)"
                nombre="distanciaMetros"
                inputMode="numeric"
                requerido
                ayuda={`Distancia por calle desde el local. Máximo: ${config.distanciaMaximaMetros} m.`}
                valorInicial={valores.distanciaMetros ?? ""}
                error={errores.distanciaMetros}
              />
              {estimacion ? (
                <p role="status" className="text-sm">
                  {estimacion.tipo === "costo"
                    ? `Costo de envío estimado: ${formatearPrecio(estimacion.centavos)}. Es una estimación: el costo final lo calcula el servidor al confirmar.`
                    : estimacion.tipo === "fuera-de-cobertura"
                      ? `Esa distancia está fuera de la zona de envío (máximo ${config.distanciaMaximaMetros} m). Revisala o elegí retiro en el local.`
                      : "No tenemos una tarifa cargada para esa distancia. El servidor la verifica al confirmar."}
                </p>
              ) : null}
            </>
          ) : null}
        </fieldset>

        <fieldset className="flex flex-col gap-4">
          <legend className="text-lg font-semibold">Pago</legend>
          <Campo
            etiqueta="Medio de pago"
            nombre="medioPago"
            valorInicial={
              efectivoDisponible && valores.medioPago === "EFECTIVO" ? "EFECTIVO" : "MERCADO_PAGO"
            }
            error={errores.medioPago}
            ayuda={
              efectivoDisponible
                ? "El efectivo se paga en el local al retirar."
                : "El pago en efectivo solo está disponible para retiro en el local."
            }
          >
            {(control) => (
              // key: si el efectivo aparece o desaparece, el control se rearma con un valor válido.
              <Select key={String(efectivoDisponible)} {...control}>
                <option value="MERCADO_PAGO">Mercado Pago</option>
                {efectivoDisponible ? <option value="EFECTIVO">Efectivo al retirar</option> : null}
              </Select>
            )}
          </Campo>
        </fieldset>

        <ResumenPedido
          vista={vista}
          envio={modo === "ENVIO" ? (estimacion ?? { tipo: "sin-tarifa" }) : null}
        />

        <Card>
          <CardContent className="flex flex-col gap-3">
            <Button type="submit" disabled={pendiente} className="w-full sm:w-auto sm:self-end">
              {pendiente ? "Confirmando…" : "Confirmar pedido"}
            </Button>
            <p className="text-sm text-muted-foreground">
              Al confirmar reservamos tus productos. Con Mercado Pago tenés 15 minutos para pagar.
            </p>
          </CardContent>
        </Card>
      </form>
    </div>
  );
}
