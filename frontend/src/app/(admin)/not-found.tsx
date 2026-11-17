import { SinDatos } from "@/components/sin-datos";

// 404 del panel: «no existe» y «no te corresponde» dicen lo mismo.
export default function NoEncontradoPanel() {
  return (
    <SinDatos
      titulo="No encontramos esta página"
      descripcion="Puede que el enlace esté mal escrito o que ya no exista."
      accion={{ texto: "Volver al panel", href: "/admin" }}
    />
  );
}
