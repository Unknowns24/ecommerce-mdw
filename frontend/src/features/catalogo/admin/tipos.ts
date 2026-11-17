// Tipos del catálogo administrativo, a partir de docs/api/catalogo.http.

export type Estado = "ACTIVO" | "INACTIVO" | "ACTIVA" | "INACTIVA";

export type Producto = {
  id: string;
  marcaId: string;
  nombre: string;
  descripcion: string;
  estado: Estado;
  metodologiaRotacion: "FIFO" | "LIFO";
  categorias?: string[];
  variantes?: VarianteAdmin[];
};

export type VarianteAdmin = {
  id: string;
  productoId: string;
  codigo: string;
  nombre: string;
  descripcion: string;
  precioMinoristaCentavos: number;
  precioMayoristaCentavos: number;
  estado: Estado;
};

export type Marca = { id: string; nombre: string; estado: Estado };
export type Categoria = { id: string; nombre: string; padreId?: string | null; estado: Estado };
export type Imagen = { id: string; varianteId: string; referenciaImagen: string; esPrincipal: boolean; orden: number };

export type Pagina<T> = { items: T[]; pagina: number; porPagina: number; total: number };

export const activo = (e: Estado) => e === "ACTIVO" || e === "ACTIVA";
