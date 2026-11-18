// Fotos de campaña de la portada. Hoy no hay ninguna: los bloques se
// muestran como campos de color con tipografía. Cuando el negocio tenga fotos
// propias (nunca de otra marca), se guardan en public/campanas/ y se listan
// acá: el bloque correspondiente pasa a mostrarlas sin tocar otro archivo.
// Medidas recomendadas en DESIGN.md, sección 12.

// Foto del hero de la portada, por ejemplo "/campanas/portada.webp".
export const FOTO_PORTADA: string | undefined = undefined;

// Foto por categoría, con el id de la categoría como clave.
export const FOTOS_CATEGORIA: Record<string, string> = {};
