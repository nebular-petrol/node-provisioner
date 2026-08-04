package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nebular-petrol/go-node-provisioner/internal/database"
	"github.com/nebular-petrol/go-node-provisioner/internal/discovery/ipmi"
)

type PowerRequest struct {
	Action   string `json:"action"` // "on", "off", "reset"
	Username string `json:"username"`
	Password string `json:"password"`
	Port     int    `json:"port,omitempty"`
}

// Request para configurar el boot device
type BootRequest struct {
	Device   string `json:"device"` // "pxe", "disk", "bios"
	Username string `json:"username"`
	Password string `json:"password"`
	Port     int    `json:"port,omitempty"`
}

// HandleGetNodes ahora es una función regular, sin el receiver (s *Server)
func HandleGetNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	// Llamamos a la base de datos.
	// Asumo que tu paquete database ya maneja su propia conexión internamente.
	nodes, err := database.GetAllNodes(database.DB)
	if err != nil {
		slog.Error("Error al consultar la lista de nodos", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error interno al obtener los nodos"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nodes)
}

// HandleGetNode devuelve un solo nodo por su ID
func HandleGetNodeByUUID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/nodes/")
	if path == "" || path == r.URL.Path {
		http.Error(w, `{"error": "UUID del nodo requerida en la URL"}`, http.StatusBadRequest)
		return
	}
	uuid := path
	nodes, err := database.GetNodeByUUID(uuid)
	if err != nil {
		slog.Error("Error al consultar el nodo", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error interno al obtener el nodo"})
		return
	}
	if nodes == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nodo no encontrado"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nodes)

}

// HandleNodePower gestiona las acciones de energía (POST /nodes/{ip}/power)
func HandleNodePower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	// Extraer la IP de la URL: /nodes/192.168.1.50/power -> IP es 192.168.1.50
	path := strings.TrimPrefix(r.URL.Path, "/nodes/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "power" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ruta inválida"})
		return
	}
	targetIP := parts[0]

	var req PowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON inválido"})
		return
	}

	// Valores por defecto si no se envían
	if req.Port <= 0 {
		req.Port = 623
	}
	if req.Username == "" {
		req.Username = "ADMIN"
	}
	if req.Password == "" {
		req.Password = "ADMIN"
	}

	prov := ipmi.NewIPMIProvider(targetIP, req.Port, req.Username, req.Password)
	ctx := r.Context()

	var err error
	switch strings.ToLower(req.Action) {
	case "on":
		err = prov.PowerOn(ctx)
	case "off":
		err = prov.PowerOff(ctx)
	case "reset", "cycle":
		err = prov.PowerReset(ctx)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Acción de energía desconocida (use: on, off, reset)"})
		return
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Fallo al ejecutar la acción de energía: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"ip":     targetIP,
		"action": req.Action,
	})
}

// HandleNodeBoot gestiona la configuración del dispositivo de arranque (POST /nodes/{ip}/boot)
func HandleNodeBoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/nodes/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "boot" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ruta inválida"})
		return
	}
	targetIP := parts[0]

	var req BootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON inválido"})
		return
	}

	if req.Port <= 0 {
		req.Port = 623
	}
	if req.Username == "" {
		req.Username = "ADMIN"
	}
	if req.Password == "" {
		req.Password = "ADMIN"
	}

	prov := ipmi.NewIPMIProvider(targetIP, req.Port, req.Username, req.Password)
	ctx := r.Context()

	err := prov.SetBootDevice(ctx, req.Device)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Fallo al configurar el boot device: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"ip":     targetIP,
		"device": req.Device,
	})
}
