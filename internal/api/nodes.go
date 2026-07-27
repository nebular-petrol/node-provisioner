package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// NodeRequest define la estructura JSON para crear o actualizar un nodo
type NodeRequest struct {
	MacAddress string `json:"mac_address"`
	IpAddress  string `json:"ip_address"`
	Vendor     string `json:"vendor"`
	Status     string `json:"status"`
}

// NodeResponse define la estructura para devolver los datos del nodo
type NodeResponse struct {
	ID         int    `json:"id"`
	MacAddress string `json:"mac_address"`
	IpAddress  string `json:"ip_address"`
	Vendor     string `json:"vendor"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// NodesHandler redirige las peticiones según el método HTTP (POST o GET)
func NodesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			RegisterNodeHandler(db)(w, r)
		case http.MethodGet:
			GetNodesHandler(db)(w, r)
		default:
			http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		}
	}
}

// RegisterNodeHandler maneja el POST para registrar un nodo
func RegisterNodeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req NodeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.Warn("Fallo al decodificar JSON del nodo", "error", err.Error())
			http.Error(w, `{"error": "JSON malformado"}`, http.StatusBadRequest)
			return
		}

		if req.MacAddress == "" || req.IpAddress == "" {
			http.Error(w, `{"error": "mac_address y ip_address son obligatorios"}`, http.StatusBadRequest)
			return
		}

		if req.Status == "" {
			req.Status = "DISCOVERED"
		}

		query := `INSERT INTO nodes (mac_address, ip_address, vendor, status) VALUES (?, ?, ?, ?)`
		_, err = db.Exec(query, req.MacAddress, req.IpAddress, req.Vendor, req.Status)
		if err != nil {
			slog.Error("Error al insertar nodo", "error", err.Error())
			http.Error(w, `{"error": "No se pudo guardar el nodo (posible MAC duplicada)"}`, http.StatusInternalServerError)
			return
		}

		slog.Info("Nodo registrado exitosamente", "mac", req.MacAddress, "ip", req.IpAddress)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status": "success", "message": "Nodo registrado correctamente"}`))
	}
}

// GetNodesHandler maneja el GET para listar todos los nodos registrados
func GetNodesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, mac_address, ip_address, vendor, status, created_at FROM nodes")
		if err != nil {
			slog.Error("Error al consultar nodos", "error", err.Error())
			http.Error(w, `{"error": "Error interno al consultar la base de datos"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var nodes []NodeResponse
		for rows.Next() {
			var n NodeResponse
			err := rows.Scan(&n.ID, &n.MacAddress, &n.IpAddress, &n.Vendor, &n.Status, &n.CreatedAt)
			if err != nil {
				slog.Error("Error al leer fila de nodo", "error", err.Error())
				continue
			}
			nodes = append(nodes, n)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(nodes)
	}
}

// NodeDetailHandler maneja operaciones sobre un nodo específico usando su ID (/nodes/1)
func NodeDetailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extraemos el ID de la URL (ej: /nodes/5 -> sacamos el "5")
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 || parts[2] == "" {
			http.Error(w, `{"error": "ID de nodo no especificado"}`, http.StatusBadRequest)
			return
		}
		nodeID := parts[2]

		switch r.Method {
		case http.MethodPut:
			UpdateNodeHandler(db, nodeID)(w, r)
		case http.MethodDelete:
			DeleteNodeHandler(db, nodeID)(w, r)
		default:
			http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		}
	}
}

// UpdateNodeHandler maneja el PUT para actualizar un nodo existente
func UpdateNodeHandler(db *sql.DB, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req NodeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, `{"error": "JSON malformado"}`, http.StatusBadRequest)
			return
		}

		query := `UPDATE nodes SET ip_address = ?, vendor = ?, status = ? WHERE id = ?`
		result, err := db.Exec(query, req.IpAddress, req.Vendor, req.Status, id)
		if err != nil {
			slog.Error("Error al actualizar nodo", "error", err.Error())
			http.Error(w, `{"error": "No se pudo actualizar el nodo"}`, http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, `{"error": "Nodo no encontrado"}`, http.StatusNotFound)
			return
		}

		slog.Info("Nodo actualizado exitosamente", "id", id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "success", "message": "Nodo actualizado correctamente"}`))
	}
}

// DeleteNodeHandler maneja el DELETE para borrar un nodo
func DeleteNodeHandler(db *sql.DB, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `DELETE FROM nodes WHERE id = ?`
		result, err := db.Exec(query, id)
		if err != nil {
			slog.Error("Error al eliminar nodo", "error", err.Error())
			http.Error(w, `{"error": "No se pudo eliminar el nodo"}`, http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, `{"error": "Nodo no encontrado"}`, http.StatusNotFound)
			return
		}

		slog.Info("Nodo eliminado exitosamente", "id", id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "success", "message": "Nodo eliminado correctamente"}`))
	}
}
