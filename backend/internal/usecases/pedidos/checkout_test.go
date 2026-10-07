package pedidos

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// ---- versiones falsas de los puertos (stubs en memoria) ----

type catalogoFalso struct {
	variantes map[uuid.UUID]out.VarianteVendible
}

func (c catalogoFalso) VariantePorID(_ context.Context, id uuid.UUID) (out.VarianteVendible, error) {
	v, ok := c.variantes[id]
	if !ok || !v.Activa {
		return out.VarianteVendible{}, apierr.ErrNoEncontrado
	}
	return v, nil
}

type stockFalso struct {
	disponible map[uuid.UUID]int
	reservas   int
}

func (s *stockFalso) Reservar(_ context.Context, _ *gorm.DB, _ uuid.UUID, items []out.ItemReservaStock) error {
	faltan := map[string]int{}
	for _, it := range items {
		if s.disponible[it.VarianteID] < it.Unidades {
			faltan[it.VarianteID.String()] = s.disponible[it.VarianteID]
		}
	}
	if len(faltan) > 0 {
		return apierr.ErrRegla{Mensaje: "No hay stock suficiente", Datos: faltan}
	}
	s.reservas++
	return nil
}

type pedidosFalso struct {
	guardado *domain.Pedido
	detalles []domain.DetallePedido
}

func (p *pedidosFalso) Crear(_ context.Context, _ *gorm.DB, ped *domain.Pedido, det []domain.DetallePedido) error {
	copia := *ped
	p.guardado, p.detalles = &copia, det
	return nil
}

type configFalsa struct {
	cfg     *domain.ConfiguracionTienda
	tarifas []domain.TarifaDistancia
}

func (c configFalsa) Actual(context.Context) (domain.ConfiguracionTienda, error) {
	if c.cfg == nil {
		return domain.ConfiguracionTienda{}, apierr.ErrNoEncontrado
	}
	return *c.cfg, nil
}
func (c configFalsa) TarifasOrdenadas(context.Context) ([]domain.TarifaDistancia, error) {
	return c.tarifas, nil
}

// txFalsa imita la transacción: anota si "confirmó" o "revirtió".
type txFalsa struct{ abiertas, confirmadas, revertidas int }

func (t *txFalsa) EnTransaccion(_ context.Context, fn func(*gorm.DB) error) error {
	t.abiertas++
	if err := fn(nil); err != nil {
		t.revertidas++
		return err
	}
	t.confirmadas++
	return nil
}

// ---- armado del escenario ----

type escenario struct {
	checkout *Checkout
	stock    *stockFalso
	pedidos  *pedidosFalso
	tx       *txFalsa
	perfume  out.VarianteVendible
	crema    out.VarianteVendible
	ahora    time.Time
}

func nuevoEscenario(cfg *domain.ConfiguracionTienda) *escenario {
	perfume := out.VarianteVendible{ID: uuid.New(), Codigo: "PERF-01", Nombre: "Perfume 100ml", PrecioMinoristaCentavos: 1500000, Activa: true}
	crema := out.VarianteVendible{ID: uuid.New(), Codigo: "CREM-01", Nombre: "Crema", PrecioMinoristaCentavos: 250050, Activa: true}
	e := &escenario{
		stock:   &stockFalso{disponible: map[uuid.UUID]int{perfume.ID: 5, crema.ID: 10}},
		pedidos: &pedidosFalso{},
		tx:      &txFalsa{},
		perfume: perfume,
		crema:   crema,
		ahora:   time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC),
	}
	catalogo := catalogoFalso{variantes: map[uuid.UUID]out.VarianteVendible{perfume.ID: perfume, crema.ID: crema}}
	tarifas := []domain.TarifaDistancia{
		{DesdeMetros: 0, HastaMetros: 1000, CostoCentavos: 50000},
		{DesdeMetros: 1000, HastaMetros: 5000, CostoCentavos: 150000},
	}
	e.checkout = NuevoCheckout(catalogo, e.stock, e.pedidos, configFalsa{cfg: cfg, tarifas: tarifas}, e.tx,
		15*time.Minute, func() time.Time { return e.ahora })
	return e
}

func configuracion(efectivo bool) *domain.ConfiguracionTienda {
	return &domain.ConfiguracionTienda{EfectivoHabilitado: efectivo, DistanciaMaximaMetros: 5000}
}

