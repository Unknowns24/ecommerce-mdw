import "server-only";

import { api } from "@/lib/api";

import type { PaginaPedidos, PedidoAdmin } from "./tipos";

export const TAMANO_PAGINA = 20;

export type FiltrosPedidos = { estado: string; desde: string; hasta: string; pagina: number };

export function listarPedidos(f: FiltrosPedidos) {
  const q = new URLSearchParams({ pagina: String(f.pagina), tamano: String(TAMANO_PAGINA) });
  if (f.estado) q.set("estado", f.estado);
  if (f.desde) q.set("desde", f.desde);
  if (f.hasta) q.set("hasta", f.hasta);
  return api<PaginaPedidos>(`/api/admin/pedidos?${q}`);
}

export function obtenerPedido(id: string) {
  return api<PedidoAdmin>(`/api/admin/pedidos/${encodeURIComponent(id)}`);
}
