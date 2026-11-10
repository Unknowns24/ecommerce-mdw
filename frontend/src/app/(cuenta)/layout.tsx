import type { ReactNode } from "react";
import { requerirSesion } from "@/features/identidad/sesion";

export default async function LayoutCuenta({ children }: { children: ReactNode }) {
  await requerirSesion();
  return children;
}
