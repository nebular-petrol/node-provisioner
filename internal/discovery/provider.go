package discovery

import "context"

// Credentials almacena la autenticación para acceder al hardware
type Credentials struct {
	Username string
	Password string
}

// Provider define el contrato estricto que cualquier proveedor (IPMI, Proxmox, Redfish) debe cumplir.
// Usamos context.Context para poder cancelar peticiones si la red se cuelga (Timeout).
type Provider interface {
	TestConnection(ctx context.Context) error
	FetchHardwareUUID(ctx context.Context) (string, error)
}
