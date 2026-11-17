package pedidos

import (
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

// ConfiguracionTiendaResponse es lo que sale por GET /api/tienda/configuracion,
// que es público a propósito: lo justo para que el checkout ofrezca efectivo y
// anticipe el costo del envío. Los campos están enumerados a mano (sin id ni
// fechas de la configuración): lo que no se agrega acá, no se publica.
type ConfiguracionTiendaResponse struct {
	EfectivoHabilitado    bool                      `json:"efectivoHabilitado"`
	DistanciaMaximaMetros int                       `json:"distanciaMaximaMetros"`
	Tarifas               []TarifaDistanciaResponse `json:"tarifas"`
}

// TarifaDistanciaResponse es un rango de distancia con su costo de envío, en
// centavos (ADR-004).
type TarifaDistanciaResponse struct {
	DesdeMetros   int   `json:"desdeMetros"`
	HastaMetros   int   `json:"hastaMetros"`
	CostoCentavos int64 `json:"costoCentavos"`
}

func AConfiguracionTienda(c uc.ConfiguracionVigente) ConfiguracionTiendaResponse {
	// make con largo 0 (no un slice nil): sin tarifas el JSON es [] y no null.
	tarifas := make([]TarifaDistanciaResponse, len(c.Tarifas))
	for i, t := range c.Tarifas {
		tarifas[i] = TarifaDistanciaResponse{DesdeMetros: t.DesdeMetros, HastaMetros: t.HastaMetros, CostoCentavos: t.CostoCentavos}
	}
	return ConfiguracionTiendaResponse{
		EfectivoHabilitado:    c.Configuracion.EfectivoHabilitado,
		DistanciaMaximaMetros: c.Configuracion.DistanciaMaximaMetros,
		Tarifas:               tarifas,
	}
}
