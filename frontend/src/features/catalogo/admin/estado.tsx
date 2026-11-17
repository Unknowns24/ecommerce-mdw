import { Badge } from "@/components/ui/badge";

import { activo, type Estado } from "./tipos";

// Estado con texto, nunca solo con color.
export function EstadoCatalogo({ estado }: { estado: Estado }) {
  return activo(estado) ? <Badge variant="secondary">Activo</Badge> : <Badge variant="outline">Inactivo</Badge>;
}
