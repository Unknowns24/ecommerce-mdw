// Los módulos del panel salen de la tabla «Otras pantallas» y del flujo
// principal del mapa de pantallas (spec, sección 5). Esconder un acceso es
// cortesía: la operación la sigue verificando el backend.

export type Modulo = { texto: string; href: string; descripcion: string; permiso: string; soloDueno?: boolean };

export const MODULOS_PANEL: Modulo[] = [
  { texto: "Pedidos", href: "/admin/pedidos", descripcion: "Despacho, entrega, cobro y cancelación.", permiso: "pedidos.gestionar" },
  { texto: "Productos", href: "/admin/productos", descripcion: "Productos, variantes, precios e imágenes.", permiso: "catalogo.escribir" },
  { texto: "Marcas", href: "/admin/marcas", descripcion: "Alta, edición y baja de marcas.", permiso: "catalogo.escribir" },
  { texto: "Categorías", href: "/admin/categorias", descripcion: "Categorías y subcategorías.", permiso: "catalogo.escribir" },
  { texto: "Proveedores", href: "/admin/proveedores", descripcion: "Listado y alta de proveedores.", permiso: "stock.gestionar" },
  { texto: "Ingreso de lote", href: "/admin/stock/lotes/nuevo", descripcion: "Unidades nuevas de una variante.", permiso: "stock.gestionar" },
  { texto: "Cuentas", href: "/admin/usuarios", descripcion: "Cuentas administrativas y sus roles.", permiso: "usuarios.gestionar", soloDueno: true },
  { texto: "Roles", href: "/admin/roles", descripcion: "Roles y sus permisos.", permiso: "usuarios.gestionar", soloDueno: true },
];

export function modulosDe(cuenta: { roles: string[]; permisos: string[] }): Modulo[] {
  return MODULOS_PANEL.filter((m) => cuenta.permisos.includes(m.permiso) && (!m.soloDueno || cuenta.roles.includes("dueno")));
}
