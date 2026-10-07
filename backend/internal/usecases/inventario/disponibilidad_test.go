package inventario

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func lote(unidades, reservadas int, diasAtras int) LoteDisponible {
	return LoteDisponible{
		LoteID:       uuid.New(),
		FechaIngreso: time.Now().AddDate(0, 0, -diasAtras),
		Unidades:     unidades,
		Reservadas:   reservadas,
	}
}

func TestAsignarLotesAlcanzaConUnLote(t *testing.T) {
	lotes := []LoteDisponible{lote(10, 0, 5)}

	asignaciones, err := AsignarLotes(lotes, 4, "FIFO")
	if err != nil {
		t.Fatalf("AsignarLotes() error = %v", err)
	}
	if len(asignaciones) != 1 || asignaciones[0].Unidades != 4 {
		t.Fatalf("asignaciones = %+v, want un solo lote con 4 unidades", asignaciones)
	}
}

func TestAsignarLotesPartenEntreDosLotes(t *testing.T) {
	viejo := lote(3, 0, 10)
	nuevo := lote(5, 0, 1)

	asignaciones, err := AsignarLotes([]LoteDisponible{nuevo, viejo}, 5, "FIFO")
	if err != nil {
		t.Fatalf("AsignarLotes() error = %v", err)
	}
	if len(asignaciones) != 2 {
		t.Fatalf("asignaciones = %+v, want dos lotes", asignaciones)
	}
	if asignaciones[0].LoteID != viejo.LoteID || asignaciones[0].Unidades != 3 {
		t.Fatalf("primer asignacion = %+v, want el lote mas viejo con 3 unidades", asignaciones[0])
	}
	if asignaciones[1].LoteID != nuevo.LoteID || asignaciones[1].Unidades != 2 {
		t.Fatalf("segunda asignacion = %+v, want el lote nuevo con 2 unidades", asignaciones[1])
	}
}

func TestAsignarLotesNoAlcanza(t *testing.T) {
	lotes := []LoteDisponible{lote(2, 0, 1)}

	_, err := AsignarLotes(lotes, 5, "FIFO")
	if err != ErrStockInsuficiente {
		t.Fatalf("AsignarLotes() error = %v, want ErrStockInsuficiente", err)
	}
}

func TestAsignarLotesUnidadesTodasReservadas(t *testing.T) {
	lotes := []LoteDisponible{lote(5, 5, 1)}

	if disponible := Disponible(lotes); disponible != 0 {
		t.Fatalf("Disponible() = %d, want 0", disponible)
	}

	_, err := AsignarLotes(lotes, 1, "FIFO")
	if err != ErrStockInsuficiente {
		t.Fatalf("AsignarLotes() error = %v, want ErrStockInsuficiente", err)
	}
}

func TestAsignarLotesLIFODevuelveOtroOrdenQueFIFO(t *testing.T) {
	viejo := lote(5, 0, 10)
	nuevo := lote(5, 0, 1)
	lotes := []LoteDisponible{viejo, nuevo}

	fifo, err := AsignarLotes(lotes, 3, "FIFO")
	if err != nil {
		t.Fatalf("AsignarLotes(FIFO) error = %v", err)
	}
	if fifo[0].LoteID != viejo.LoteID {
		t.Fatalf("FIFO eligió %v primero, want el lote viejo", fifo[0].LoteID)
	}

	lifo, err := AsignarLotes(lotes, 3, "LIFO")
	if err != nil {
		t.Fatalf("AsignarLotes(LIFO) error = %v", err)
	}
	if lifo[0].LoteID != nuevo.LoteID {
		t.Fatalf("LIFO eligió %v primero, want el lote nuevo", lifo[0].LoteID)
	}
}

func TestDisponibleIgnoraLotesSobrecomprometidos(t *testing.T) {
	// Un lote nunca debería tener más reservas que unidades, pero si pasara,
	// Disponible no debe devolver un número negativo que reste al total.
	lotes := []LoteDisponible{lote(2, 5, 1), lote(3, 0, 2)}

	if disponible := Disponible(lotes); disponible != 3 {
		t.Fatalf("Disponible() = %d, want 3", disponible)
	}
}
