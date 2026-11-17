import { Alert, AlertDescription } from "@/components/ui/alert";

// El 409 (o el rechazo que no es de un campo) va arriba del formulario.
export function AvisoFormulario({ mensaje }: { mensaje?: string }) {
  if (!mensaje) return null;
  return (
    <Alert variant="destructive" role="alert">
      <AlertDescription>{mensaje}</AlertDescription>
    </Alert>
  );
}
