package database

import (
	"database/sql"
	"time"
)

type Node struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	IPAddress string    `json:"ip_address"`
	Status    string    `json:"status"`
	LastSeen  time.Time `json:"last_seen"`
}

// SaveOrUpdateNode guarda un nodo descubierto en SQLite
func SaveOrUpdateNode(db *sql.DB, ip, uuid, status string) error {
	query := `
        INSERT INTO nodes (ip_address, uuid, status, last_seen)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(uuid) 
        DO UPDATE SET 
            uuid = excluded.uuid,
            status = excluded.status,
            last_seen = excluded.last_seen;
    `

	_, err := db.Exec(query, ip, uuid, status, time.Now())
	return err
}

// GetAllNodes consulta y retorna todos los nodos ordenados por la última vez vistos
func GetAllNodes(db *sql.DB) ([]Node, error) {
	query := `SELECT id, uuid, ip_address, status, last_seen FROM nodes ORDER BY last_seen DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Inicializamos como slice vacío para que en JSON se serialice como [] en lugar de null si no hay nodos
	nodes := make([]Node, 0)

	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.UUID, &n.IPAddress, &n.Status, &n.LastSeen); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}
func GetNodeByUUID(uuid string) (*Node, error) {
	query := `SELECT id, uuid, ip_address, status, last_seen FROM nodes WHERE uuid = ?`

	var n Node
	err := DB.QueryRow(query, uuid).Scan(&n.ID, &n.UUID, &n.IPAddress, &n.Status, &n.LastSeen)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Retornamos nil si no se encuentra (no es un error fatal)
		}
		return nil, err
	}

	return &n, nil
}

// DeleteNode elimina un nodo de la base de datos
func DeleteNode(id int) error {
	query := `DELETE FROM nodes WHERE id = ?`
	_, err := DB.Exec(query, id)
	return err
}
