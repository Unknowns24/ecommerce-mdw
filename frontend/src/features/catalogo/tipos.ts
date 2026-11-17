// Tipos del catálogo público, a partir de docs/api/catalogo.http y de lo que
// responde GET /api/catalogo/*.

export type VarianteResumen = {
  id: string;
  codigo: string;
  nombre: string;
  marca: string;
  precioMinoristaCentavos: number;
  imagenPrincipal: string;
  disponible: boolean;
};

export type PaginaVariantes = {
  items: VarianteResumen[];
  pagina: number;
  porPagina: number;
  total: number;
};

export type CategoriaPublica = { id: string; nombre: string; hijas?: CategoriaPublica[] };

export type MarcaPublica = { id: string; nombre: string };

export type DetalleVariante = VarianteResumen & {
  descripcion?: string;
  precioMayoristaCentavos?: number;
  imagenes: string[];
  categorias: { id: string; nombre: string }[];
  otrasVariantes: VarianteResumen[];
};

export const ORDENES = {
  nombre_asc: "Nombre (A a Z)",
  precio_asc: "Precio: menor a mayor",
  precio_desc: "Precio: mayor a menor",
} as const;

export type Orden = keyof typeof ORDENES;

export type FiltrosCatalogo = { q: string; categoria: string; marca: string; orden: Orden; pagina: number };
