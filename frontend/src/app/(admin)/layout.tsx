import { notFound } from "next/navigation";
import type { ReactNode } from "react";
import { requerirSesion } from "@/features/identidad/sesion";

export default async function LayoutAdmin({ children }: { children: ReactNode }) {
  const cuenta = await requerirSesion();
  if (cuenta.permisos.length === 0) {
    notFound();
  }
  return children;
}
