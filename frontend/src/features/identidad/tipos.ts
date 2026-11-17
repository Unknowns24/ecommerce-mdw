// Tipos de identidad, a partir de docs/api/identidad.http.

export type Usuario = {
  id: string;
  nombre: string;
  apellido: string;
  correo: string;
  telefono?: string | null;
  estado: "ACTIVO" | "INACTIVO";
  origen: "ADMINISTRATIVO" | "PUBLICO";
  es_dueno_inicial: boolean;
  creado_en: string;
  roles?: string[];
};

export type Rol = { id: string; nombre: string; creado_en: string; permisos: string[] };

export const NOMBRES_PERMISOS: Record<string, string> = {
  "catalogo.escribir": "Catálogo: productos, marcas y categorías",
  "stock.gestionar": "Stock: proveedores, lotes y ajustes",
  "pedidos.gestionar": "Pedidos: despacho, entrega, cobro y cancelación",
  "usuarios.gestionar": "Cuentas administrativas y roles",
};
