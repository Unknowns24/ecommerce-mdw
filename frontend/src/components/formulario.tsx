"use client";

import { useActionState } from "react";

import { Campo } from "@/components/campo";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { SIN_ENVIAR, type EstadoFormulario } from "@/lib/acciones";
import { cn } from "@/lib/utils";

// Formulario genérico de una Server Action (sección 7, «Formularios»):
// 409 arriba, 400 debajo de cada campo, no se vacía al fallar y el botón se
// deshabilita diciendo qué pasa. No sabe nada del negocio: los campos llegan
// como datos desde la página.

export type DefCampo = {
  nombre: string;
  etiqueta: string;
  tipo?: "text" | "email" | "tel" | "url" | "password" | "date" | "textarea" | "select" | "casillas";
  inputMode?: "numeric" | "text" | "email" | "tel";
  requerido?: boolean;
  ayuda?: string;
  valorInicial?: string; // en casillas: valores separados por coma
  opciones?: { valor: string; texto: string }[];
  ancho?: "completo";
};

type Accion = (prev: EstadoFormulario, formData: FormData) => Promise<EstadoFormulario>;

export function Formulario({
  accion,
  campos,
  textoEnviar,
  textoEnviando,
  columnas = 1,
  varianteBoton = "default",
  detalle,
}: {
  accion: Accion;
  campos: DefCampo[];
  textoEnviar: string;
  textoEnviando: string;
  columnas?: 1 | 2;
  varianteBoton?: "default" | "outline" | "destructive";
  /** Texto extra para el 409, si la historia pide enumerar algo. */
  detalle?: (detalles: unknown) => string | null;
}) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  const valores = estado.valores;
  const errores = estado.errores ?? {};
  const extra = estado.mensaje && detalle ? detalle(estado.detalles) : null;
  const errorGeneral = errores.body ?? errores.id;

  return (
    <form action={enviar} noValidate className={cn("grid gap-4", columnas === 2 && "sm:grid-cols-2")}>
      {estado.mensaje || errorGeneral ? (
        <Alert variant="destructive" role="alert" className={cn(columnas === 2 && "sm:col-span-2")}>
          <AlertDescription>
            {estado.mensaje ?? errorGeneral}
            {extra ? ` ${extra}` : null}
          </AlertDescription>
        </Alert>
      ) : null}

      {campos.map((c) => {
        const valor = valores ? (valores[c.nombre] ?? "") : c.valorInicial;
        const clase = cn(c.ancho === "completo" && columnas === 2 && "sm:col-span-2");
        if (c.tipo === "casillas") {
          const marcados = (valor ?? "").split(",").filter(Boolean);
          const idError = `${c.nombre}-error`;
          return (
            <fieldset key={c.nombre} className={cn("grid gap-2", clase)} aria-describedby={errores[c.nombre] ? idError : undefined}>
              <legend className="mb-1 text-sm font-medium">{c.etiqueta}</legend>
              {(c.opciones ?? []).map((o) => (
                <label key={o.valor} className="flex items-center gap-2 text-sm">
                  <input type="checkbox" name={c.nombre} value={o.valor} defaultChecked={marcados.includes(o.valor)} className="size-4 accent-primary" />
                  {o.texto}
                </label>
              ))}
              {c.ayuda ? <p className="text-sm text-muted-foreground">{c.ayuda}</p> : null}
              {errores[c.nombre] ? (
                <p id={idError} className="text-sm text-destructive">
                  {errores[c.nombre]}
                </p>
              ) : null}
            </fieldset>
          );
        }
        return (
          <div key={c.nombre} className={clase}>
            <Campo
              etiqueta={c.etiqueta}
              nombre={c.nombre}
              tipo={c.tipo === "select" || c.tipo === "textarea" ? undefined : c.tipo}
              inputMode={c.inputMode}
              requerido={c.requerido}
              ayuda={c.ayuda}
              error={errores[c.nombre]}
              valorInicial={valor}
            >
              {c.tipo === "select"
                ? (control) => (
                    <Select {...control}>
                      {(c.opciones ?? []).map((o) => (
                        <option key={o.valor} value={o.valor}>
                          {o.texto}
                        </option>
                      ))}
                    </Select>
                  )
                : c.tipo === "textarea"
                  ? (control) => <Textarea {...control} rows={3} />
                  : undefined}
            </Campo>
          </div>
        );
      })}

      <div className={cn(columnas === 2 && "sm:col-span-2")}>
        <Button type="submit" variant={varianteBoton} disabled={pendiente} className="w-full sm:w-auto">
          {pendiente ? textoEnviando : textoEnviar}
        </Button>
      </div>
    </form>
  );
}

// Un botón solo (desactivar, quitar, elegir principal…) con su 409 arriba.
export function BotonAccion({
  accion,
  texto,
  textoEnviando,
  variante = "outline",
  etiqueta,
}: {
  accion: Accion;
  texto: string;
  textoEnviando: string;
  variante?: "default" | "outline" | "destructive" | "ghost";
  etiqueta?: string;
}) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  return (
    <form action={enviar} className="grid gap-2">
      {estado.mensaje ? (
        <Alert variant="destructive" role="alert">
          <AlertDescription>{estado.mensaje}</AlertDescription>
        </Alert>
      ) : null}
      <Button type="submit" variant={variante} disabled={pendiente} aria-label={etiqueta}>
        {pendiente ? textoEnviando : texto}
      </Button>
    </form>
  );
}
