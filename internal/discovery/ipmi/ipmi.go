package ipmi

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bmc-toolbox/bmclib/v2"
	"github.com/google/uuid"
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

func GenerarUUIDAleatorio() (string, error) {
	nuevoUUID := uuid.New()
	return nuevoUUID.String(), nil
}

// setupClient inicializa el cliente de bmclib
func (p *IPMIProvider) setupClient() *bmclib.Client {
	portStr := fmt.Sprintf("%d", p.Port)
	client := bmclib.NewClient(p.Host, p.Username, p.Password, bmclib.WithIpmitoolPort(portStr))
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

	state, err := client.GetPowerState(ctx)
	if err != nil {
		return fmt.Errorf("fallo al obtener el estado de energía: %w", err)
	}

	slog.Debug("Conexión IPMI exitosa", "estado_energia", state)
	return nil
}

// FetchHardwareUUID maneja el flujo de respaldo en cascada
func (p *IPMIProvider) FetchHardwareUUID(ctx context.Context) (string, error) {
	client := p.setupClient()

	err := client.Open(ctx)
	if err != nil {
		return "", fmt.Errorf("fallo al abrir conexión IPMI: %w", err)
	}
	defer client.Close(ctx)

	// PASO 1: Intentar obtener el inventario profundo
	deviceInfo, err := client.Inventory(ctx)

	// Verificamos si el inventario fue exitoso Y el serial no está vacío
	if err == nil && deviceInfo != nil && deviceInfo.Serial != "" {
		slog.Info("UUID obtenido del inventario profundo", "serial", deviceInfo.Serial)
		return deviceInfo.Serial, nil
	}

	// Si hubo error en el inventario o el serial vino vacío, lo registramos
	if err != nil {
		slog.Warn("No se pudo obtener inventario profundo, pasando a fallback aleatorio", "error", err.Error())
	} else {
		slog.Warn("Inventario obtenido pero el campo Serial está vacío, pasando a fallback aleatorio")
	}

	// PASO 2: Fallback - Generar un UUID aleatorio
	nuevoUUID, errGen := GenerarUUIDAleatorio()
	if errGen == nil && nuevoUUID != "" {
		slog.Info("UUID aleatorio generado exitosamente como fallback", "uuid", nuevoUUID)
		return nuevoUUID, nil
	}

	if errGen != nil {
		slog.Warn("Fallo al generar el UUID aleatorio", "error", errGen.Error())
	}

	// PASO 3: Último recurso si todo lo anterior falló
	slog.Error("No se pudo obtener ni generar un UUID válido, usando unknown")
	return "unknown-uuid-0000", nil
}

// PowerOn enciende el nodo
func (p *IPMIProvider) PowerOn(ctx context.Context) error {
	client := bmclib.NewClient(p.Host, p.Username, p.Password)
	// Nota: Dependiendo de tu versión de bmclib, se abre la sesión primero o se usa Open()
	// Aquí usamos el flujo estándar de comandos de energía de bmclib

	// Ejemplo conceptual usando el cliente de bmclib para energía:
	err := client.Open(ctx)
	if err != nil {
		slog.Error("Error en la conexión del nodo seleccionado")
		return err
	}
	defer client.Close(ctx)

	nodestate, err := client.SetPowerState(ctx, "on")
	if err != nil {
		slog.Error("Error al ejecutar encendido del nodo seleccionado")
		return err
	}
	if !nodestate {
		return fmt.Errorf("no se pudo encender el nodo seleccionado")
	}

	return nil
}

// PowerOff apaga el nodo
func (p *IPMIProvider) PowerOff(ctx context.Context) error {
	client := bmclib.NewClient(p.Host, p.Username, p.Password)
	// Nota: Dependiendo de tu versión de bmclib, se abre la sesión primero o se usa Open()
	// Aquí usamos el flujo estándar de comandos de energía de bmclib

	// Ejemplo conceptual usando el cliente de bmclib para energía:
	err := client.Open(ctx)
	if err != nil {
		slog.Error("Error en la conexión del nodo seleccionado")
		return err
	}
	defer client.Close(ctx)

	nodestate, err := client.SetPowerState(ctx, "off")
	if err != nil {
		slog.Error("Error al ejecutar apagado del nodo seleccionado")
		return err
	}
	if !nodestate {
		return fmt.Errorf("no se pudo apagar el nodo seleccionado")
	}

	return nil
}

// PowerReset reinicia el nodo (Ciclo de energía)
func (p *IPMIProvider) PowerReset(ctx context.Context) error {
	client := bmclib.NewClient(p.Host, p.Username, p.Password)
	// Nota: Dependiendo de tu versión de bmclib, se abre la sesión primero o se usa Open()
	// Aquí usamos el flujo estándar de comandos de energía de bmclib

	// Ejemplo conceptual usando el cliente de bmclib para energía:
	err := client.Open(ctx)
	if err != nil {
		slog.Error("Error en la conexión del nodo seleccionado")
		return err
	}
	defer client.Close(ctx)

	nodestate, err := client.SetPowerState(ctx, "cycle")
	if err != nil {
		slog.Error("Error al ejecutar reinicio del nodo seleccionado")
		return err
	}
	if !nodestate {
		return fmt.Errorf("no se pudo reiniciar el nodo seleccionado")
	}

	return nil
}

// SetBootDevice configura el siguiente dispositivo de arranque (ej. pxe, disk)
func (p *IPMIProvider) SetBootDevice(ctx context.Context, device string) error {
	client := bmclib.NewClient(p.Host, p.Username, p.Password)
	// Nota: Dependiendo de tu versión de bmclib, se abre la sesión primero o se usa Open()
	// Aquí usamos el flujo estándar de comandos de energía de bmclib

	// Ejemplo conceptual usando el cliente de bmclib para energía:
	err := client.Open(ctx)
	if err != nil {
		slog.Error("Error en la conexión del nodo seleccionado")
		return err
	}
	defer client.Close(ctx)

	nodestate, err := client.SetBootDevice(ctx, device, false, true)
	if err != nil {
		slog.Error("Error al ejecutar configuración del dispositivo de arranque")
		return err
	}
	if !nodestate {
		return fmt.Errorf("no se pudo configurar el dispositivo de arranque")
	}

	return nil
}
