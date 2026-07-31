package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nebular-petrol/go-node-provisioner/internal/database"
)

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
