package pedidos

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// Comprador son los datos de contacto que el comprador escribe en el checkout.
type Comprador struct {
	Nombre, Apellido, Correo, Telefono, DNI string
}

// DatosPedido es TODO lo que el comprador decide al crear un pedido: qué
// quiere, quién es, cómo lo recibe y cómo paga. Fijate lo que NO tiene: ni
// precios, ni totales, ni estados, ni id de usuario. Esos los decide el
// servidor; si un campo así existiera acá, cualquiera podría crear desde
// Postman un pedido pagado de $0. UsuarioID lo completa el handler desde la
// sesión (nil si compra un invitado), nunca desde el body.
type DatosPedido struct {
	UsuarioID       *uuid.UUID
	Items           []ItemPedido
	Comprador       Comprador
	ModoEntrega     domain.ModoEntrega
	Domicilio       *string
	DistanciaMetros *int
	MedioPago       domain.MedioPago
}

// Checkout crea pedidos. Recibe todo lo que necesita por parámetro (puertos y
// reloj) para poder probarse con versiones falsas, sin base ni espera.
type Checkout struct {
	catalogo   out.LectorCatalogo
	stock      out.ReservadorStock
	pedidos    out.CreadorDePedidos
	config     out.LectorConfiguracion
	tx         out.Transaccionador
	ttlReserva time.Duration
	ahora      func() time.Time
}

func NuevoCheckout(
	catalogo out.LectorCatalogo,
	stock out.ReservadorStock,
	pedidos out.CreadorDePedidos,
	config out.LectorConfiguracion,
	tx out.Transaccionador,
	ttlReserva time.Duration,
	ahora func() time.Time,
) *Checkout {
	return &Checkout{catalogo: catalogo, stock: stock, pedidos: pedidos, config: config, tx: tx, ttlReserva: ttlReserva, ahora: ahora}
}

// Crear ejecuta el checkout en el orden de la clase 6: primero se leen los
// datos y se validan las reglas (devuelven errores esperados: 404/409), y
// recién entonces se hace el trabajo, dentro de una sola transacción.
func (c *Checkout) Crear(ctx context.Context, d DatosPedido) (domain.Pedido, error) {
	// 1. Cada variante se busca en el catálogo: ahí está el precio real.
	// Inexistente o inactiva → ErrNoEncontrado (404).
	precios := make(map[uuid.UUID]out.VarianteVendible, len(d.Items))
	for _, it := range d.Items {
		if _, ya := precios[it.VarianteID]; ya {
			continue
		}
		v, err := c.catalogo.VariantePorID(ctx, it.VarianteID)
		if err != nil {
			return domain.Pedido{}, err
		}
		precios[it.VarianteID] = v
	}

	// 2. Líneas valorizadas con los precios del servidor, nunca los del body.
	lineas, err := CalcularLineas(d.Items, precios)
	if err != nil {
		return domain.Pedido{}, err
	}

	// 3 y 4. Entrega y medio de pago contra la configuración de la tienda.
	envio, err := c.cotizarEntrega(ctx, d)
	if err != nil {
		return domain.Pedido{}, err
	}

	// 5. Totales.
	subtotal, total := CalcularTotal(lineas, envio)

	// 6. El pedido, con su secreto de acceso y su copia histórica.
	token, err := nuevoToken()
	if err != nil {
		return domain.Pedido{}, err
	}
	ahora := c.ahora()
	pedido := domain.Pedido{
		ID:                uuid.New(),
		UsuarioID:         d.UsuarioID,
		CompradorNombre:   d.Comprador.Nombre,
		CompradorApellido: d.Comprador.Apellido,
		CompradorCorreo:   d.Comprador.Correo,
		CompradorTelefono: d.Comprador.Telefono,
		CompradorDNI:      d.Comprador.DNI,
		ModoEntrega:       d.ModoEntrega,
		EnvioCentavos:     envio,
		SubtotalCentavos:  subtotal,
		TotalCentavos:     total,
		MedioPago:         d.MedioPago,
		EstadoPago:        domain.PagoPendiente,
		TokenAcceso:       token,
		CreadoEn:          ahora,
	}
	if d.ModoEntrega == domain.EntregaEnvio {
		pedido.Domicilio = d.Domicilio
		pedido.DistanciaMetros = d.DistanciaMetros
	}

	// 8. Estado inicial según cómo se paga.
	switch d.MedioPago {
	case domain.PagoMercadoPago:
		pedido.EstadoPedido = domain.PedidoPendienteDePago
		vence := ahora.Add(c.ttlReserva) // reserva de 15 minutos
		pedido.VenceEn = &vence
	case domain.PagoEfectivo:
		pedido.EstadoPedido = domain.PedidoConfirmado // sin vencimiento automático
	}

	detalles := make([]domain.DetallePedido, len(lineas))
	items := make([]out.ItemReservaStock, len(lineas))
	for i, l := range lineas {
		detalles[i] = domain.DetallePedido{
			VarianteID:             l.VarianteID,
			Codigo:                 l.Codigo,
			Nombre:                 l.Nombre,
			Unidades:               l.Unidades,
			PrecioUnitarioCentavos: l.PrecioUnitarioCentavos,
			SubtotalCentavos:       l.SubtotalCentavos,
		}
		items[i] = out.ItemReservaStock{VarianteID: l.VarianteID, Unidades: l.Unidades}
	}

	// 7 y 9. Crear el pedido y reservar el stock, o ninguna de las dos
	// cosas. Si Reservar falla (409 por stock), la transacción revierte el
	// pedido: no puede quedar uno creado sin su stock reservado, ni al revés.
	err = c.tx.EnTransaccion(ctx, func(tx *gorm.DB) error {
		if err := c.pedidos.Crear(ctx, tx, &pedido, detalles); err != nil {
			return err
		}
		return c.stock.Reservar(ctx, tx, pedido.ID, items)
	})
	if err != nil {
		return domain.Pedido{}, err
	}
	pedido.Detalles = detalles
	return pedido, nil
}

