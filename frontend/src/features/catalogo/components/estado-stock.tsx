import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

// «SIN STOCK» con texto, nunca solo con color (H04, sección 7). Una etiqueta
// recta con borde, que se lee sobre la foto y sobre el fondo blanco.
export function SinStock({ className }: { className?: string }) {
  return (
    <Badge
      variant="outline"
      className={cn("h-6 border-foreground bg-background/90 px-2 text-[0.6875rem] font-normal tracking-[0.12em] text-foreground", className)}
    >
      SIN STOCK
    </Badge>
  );
}
