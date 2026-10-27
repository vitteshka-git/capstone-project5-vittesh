package gateway

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")

		if id == "" {
			var b [16]byte
			if _, err := rand.Read(b[:]); err == nil {
				id = hex.EncodeToString(b[:])
			} else {
				id = "request-id-unavailable"
			}
		}

		r.Header.Set("X-Request-Id", id)
		w.Header().Set("X-Request-Id", id)

		next.ServeHTTP(w, r)
	})
}