package discovery

import (
	"context"
	"log/slog"

	"time"

	"github.com/nebular-petrol/go-node-provisioner/internal/database"
	"github.com/nebular-petrol/go-node-provisioner/internal/discovery/ipmi"
)

type ScanRequest struct {
	TargetIP string `json:"target_ip"`
	Provider string `json:"provider"`
	Port     int    `json:"port,omitempty"` // Puerto opcional, por defecto 623 para IPMI
}

func RunDiscoveryAsync(req ScanRequest) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Pánico recuperado en la goroutine de descubrimiento", "ip", req.TargetIP, "error", r)
		}
	}()

	slog.Info("Iniciando escaneo", "ip", req.TargetIP, "port", req.Port, "proveedor", req.Provider)

	// 1. Configuramos un Timeout de 10 segundos.
	// Si el servidor no responde en ese tiempo, cortamos la conexión para no quedarnos colgados.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	var prov Provider

	// 2. Seleccionamos el proveedor (Factory simple)
	switch req.Provider {
	case "ipmi":
		// Usamos el puerto 6230 y credenciales de nuestro contenedor Docker
		prov = ipmi.NewIPMIProvider(req.TargetIP, req.Port, "ADMIN", "ADMIN")
	default:
		slog.Error("Proveedor no soportado", "ip", req.TargetIP, "proveedor", req.Provider)
		return
	}

	// 3. Probamos la conexión
	if err := prov.TestConnection(ctx); err != nil {
		slog.Error("Fallo en la prueba de conexión (Nodo FAILED)", "ip", req.TargetIP, "error", err.Error())
		// Aquí (en el futuro) actualizaríamos el SQLite a StateFailed
		return
	}

	// 4. Obtenemos el UUID
	uuid, err := prov.FetchHardwareUUID(ctx)
	if err != nil {
		slog.Error("No se pudo obtener el UUID", "ip", req.TargetIP, "error", err.Error())
		return
	}

	// 5. ¡Éxito!
	slog.Info("Escaneo finalizado con éxito (Hardware Real contactado)",
		"ip", req.TargetIP,
		"uuid_descubierto", uuid,
		"nuevo_estado", StateDiscovered)

	err = database.SaveOrUpdateNode(database.DB, req.TargetIP, uuid, "DISCOVERED")
	if err != nil {
		slog.Error("Error al guardar el nodo en la base de datos", "error", err)
		return
	}
	slog.Info("Nodo persistido exitosamente en la base de datos", "ip", req.TargetIP, "uuid", uuid)

}
