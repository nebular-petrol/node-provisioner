package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	// Importamos nuestro nuevo paquete de discovery
	"github.com/nebular-petrol/go-node-provisioner/internal/discovery"
)

// TriggerDiscoveryHandler recibe las peticiones de Node.js para escanear un nodo
func TriggerDiscoveryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Método no permitido, se requiere POST"}`, http.StatusMethodNotAllowed)
			return
		}

		var req discovery.ScanRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.Warn("JSON inválido recibido desde Node.js", "error", err.Error())
			http.Error(w, `{"error": "JSON malformado"}`, http.StatusBadRequest)
			return
		}

		// Validaciones básicas
		if req.TargetIP == "" || req.Provider == "" {
			http.Error(w, `{"error": "target_ip y provider son campos obligatorios"}`, http.StatusBadRequest)
			return
		}
		if req.Port == 0 {
			req.Port = 623 // Asignamos el puerto por defecto de IPMI
		}
		// ¡LA MAGIA DE GO!
		// Lanzamos el proceso en segundo plano usando la palabra reservada 'go'
		go discovery.RunDiscoveryAsync(req)

		// Le respondemos inmediatamente a Node.js sin esperar a que termine el escaneo
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
		_, _ = w.Write([]byte(`{"status": "accepted", "message": "Proceso de descubrimiento lanzado en segundo plano"}`))
	}
}
