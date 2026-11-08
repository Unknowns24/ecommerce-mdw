"use client";

import { useEffect, useMemo, useState, useSyncExternalStore } from "react";
import { ErrorDeCarga } from "@/components/error-de-carga";
import { SinDatos } from "@/components/sin-datos";
import { valorizarInvitadoAction } from "@/features/carrito/acciones";
import {
  instantaneaCarritoInvitado,
  instantaneaServidor,
  suscribirseAlCarritoInvitado,
  unidadesDeInstantanea,
} from "@/features/carrito/almacen-invitado";
import type { VistaCarrito } from "@/features/carrito/tipos";
import { sanearItems } from "../formulario";
import type { ConfiguracionTienda, ItemPedido } from "../tipos";
import { EsqueletoCheckout } from "./esqueleto-checkout";
import { FormularioCheckout } from "./formulario-checkout";

type Resultado = { crudo: string; vista: VistaCarrito } | { crudo: string; fallo: true };

const sinSuscripcion = () => () => {};

function itemsDe(crudo: string): ItemPedido[] {
  try {
    return sanearItems(JSON.parse(crudo));
  } catch {
    return [];
  }
}

// Checkout de quien no tiene sesión. Su carrito vive en localStorage
// ({varianteId, unidades}); acá se lee, se valoriza con la API (precios y
// disponibilidad salen del servidor) y se lo pasa al formulario, que lo manda
// en la Server Action. Nunca viaja un precio.
export function CheckoutInvitado({ config }: { config: ConfiguracionTienda | null }) {
  const crudo = useSyncExternalStore(
    suscribirseAlCarritoInvitado,
    instantaneaCarritoInvitado,
    instantaneaServidor,
  );
  const montado = useSyncExternalStore(sinSuscripcion, () => true, () => false);
  const hayItems = unidadesDeInstantanea(crudo) > 0;

  const [resultado, setResultado] = useState<Resultado | null>(null);
  const [intento, setIntento] = useState(0);

  useEffect(() => {
    if (!montado || !hayItems) return;
    let vigente = true;
    valorizarInvitadoAction(itemsDe(crudo))
      .then((vista) => {
        if (vigente) setResultado({ crudo, vista });
      })
      .catch((e: unknown) => {
        console.error("valorizar el carrito del invitado", e);
        if (vigente) setResultado({ crudo, fallo: true });
      });
    return () => {
      vigente = false;
    };
  }, [crudo, montado, hayItems, intento]);

  // Los ítems que se mandan son los que se valorizaron y se están mostrando.
  const items = useMemo(() => (resultado ? itemsDe(resultado.crudo) : []), [resultado]);

  if (!montado) return <EsqueletoCheckout />;
  if (resultado === null) {
    return hayItems ? (
      <EsqueletoCheckout />
    ) : (
      <SinDatos
        titulo="Tu carrito está vacío"
        descripcion="Para hacer un pedido primero agregá productos desde el catálogo."
        accion={{ texto: "Ir al catálogo", href: "/" }}
      />
    );
  }
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

  // Una vez que el formulario está en pantalla se queda, aunque el carrito
  // local se vacíe porque el pedido se creó: conserva su estado (por ejemplo,
  // el panel «pedido creado, pago sin iniciar»).
  return (
    <FormularioCheckout
      config={config}
      inicial={{ nombre: "", apellido: "", correo: "" }}
      vista={resultado.vista}
      itemsInvitado={items}
    />
  );
}
