"use client";

import { useState } from "react";

import { Campo } from "@/components/campo";
import { ErrorDeCarga } from "@/components/error-de-carga";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";

// Las piezas que necesitan el navegador, para verlas en la galería.
export function DemoErrorDeCarga() {
  const [veces, setVeces] = useState(0);
  return (
    <div className="grid gap-2">
      <ErrorDeCarga reintentar={() => setVeces((v) => v + 1)} />
      <p className="text-sm text-muted-foreground" role="status">
        Reintentos: {veces}
      </p>
    </div>
  );
}

export function DemoCampos() {
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Campo etiqueta="Campo normal" nombre="demo-normal" ayuda="Texto de ayuda debajo del campo." />
      <Campo etiqueta="Campo con error" nombre="demo-error" valorInicial="no-es-un-correo" error="Ingresá un correo válido" tipo="email" requerido />
      <Campo etiqueta="DNI (teclado numérico)" nombre="demo-dni" inputMode="numeric" />
      <Campo etiqueta="Select nativo" nombre="demo-select">
        {(control) => (
          <Select {...control}>
            <option>Retiro en el local</option>
            <option>Envío a domicilio</option>
          </Select>
        )}
      </Campo>
      <Campo etiqueta="Área de texto" nombre="demo-textarea">
        {(control) => <Textarea {...control} rows={3} />}
      </Campo>
      <Campo etiqueta="Campo deshabilitado" nombre="demo-deshabilitado">
        {(control) => <input {...control} disabled className="h-8 rounded-lg border px-2.5 text-sm opacity-50" />}
      </Campo>
    </div>
  );
}
