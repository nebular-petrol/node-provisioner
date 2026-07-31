package database

import (
	"database/sql"
	"log/slog"

	// Importamos el driver de SQLite que acabamos de descargar
	_ "github.com/glebarez/sqlite"
)

var DB *sql.DB

// InitDB inicializa la conexión con SQLite y crea las tablas necesarias si no existen
func InitDB(dataSourceName string) (*sql.DB, error) {
	// Si dataSourceName está vacío, guardaremos la base de datos en un archivo local llamado nod.db
	if dataSourceName == "" {
		dataSourceName = "nod.db"
	}

	slog.Info("Conectando a la base de datos SQLite", "archivo", dataSourceName)

	// Abrimos la conexión con la base de datos
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Verificamos que la conexión responda correctamente haciendo un "ping"
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Creamos nuestra tabla inicial para almacenar los nodos descubiertos
	// Guardaremos su ID, Dirección MAC, Dirección IP, el fabricante (Dell/HP) y su estado actual
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT UNIQUE, -- ¡ESTA ES LA PALABRA MÁGICA!
    ip_address TEXT, 
    status TEXT,
    last_seen DATETIME);
	`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	slog.Info("Base de datos inicializada y tabla 'nodes' lista correctamente")
	DB = db
	return DB, nil
}
