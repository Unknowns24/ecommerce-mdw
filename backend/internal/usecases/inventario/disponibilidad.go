// Package inventario contains the inventory module's business rules and the
// stock service other modules consume.
package inventario

import (
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
)

// LoteDisponible is the data AsignarLotes needs about one lot: how many
// units it has in total (intake + adjustments - outbound) and how many of
// those are already committed to an active reservation.
type LoteDisponible struct {
	LoteID       uuid.UUID
	FechaIngreso time.Time
	Unidades     int // ingresadas + ajustes - salidas
	Reservadas   int // reservas ACTIVA sobre ese lote
}

// Asignacion is how many units of a specific lot cover part of a request.
type Asignacion struct {
	LoteID   uuid.UUID
	Unidades int
}

// ErrStockInsuficiente is returned by AsignarLotes when the lots together
// don't cover the requested units.
var ErrStockInsuficiente = errors.New("stock insuficiente")

// Disponible is the sellable quantity across every lot: existencias de lotes
// menos reservas activas. A lot fully committed to reservations contributes
// zero, never a negative amount.
func Disponible(lotes []LoteDisponible) int {
	total := 0
	for _, lote := range lotes {
		if libres := lote.Unidades - lote.Reservadas; libres > 0 {
			total += libres
		}
	}
	return total
}

// AsignarLotes decides which lots, and how many units of each, cover a
// request of unidades units, following FIFO (lotes más antiguos primero) or
// LIFO (lotes más recientes primero). A single request can span several
// lots. It does not mutate lotes nor assume any particular input order.
func AsignarLotes(lotes []LoteDisponible, unidades int, metodologia string) ([]Asignacion, error) {
	if unidades <= 0 {
		return nil, errors.New("unidades debe ser mayor a cero")
	}
	if Disponible(lotes) < unidades {
		return nil, ErrStockInsuficiente
	}

	ordenados := append([]LoteDisponible(nil), lotes...)
	sort.SliceStable(ordenados, func(i, j int) bool {
		if metodologia == "LIFO" {
			return ordenados[i].FechaIngreso.After(ordenados[j].FechaIngreso)
		}
		return ordenados[i].FechaIngreso.Before(ordenados[j].FechaIngreso)
	})

	var asignaciones []Asignacion
	restantes := unidades
	for _, lote := range ordenados {
		if restantes == 0 {
			break
		}
		libres := lote.Unidades - lote.Reservadas
		if libres <= 0 {
			continue
		}
		tomar := libres
		if tomar > restantes {
			tomar = restantes
		}
		asignaciones = append(asignaciones, Asignacion{LoteID: lote.LoteID, Unidades: tomar})
		restantes -= tomar
	}

	return asignaciones, nil
}
