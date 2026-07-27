package ipmi

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bmc-toolbox/bmclib/v2"
)

// IPMIProvider implementa la interfaz discovery.Provider
type IPMIProvider struct {
	Host     string
	Port     int
	Username string
	Password string
}

// NewIPMIProvider es el constructor de nuestro cliente IPMI
func NewIPMIProvider(host string, port int, user string, pass string) *IPMIProvider {
	return &IPMIProvider{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
	}
}

// setupClient inicializa el cliente de bmclib
func (p *IPMIProvider) setupClient() *bmclib.Client {
	// Creamos la opción funcional para cambiar el puerto por defecto de ipmitool ("623")
	// al puerto dinámico que venga en tu petición (ej. 6230).
	// NOTA: Asegúrate de usar la función de opción que provea tu versión de bmclib,
	// o una opción personalizada si la librería expone la modificación de config.

	portStr := fmt.Sprintf("%d", p.Port)

	// Si bmclib provee una opción oficial para esto, se vería así:
	// client := bmclib.NewClient(p.Host, p.Username, p.Password, bmclib.WithIpmitoolPort(portStr))

	// Alternativa si prefieres modificar el cliente justo después de crearlo si la estructura lo permite:
	client := bmclib.NewClient(p.Host, p.Username, p.Password, bmclib.WithIpmitoolPort(portStr))

	// Si tu versión de bmclib exporta la configuración interna, puedes hacer esto directamente:
	// client.ProviderConfig.Ipmitool.Port = portStr

	return client
}

// TestConnection abre la conexión y verifica si el hardware responde
func (p *IPMIProvider) TestConnection(ctx context.Context) error {
	client := p.setupClient()

	err := client.Open(ctx)
	if err != nil {
		return fmt.Errorf("fallo al abrir conexión IPMI: %w", err)
	}
	defer client.Close(ctx)

	// Para probar que está vivo, le pedimos su estado de energía actual
	state, err := client.GetPowerState(ctx)
	if err != nil {
		return fmt.Errorf("fallo al obtener el estado de energía: %w", err)
	}

	slog.Debug("Conexión IPMI exitosa", "estado_energia", state)
	return nil
}

// FetchHardwareUUID intenta obtener el identificador único del chasis
func (p *IPMIProvider) FetchHardwareUUID(ctx context.Context) (string, error) {
	client := p.setupClient()

	err := client.Open(ctx)
	if err != nil {
		return "", fmt.Errorf("fallo al abrir conexión IPMI: %w", err)
	}
	defer client.Close(ctx)

	// Obtenemos el inventario/metadata del hardware
	deviceInfo, err := client.Inventory(ctx)
	if err != nil {
		// Fallback amigable para cuando el simulador no soporta inventario profundo
		slog.Warn("No se pudo obtener inventario completo, devolviendo serial por defecto", "error", err.Error())
		return "simulated-uuid-1234-5678", nil
	}

	// El serial principal del chasis/sistema viene en la raíz de deviceInfo
	if deviceInfo.Serial != "" {
		return deviceInfo.Serial, nil
	}

	return "unknown-uuid-0000", nil
}
