import { notFound, redirect } from "next/navigation";
import { api, ErrorApi } from "@/lib/api";

export type Cuenta = {
  id: string;
  nombre: string;
  apellido: string;
  correo: string;
  roles: string[];
  permisos: string[];
};

export async function obtenerCuenta(): Promise<Cuenta | null> {
  try {
    return await api<Cuenta>("/api/cuenta");
  } catch (error) {
    if (error instanceof ErrorApi && error.status === 401) {
      return null;
    }
    throw error;
  }
}

export async function requerirSesion(): Promise<Cuenta> {
  const cuenta = await obtenerCuenta();
  if (!cuenta) {
    redirect("/ingresar");
  }
  return cuenta;
}

export async function requerirPermiso(permiso: string): Promise<Cuenta> {
  const cuenta = await requerirSesion();
  if (!cuenta.permisos.includes(permiso)) {
    notFound();
  }
  return cuenta;
}
