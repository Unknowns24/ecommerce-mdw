export type CuerpoError = { error: string; detalles?: unknown };

export class ErrorApi extends Error {
  constructor(public status: number, public cuerpo: CuerpoError) {
    super(cuerpo.error);
  }
}