func (e *escenario) datos(modo domain.ModoEntrega, medio domain.MedioPago) DatosPedido {
	return DatosPedido{
		Items:       []ItemPedido{{VarianteID: e.perfume.ID, Unidades: 2}, {VarianteID: e.crema.ID, Unidades: 1}},
		Comprador:   Comprador{Nombre: "Ana", Apellido: "Pérez", Correo: "ana@example.com", Telefono: "3415550000", DNI: "30111222"},
		ModoEntrega: modo,
		MedioPago:   medio,
	}
}

func conEnvio(d DatosPedido, metros int) DatosPedido {
	dom := "San Martín 123"
	d.Domicilio, d.DistanciaMetros = &dom, &metros
	return d
}

// ---- tests ----

func TestCheckout_MercadoPagoRetiro_CreaPendienteConPreciosDelServidor(t *testing.T) {
	e := nuevoEscenario(nil)
	p, err := e.checkout.Crear(context.Background(), e.datos(domain.EntregaRetiro, domain.PagoMercadoPago))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if p.SubtotalCentavos != 2*1500000+250050 || p.TotalCentavos != p.SubtotalCentavos || p.EnvioCentavos != 0 {
		t.Errorf("importes mal calculados: %+v", p)
	}
	if p.EstadoPedido != domain.PedidoPendienteDePago || p.EstadoPago != domain.PagoPendiente {
		t.Errorf("estado inicial incorrecto: %s/%s", p.EstadoPedido, p.EstadoPago)
	}
	if p.VenceEn == nil || !p.VenceEn.Equal(e.ahora.Add(15*time.Minute)) {
		t.Errorf("la reserva debe vencer a los 15 minutos: %v", p.VenceEn)
	}
	if len(p.TokenAcceso) != 43 { // 32 bytes en base64 URL sin relleno
		t.Errorf("token de 32 bytes esperado, longitud %d", len(p.TokenAcceso))
	}
	if len(p.Detalles) != 2 || p.Detalles[0].Codigo != "PERF-01" || p.Detalles[0].PrecioUnitarioCentavos != 1500000 {
		t.Errorf("copia histórica incorrecta: %+v", p.Detalles)
	}
	if e.stock.reservas != 1 || e.tx.confirmadas != 1 {
		t.Errorf("debía reservar y confirmar una vez: reservas=%d confirmadas=%d", e.stock.reservas, e.tx.confirmadas)
	}
}

func TestCheckout_EfectivoRetiro_NaceConfirmadoSinVencimiento(t *testing.T) {
	e := nuevoEscenario(configuracion(true))
	p, err := e.checkout.Crear(context.Background(), e.datos(domain.EntregaRetiro, domain.PagoEfectivo))
	if err != nil {
		t.Fatal(err)
	}
	if p.EstadoPedido != domain.PedidoConfirmado || p.EstadoPago != domain.PagoPendiente || p.VenceEn != nil {
		t.Errorf("efectivo: esperaba CONFIRMADO/PENDIENTE sin vencimiento: %s/%s venceEn=%v", p.EstadoPedido, p.EstadoPago, p.VenceEn)
	}
}

func TestCheckout_ReglasDeEfectivo(t *testing.T) {
	casos := []struct {
		nombre string
		cfg    *domain.ConfiguracionTienda
		datos  func(*escenario) DatosPedido
	}{
		{"efectivo deshabilitado", configuracion(false), func(e *escenario) DatosPedido { return e.datos(domain.EntregaRetiro, domain.PagoEfectivo) }},
		{"sin configuración cargada", nil, func(e *escenario) DatosPedido { return e.datos(domain.EntregaRetiro, domain.PagoEfectivo) }},
		{"efectivo con envío", configuracion(true), func(e *escenario) DatosPedido {
			return conEnvio(e.datos(domain.EntregaEnvio, domain.PagoEfectivo), 500)
		}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := nuevoEscenario(c.cfg)
			_, err := e.checkout.Crear(context.Background(), c.datos(e))
			if !esRegla(err) {
				t.Fatalf("esperaba ErrRegla (409), vino %v", err)
			}
			if e.tx.abiertas != 0 {
				t.Error("una regla incumplida no debe ni abrir la transacción")
			}
		})
	}
}

