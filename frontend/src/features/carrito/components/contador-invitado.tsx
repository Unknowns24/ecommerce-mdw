"use client";

import { useSyncExternalStore } from "react";
import {
  instantaneaCarritoInvitado,
  instantaneaServidor,
  suscribirseAlCarritoInvitado,
  unidadesDeInstantanea,
} from "../almacen-invitado";

export function ContadorInvitado() {
  const crudo = useSyncExternalStore(
    suscribirseAlCarritoInvitado,
    instantaneaCarritoInvitado,
    instantaneaServidor,
  );
  const unidades = unidadesDeInstantanea(crudo);
  return unidades > 0 ? <span>({unidades})</span> : null;
}
