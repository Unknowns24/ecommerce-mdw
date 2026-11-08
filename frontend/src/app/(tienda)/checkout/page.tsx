import type { Metadata } from "next";
import { SinDatos } from "@/components/sin-datos";
import { verCarrito } from "@/features/carrito/api";
import { obtenerCuenta } from "@/features/identidad/sesion";
import { leerConfiguracion } from "@/features/pedidos/api";
import { CheckoutInvitado } from "@/features/pedidos/components/checkout-invitado";
import { FormularioCheckout } from "@/features/pedidos/components/formulario-checkout";

export const metadata: Metadata = { title: "Checkout" };

// /checkout (H13, H14, H15). Sesión opcional, dos compradores:
// - cliente: el carrito vive en la API; se lee acá, en el servidor, y la
//   acción lo vuelve a leer (el navegador no manda ítems).
// - invitado: el carrito vive en localStorage; lo lee y valoriza un Client
//   Component, y la acción recibe solo {varianteId, unidades}.
//
// Los cinco estados: cargando (loading.tsx), con datos (el formulario), vacío
// (carrito sin productos: ofrece ir al catálogo), error (error.tsx). «Sin
// resultados» no aplica: la pantalla no tiene búsqueda ni filtros.
export default async function PaginaCheckout() {
  const cuenta = await obtenerCuenta();
  // Independientes entre sí: en paralelo (ADR-007).
  const [config, vista] = await Promise.all([leerConfiguracion(), cuenta ? verCarrito() : null]);

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
      <h1 className="text-2xl font-semibold">Checkout</h1>
      {cuenta ? (
        vista && vista.lineas.length > 0 ? (
          <FormularioCheckout
            config={config}
            inicial={{ nombre: cuenta.nombre, apellido: cuenta.apellido, correo: cuenta.correo }}
            vista={vista}
            itemsInvitado={null}
          />
        ) : (
          <SinDatos
            titulo="Tu carrito está vacío"
            descripcion="Para hacer un pedido primero agregá productos desde el catálogo."
            accion={{ texto: "Ir al catálogo", href: "/" }}
          />
        )
      ) : (
        <CheckoutInvitado config={config} />
      )}
    </main>
  );
}
