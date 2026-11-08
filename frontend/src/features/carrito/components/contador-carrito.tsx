import { obtenerCuenta } from "@/features/identidad/sesion";
import { unidadesDelCliente } from "../api";
import { ContadorInvitado } from "./contador-invitado";

// Server Component: decide de dónde sale la cantidad según haya sesión o no.
export async function ContadorCarrito() {
  const cuenta = await obtenerCuenta();
  if (!cuenta) return <ContadorInvitado />;
  let unidades = 0;
  try {
    unidades = await unidadesDelCliente();
  } catch (e) {
    // El encabezado no se cae por el carrito: queda solo el link.
    console.error("contador del carrito", e);
  }
  return unidades > 0 ? <span>({unidades})</span> : null;
}
