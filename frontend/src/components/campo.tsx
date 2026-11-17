import { useId, type ReactElement, type ReactNode } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type PropsControl = {
  id: string;
  name: string;
  "aria-describedby"?: string;
  "aria-invalid"?: boolean;
  defaultValue?: string;
};

export function Campo({
  etiqueta,
  nombre,
  error,
  ayuda,
  tipo = "text",
  inputMode,
  requerido,
  valorInicial,
  children,
}: {
  etiqueta: string;
  nombre: string;
  error?: string;
  ayuda?: string;
  tipo?: string;
  inputMode?: "numeric" | "text" | "email" | "tel";
  requerido?: boolean;
  valorInicial?: string;
  children?: (control: PropsControl) => ReactNode;
}): ReactElement {
  const id = useId();
  const idAyuda = ayuda ? `${id}-ayuda` : undefined;
  const idError = error ? `${id}-error` : undefined;
  const control = {
    id,
    name: nombre,
    "aria-describedby": [idAyuda, idError].filter(Boolean).join(" ") || undefined,
    "aria-invalid": error ? true : undefined,
    defaultValue: valorInicial,
    required: requerido,
  };

  return (
    <div className="grid gap-2">
      <Label htmlFor={id}>{etiqueta}</Label>
      {children ? children(control) : <Input {...control} type={tipo} inputMode={inputMode} />}
      {ayuda && <p id={idAyuda} className="text-sm text-muted-foreground">{ayuda}</p>}
      {error && <p id={idError} className="text-sm text-destructive">{error}</p>}
    </div>
  );
}
