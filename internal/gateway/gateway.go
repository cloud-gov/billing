package gateway

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

// New returns a handler that forwards requests on to the specified destHost. destHost is expected to be a broker.
func New(destHost string, logger *slog.Logger) http.Handler {
	mux := chi.NewMux()
	l := logger.WithGroup("gateway")
	mux.Get("/", HandleRequest(destHost, l))
	return mux
}

// HandleRequest returns an [http.HandlerFunc] that forwards all requests to destHost. destHost is a host or host:port and is assumed to have already been validated.
func HandleRequest(destHost string, logger *slog.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Update the URL's Host to the target. The other URL-related fields directly on r, like Host and Proto, do not need to be set.
		u := url.URL{
			Scheme:  "http",
			Host:    destHost,
			Path:    r.URL.Path,
			RawPath: r.URL.RawPath,
		}
		r.URL = &u
		resp, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			logger.Error("making request to upstream", "err", err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			logger.Error("writing response body to downstream", "err", err)
			// Already wrote the response status code, so we can't write another.
			return
		}
	})
}
