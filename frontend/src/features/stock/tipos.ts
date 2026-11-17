// Tipos de inventario, a partir de docs/api/inventario.http.

export type Proveedor = { id: string; nombre: string; correo?: string; telefono?: string; web?: string; creadoEn: string };

export type Lote = {
  id: string;
  varianteId: string;
  proveedorId: string;
  unidadesIngresadas: number;
  costoUnitarioCentavos: number;
  fechaIngreso: string;
};

export type StockVariante = { varianteId: string; disponible: number; lotes: Lote[] };

export type Movimiento = { id: string; loteId: string; tipo: string; unidades: number; motivo: string; creadoEn: string };

export const TIPOS_MOVIMIENTO: Record<string, string> = {
  INGRESO: "Ingreso de lote",
  AJUSTE: "Ajuste",
  SALIDA: "Salida por venta",
};
