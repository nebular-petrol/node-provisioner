package api

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	healthCheckCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nod_healthcheck_peticiones_total",
		Help: "Número total de veces que se ha consultado el healthcheck",
	})
)

func SetupRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/health", healthCheckHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Rutas de Inventario de Nodos
	mux.HandleFunc("/nodes", NodesHandler(db))
	mux.HandleFunc("/nodes/", NodeDetailHandler(db)) // Captura /nodes/{id}

	// NUEVA RUTA: Disparador de descubrimiento para Node.js
	mux.HandleFunc("/discover", AuthMiddleware(TriggerDiscoveryHandler()))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	healthCheckCounter.Inc()

	slog.Info("Petición recibida", "ruta", r.URL.Path, "metodo", r.Method, "ip_cliente", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status": "ok", "message": "El cerebro está vivo, modularizado y monitorizado"}`))
	if err != nil {
		slog.Error("Fallo al escribir la respuesta HTTP", "error", err.Error())
	}
}
