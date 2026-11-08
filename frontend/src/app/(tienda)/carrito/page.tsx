import type { Metadata } from "next";
import { obtenerCuenta } from "@/features/identidad/sesion";
import { verCarrito } from "@/features/carrito/api";
import { CarritoCliente } from "@/features/carrito/components/carrito-cliente";
import { CarritoInvitado } from "@/features/carrito/components/carrito-invitado";

export const metadata: Metadata = { title: "Carrito" };

// /carrito (SPEC-H07). Sesión opcional, dos compradores:
// - cliente: el carrito vive en la API; se lee acá, en el servidor.
// - invitado: vive en localStorage; lo lee y valoriza un Client Component.
//
// Los cinco estados: cargando (loading.tsx), con datos y vacío (acá y en
// CarritoCliente / CarritoInvitado), error (error.tsx). «Sin resultados» no
// aplica: la pantalla no tiene búsqueda ni filtros.
export default async function PaginaCarrito() {
  const cuenta = await obtenerCuenta();
  // Si verCarrito falla, el error sube a error.tsx. La API decide: esta
  // pantalla no autoriza nada por tener o no cuenta.
  const vista = cuenta ? await verCarrito() : null;

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
      <h1 className="text-2xl font-semibold">Tu carrito</h1>
      {vista ? <CarritoCliente vista={vista} /> : <CarritoInvitado />}
    </main>
  );
}
