package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nebular-petrol/go-node-provisioner/internal/api"
	"github.com/nebular-petrol/go-node-provisioner/internal/database" // Importamos nuestro paquete de DB
)

func main() {
	// 1. Configuración de Logs (LOG_LEVEL)
	logLevelEnv := strings.ToLower(os.Getenv("LOG_LEVEL"))
	var level slog.Level

	switch logLevelEnv {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.Info("Iniciando servicio NOD Provisioner", "version", "0.6.0-db")

	// 2. INICIALIZAR LA BASE DE DATOS
	// Podemos leer una variable de entorno DB_PATH o usar "nod.db" por defecto
	dbPath := os.Getenv("DB_PATH")
	db, err := database.InitDB(dbPath)
	if err != nil {
		slog.Error("No se pudo iniciar la base de datos", "error", err.Error())
		os.Exit(1)
	}
	// Nos aseguramos de cerrar la conexión a la base de datos cuando el programa termine
	defer func(db *sql.DB) {
		errClose := db.Close()
		if errClose != nil {
			slog.Error("Error al cerrar la base de datos", "error", errClose.Error())
		}
	}(db)

	// 3. LEER Y CONFIGURAR EL PUERTO (PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	// 4. CONFIGURAR EL ENRUTADOR (MUX)
	mux := http.NewServeMux()
	api.SetupRoutes(mux, db)

	// 5. CONFIGURACIÓN DEL SERVIDOR SEGURO
	protocol := strings.ToLower(os.Getenv("PROTOCOL"))
	if protocol == "" {
		protocol = "http"
	}

	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Debug("Configuración cargada", "puerto", port, "protocolo", protocol, "log_level", logLevelEnv)

	if protocol == "https" {
		slog.Info("Servidor HTTPS escuchando", "puerto", port)
		err = server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem")
	} else {
		slog.Info("Servidor HTTP (Inseguro) escuchando", "puerto", port)
		err = server.ListenAndServe()
	}

	if err != nil {
		slog.Error("El servidor falló de forma crítica", "error", err.Error())
		os.Exit(1)
	}
}
