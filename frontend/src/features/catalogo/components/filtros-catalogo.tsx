import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";

import { ORDENES, type FiltrosCatalogo, type MarcaPublica } from "../tipos";

// Formulario GET: búsqueda, filtros y orden se combinan en la query string y
// funcionan sin JavaScript. Cambiar un filtro vuelve a la página 1.
export function FiltrosCatalogoForm({
  filtros,
  marcas,
}: {
  filtros: FiltrosCatalogo;
  marcas: MarcaPublica[];
}) {
  return (
    <form action="/" className="grid grid-cols-2 gap-3 lg:grid-cols-1" aria-label="Filtrar el catálogo">
      {/* La búsqueda se escribe en el encabezado; acá solo se conserva al
          cambiar filtros u orden. */}
      {filtros.q ? <input type="hidden" name="q" value={filtros.q} /> : null}
      {/* La categoría se elige con los accesos sobre la grilla; acá se conserva. */}
      {filtros.categoria ? <input type="hidden" name="categoria" value={filtros.categoria} /> : null}
      <div className="col-span-2 grid gap-1.5 lg:col-span-1">
        <Label htmlFor="filtro-marca">Marca</Label>
        <Select id="filtro-marca" name="marca" defaultValue={filtros.marca}>
          <option value="">Todas</option>
          {marcas.map((m) => (
            <option key={m.id} value={m.id}>
              {m.nombre}
            </option>
          ))}
        </Select>
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="filtro-orden">Ordenar por</Label>
        <Select id="filtro-orden" name="orden" defaultValue={filtros.orden}>
          {Object.entries(ORDENES).map(([valor, texto]) => (
            <option key={valor} value={valor}>
              {texto}
            </option>
          ))}
        </Select>
      </div>
      <div className="flex flex-col justify-end gap-2">
        <Button type="submit" className="w-full">
          Aplicar
        </Button>
      </div>
    </form>
  );
}
