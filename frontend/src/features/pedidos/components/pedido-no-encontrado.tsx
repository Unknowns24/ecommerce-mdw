import { SinDatos } from "@/components/sin-datos";

// El mismo texto para «no existe» y «no es tuyo»: el 404 no cuenta de más
// (spec, sección 7) y nunca nombra un permiso.
export function PedidoNoEncontrado({ accion }: { accion: { texto: string; href: string } }) {
  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
      <SinDatos
        titulo="No encontramos ese pedido"
        descripcion="Revisá que la dirección esté completa y bien copiada."
        accion={accion}
      />
    </div>
  );
}
