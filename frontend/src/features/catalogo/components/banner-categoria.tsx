import { CampoCampana } from "./campo-campana";

// Banner a todo el ancho con el título del catálogo (la categoría, la marca
// o la búsqueda) sobre una franja blanca translúcida. Si la categoría tiene
// foto de campaña (campanas.ts), va de fondo; si no, un campo humo.
export function BannerCategoria({ titulo, foto }: { titulo: string; foto?: string }) {
  return (
    <div data-ancho="completo" className="-mt-6">
      <CampoCampana foto={foto} color="bg-muted text-foreground" className="grid h-36 place-items-center px-4 sm:h-56">
        <h1 className="w-full max-w-3xl bg-background/75 px-4 py-3 text-center text-2xl leading-tight font-light tracking-[0.2em] break-words text-foreground uppercase sm:py-4 sm:text-3xl">
          {titulo}
        </h1>
      </CampoCampana>
    </div>
  );
}
