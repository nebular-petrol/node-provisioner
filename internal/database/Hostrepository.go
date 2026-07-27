package database

import (
	"database/sql"
	"time"
)

// SaveOrUpdateNode guarda un nodo descubierto en SQLite
func SaveOrUpdateNode(db *sql.DB, ip, uuid, status string) error {
	query := `
        INSERT INTO nodes (ip_address, uuid, status, last_seen)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(ip_address) 
        DO UPDATE SET 
            uuid = excluded.uuid,
            status = excluded.status,
            last_seen = excluded.last_seen;
    `

	_, err := db.Exec(query, ip, uuid, status, time.Now())
	return err
}
