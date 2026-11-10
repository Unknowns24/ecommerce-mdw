import { Button } from "@/components/ui/button";

export function BotonAgregarAlCarrito({ disponible }: { varianteId: string; disponible: boolean }) {
  return <Button type="button" disabled>{disponible ? "Agregar al carrito" : "SIN STOCK"}</Button>;
}
