package api

import (
	"log/slog"
	"net/http"
)

// AuthMiddleware intercepta la petición HTTP y valida un Token antes de continuar
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// En producción, este token debería venir de una variable de entorno
		// ej: expectedToken := os.Getenv("NOD_API_KEY")
		expectedToken := "nod-secreto-123"

		// Extraemos el token del header de la petición
		clientToken := r.Header.Get("X-API-Key")

		if clientToken != expectedToken {
			slog.Warn("Intento de acceso no autorizado bloqueado", "ip", r.RemoteAddr)
			http.Error(w, `{"error": "No autorizado: API Key inválida o ausente"}`, http.StatusUnauthorized)
			return
		}

		// Si el token es válido, pasamos el control al handler original
		next.ServeHTTP(w, r)
	}
}
