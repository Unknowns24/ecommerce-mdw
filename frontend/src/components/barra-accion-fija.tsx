import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

// La acción principal fija abajo en el teléfono, al alcance del pulgar
// (DESIGN.md, sección 5). Desde md vuelve a su lugar en la página. La
// pantalla que la usa deja espacio abajo (pb) para no tapar contenido.
export function BarraAccionFija({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div
      className={cn(
        "fixed inset-x-0 bottom-0 z-10 border-t bg-background px-4 py-3",
        "md:static md:z-auto md:border-0 md:bg-transparent md:p-0",
        className,
      )}
    >
      {children}
    </div>
  );
}
