"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { traducirError, valoresDe } from "@/lib/acciones";
import { ErrorApi } from "@/lib/api";
import { obtenerCuenta } from "@/features/identidad/sesion";
import { vaciarCarrito, verCarrito } from "@/features/carrito/api";
import { guardarTokenAcceso, leerTokenAcceso } from "@/features/pedidos/acceso";
import { crearPedido, iniciarPago } from "@/features/pedidos/api";
import { aCuerpoPedido, sanearItems, ubicarErrores } from "@/features/pedidos/formulario";
import type { EstadoCheckout, PedidoCreado, PedidoSinPago } from "@/features/pedidos/tipos";

// Server Actions del checkout (H13, H14, H15). Sesión OPCIONAL: el invitado
// compra igual, así que acá la sesión no se exige; se mira para decidir de
// dónde salen los ítems y cómo se identifica quien paga.

function refrescarPantallasViejas() {
  // El carrito quedó vacío (/carrito y el contador del encabezado) y hay un
  // pedido nuevo en «Mis pedidos».
  revalidatePath("/", "layout");
  revalidatePath("/carrito");
  revalidatePath("/cuenta/pedidos");
}

/** initPoint viene de nuestra API, pero solo se redirige afuera a https. */
function esUrlSegura(destino: string): boolean {
  try {
    return new URL(destino).protocol === "https:";
  } catch {
    return false;
  }
}

/**
 * «Confirmar». `itemsInvitado` entra por .bind desde el Client Component que
 * lee localStorage y solo se usa SIN sesión (se sanea: {varianteId, unidades}).
 * Con sesión los ítems los lee el servidor del carrito de la API: el navegador
 * no manda nada de eso.
 */
export async function confirmarPedidoAction(
  itemsInvitado: unknown,
  _prev: EstadoCheckout,
  formData: FormData,
): Promise<EstadoCheckout> {
  const cuenta = await obtenerCuenta();
  const valores = valoresDe(formData);

  const items = cuenta
    ? (await verCarrito()).lineas.map((l) => ({ varianteId: l.varianteId, unidades: l.unidades }))
    : sanearItems(itemsInvitado);
  if (items.length === 0) {
    return {
      enviado: true,
      mensaje: "Tu carrito está vacío. Agregá productos desde el catálogo para hacer un pedido.",
      valores,
    };
  }

  const cuerpo = aCuerpoPedido(valores, items);

  let pedido: PedidoCreado;
  try {
    pedido = await crearPedido(cuerpo);
  } catch (e) {
    if (e instanceof ErrorApi && e.status === 400) {
      // 400 → debajo de cada campo; lo que no tiene campo va arriba.
      const { porCampo, generales } = ubicarErrores(
        (e.cuerpo.detalles as { campos?: Record<string, string> } | undefined)?.campos,
      );
      return {
        enviado: true,
        errores: porCampo,
        mensaje: generales.length > 0 ? generales.join(" ") : undefined,
        valores,
      };
    }
    // 409 → arriba, con variante → disponibles en `detalles`; 401/404 y el resto: ADR-007.
    return traducirError(e, valores);
  }

  // El pedido ya existe y reservó stock. De acá en adelante nada puede volver a
  // pedirle a la persona que «Confirme» de nuevo: crearía un segundo pedido.
  const privado = !cuenta;
  const enlace = cuenta ? `/cuenta/pedidos/${pedido.id}` : `/seguimiento/${pedido.tokenAcceso}`;

  // La API no vacía el carrito al crear el pedido. El del cliente se vacía acá
  // (mejor esfuerzo: un fallo no deshace el pedido); el del invitado, en el
  // navegador, después de este redirect.
  if (cuenta) {
    try {
      await vaciarCarrito();
    } catch (e) {
      console.error("vaciar el carrito después del pedido", e);
    }
  }
  refrescarPantallasViejas();

  if (cuerpo.medioPago === "EFECTIVO") {
    // Nace CONFIRMADO con pago pendiente: se va derecho al detalle (spec, sección 5).
    redirect(enlace);
  }

  // Mercado Pago. El invitado guarda su token en una cookie propia ANTES de
  // iniciar el pago, para que /pago/retorno lo pueda llevar a su pedido.
  if (privado) await guardarTokenAcceso(pedido.tokenAcceso);

  const sinPago: PedidoSinPago = { id: pedido.id, numero: pedido.numero, enlace, privado };
  let initPoint: string;
  try {
    ({ initPoint } = await iniciarPago(pedido.id, privado ? pedido.tokenAcceso : undefined));
  } catch (e) {
    return { ...traducirError(e, valores), pedido: sinPago };
  }
  if (!esUrlSegura(initPoint)) {
    console.error("initPoint inesperado", initPoint);
    return {
      enviado: true,
      mensaje: "No pudimos iniciar el pago, probá de nuevo en unos minutos.",
      pedido: sinPago,
      valores,
    };
  }
  redirect(initPoint);
}

/**
 * «Reintentar el pago» de un pedido que ya existe. El id sale de la acción
 * ligada con .bind, no de un campo; el invitado se identifica con la cookie.
 * Un 409 («ya hay un pago en curso», «la reserva venció», «no pudimos iniciar
 * el pago») vuelve arriba y el pedido NO se da por pagado.
 */
export async function reintentarPagoAction(pedido: PedidoSinPago): Promise<EstadoCheckout> {
  const cuenta = await obtenerCuenta();
  const token = cuenta ? undefined : await leerTokenAcceso();
  if (!cuenta && !token) {
    return {
      enviado: true,
      mensaje:
        "Pasó demasiado tiempo para reintentar desde esta pantalla. Abrí tu enlace privado para ver el estado del pedido.",
      pedido,
    };
  }

  let initPoint: string;
  try {
    ({ initPoint } = await iniciarPago(pedido.id, token));
  } catch (e) {
    return { ...traducirError(e, {}), pedido };
  }
  if (!esUrlSegura(initPoint)) {
    console.error("initPoint inesperado", initPoint);
    return {
      enviado: true,
      mensaje: "No pudimos iniciar el pago, probá de nuevo en unos minutos.",
      pedido,
    };
  }
  redirect(initPoint);
}