func TestCheckout_Envio(t *testing.T) {
	e := nuevoEscenario(configuracion(false))
	p, err := e.checkout.Crear(context.Background(), conEnvio(e.datos(domain.EntregaEnvio, domain.PagoMercadoPago), 1000))
	if err != nil {
		t.Fatal(err)
	}
	if p.EnvioCentavos != 150000 || p.TotalCentavos != p.SubtotalCentavos+150000 { // 1000 m exactos → segundo rango
		t.Errorf("envío mal cotizado: %+v", p)
	}
	if p.Domicilio == nil || p.DistanciaMetros == nil || *p.DistanciaMetros != 1000 {
		t.Errorf("domicilio y distancia deben guardarse en el pedido: %+v", p)
	}
}

func TestCheckout_EnvioFueraDeCobertura_OfreceRetiro(t *testing.T) {
	e := nuevoEscenario(configuracion(false))
	_, err := e.checkout.Crear(context.Background(), conEnvio(e.datos(domain.EntregaEnvio, domain.PagoMercadoPago), 5001))
	var r apierr.ErrRegla
	if !errors.As(err, &r) || r.Mensaje != "No hacemos envíos a esa distancia. Podés elegir retiro en el local." {
		t.Fatalf("esperaba el 409 que ofrece retiro, vino %v", err)
	}
	if e.pedidos.guardado != nil || e.stock.reservas != 0 {
		t.Error("no debe crearse nada si el envío no es cotizable")
	}
}

func TestCheckout_EnvioSinConfiguracion_NoInventaCosto(t *testing.T) {
	e := nuevoEscenario(nil)
	_, err := e.checkout.Crear(context.Background(), conEnvio(e.datos(domain.EntregaEnvio, domain.PagoMercadoPago), 500))
	if !esRegla(err) {
		t.Fatalf("esperaba ErrRegla, vino %v", err)
	}
}

func TestCheckout_StockInsuficiente_RevierteYEnumera(t *testing.T) {
	e := nuevoEscenario(nil)
	d := e.datos(domain.EntregaRetiro, domain.PagoMercadoPago)
	d.Items[0].Unidades = 9 // solo hay 5 perfumes

	_, err := e.checkout.Crear(context.Background(), d)

	var r apierr.ErrRegla
	if !errors.As(err, &r) {
		t.Fatalf("esperaba ErrRegla (409), no un 500: %v", err)
	}
	faltan, ok := r.Datos.(map[string]int)
	if !ok || faltan[e.perfume.ID.String()] != 5 {
		t.Errorf("el 409 debe enumerar qué falta (variante → disponible): %#v", r.Datos)
	}
	if e.tx.revertidas != 1 || e.tx.confirmadas != 0 {
		t.Errorf("la transacción debía revertirse: revertidas=%d confirmadas=%d", e.tx.revertidas, e.tx.confirmadas)
	}
}

func TestCheckout_VarianteInexistente_404SinAbrirTransaccion(t *testing.T) {
	e := nuevoEscenario(nil)
	d := e.datos(domain.EntregaRetiro, domain.PagoMercadoPago)
	d.Items = append(d.Items, ItemPedido{VarianteID: uuid.New(), Unidades: 1})

	_, err := e.checkout.Crear(context.Background(), d)
	if !errors.Is(err, apierr.ErrNoEncontrado) {
		t.Fatalf("esperaba ErrNoEncontrado, vino %v", err)
	}
	if e.tx.abiertas != 0 {
		t.Error("no debe abrirse la transacción si falta una variante")
	}
}

func TestCheckout_VinculaElUsuarioDeLaSesion(t *testing.T) {
	e := nuevoEscenario(nil)
	d := e.datos(domain.EntregaRetiro, domain.PagoMercadoPago)

	invitado, _ := e.checkout.Crear(context.Background(), d)
	if invitado.UsuarioID != nil {
		t.Error("el invitado no debe tener usuario")
	}

	id := uuid.New()
	d.UsuarioID = &id
	registrado, _ := e.checkout.Crear(context.Background(), d)
	if registrado.UsuarioID == nil || *registrado.UsuarioID != id {
		t.Error("el pedido debe quedar vinculado al usuario de la sesión")
	}
	if invitado.TokenAcceso == registrado.TokenAcceso {
		t.Error("cada pedido necesita su propio token")
	}
}
