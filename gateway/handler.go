package gateway

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewHandler(cfg Config) (http.Handler, error) {
	target, err := url.Parse(cfg.OrderServiceURL)
	if err != nil {
		return nil, err
	}

	if target.Scheme == "" || target.Host == "" {
		return nil, http.ErrNotSupported
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.ErrorHandler = func(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	) {
		log.Printf(
			"gateway upstream failure request_id=%s error=%v",
			r.Header.Get("X-Request-Id"),
			err,
		)

		http.Error(
			w,
			`{"error":"Order service unavailable"}`,
			http.StatusBadGateway,
		)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health/live", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"alive"}`))
	})

	// Forward all other paths to the Order Management API.
	// The original path, query parameters, body, and authorization
	// header are preserved by the reverse proxy.
	mux.Handle("/", proxy)

	return RequestID(Logging(Recovery(mux))), nil
}