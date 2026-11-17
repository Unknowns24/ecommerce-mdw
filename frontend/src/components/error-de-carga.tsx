import type { ReactElement } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

export function ErrorDeCarga({ reintentar }: { reintentar: () => void }): ReactElement {
  return (
    <section className="grid gap-4">
      <Alert variant="destructive">
        <AlertTitle>No pudimos cargar esta página</AlertTitle>
        <AlertDescription>Hubo un problema en el sistema. Tus datos siguen guardados.</AlertDescription>
      </Alert>
      <Button type="button" onClick={reintentar} className="w-full sm:justify-self-start sm:w-auto">
        Reintentar
      </Button>
    </section>
  );
}