// cotizarEntrega aplica las reglas de entrega y medio de pago y devuelve el
// costo de envío en centavos (cero para retiro).
func (c *Checkout) cotizarEntrega(ctx context.Context, d DatosPedido) (int64, error) {
	if d.ModoEntrega == domain.EntregaRetiro && d.MedioPago != domain.PagoEfectivo {
		return 0, nil // retiro con Mercado Pago: no necesita configuración
	}

	cfg, err := c.config.Actual(ctx)
	if err != nil && !errors.Is(err, apierr.ErrNoEncontrado) {
		return 0, err
	}
	sinConfiguracion := errors.Is(err, apierr.ErrNoEncontrado)

	if d.MedioPago == domain.PagoEfectivo {
		if d.ModoEntrega != domain.EntregaRetiro {
			return 0, apierr.ErrRegla{Mensaje: "El pago en efectivo solo está disponible con retiro en el local."}
		}
		if sinConfiguracion || !cfg.EfectivoHabilitado {
			return 0, apierr.ErrRegla{Mensaje: "El pago en efectivo no está habilitado por el momento. Podés pagar con Mercado Pago."}
		}
		return 0, nil
	}

	// Envío con Mercado Pago.
	if sinConfiguracion {
		return 0, apierr.ErrRegla{Mensaje: "El envío no está disponible por el momento. Podés elegir retiro en el local."}
	}
	tarifas, err := c.config.TarifasOrdenadas(ctx)
	if err != nil {
		return 0, err
	}
	reglas := make([]Tarifa, len(tarifas))
	for i, t := range tarifas {
		reglas[i] = Tarifa{DesdeMetros: t.DesdeMetros, HastaMetros: t.HastaMetros, CostoCentavos: t.CostoCentavos}
	}
	distancia := 0
	if d.DistanciaMetros != nil {
		distancia = *d.DistanciaMetros
	}
	return CotizarEnvio(distancia, cfg.DistanciaMaximaMetros, reglas)
}

// nuevoToken genera el secreto del enlace privado del invitado: 32 bytes del
// generador criptográfico del sistema operativo, en base64 URL. No es
// adivinable y nunca se escribe en un log.
func nuevoToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
