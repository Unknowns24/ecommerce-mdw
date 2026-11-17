import "server-only";

import { api } from "@/lib/api";

import type { Rol, Usuario } from "./tipos";

export const POR_PAGINA_USUARIOS = 25;

export async function listarUsuarios(pagina: number): Promise<Usuario[]> {
  const q = new URLSearchParams({ limit: String(POR_PAGINA_USUARIOS), offset: String((pagina - 1) * POR_PAGINA_USUARIOS) });
  return (await api<{ items: Usuario[] }>(`/api/admin/usuarios?${q}`)).items ?? [];
}

export async function listarRoles(): Promise<Rol[]> {
  return (await api<{ items: Rol[] }>("/api/admin/roles?limit=100")).items ?? [];
}
