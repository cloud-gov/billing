// Package meter reads usage information from Cloud Foundry and AWS.
package meter

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/go-chi/chi/v5"
)

func New(logger *slog.Logger, cf *client.Client) http.Handler {
	mux := chi.NewMux()
	l := logger.WithGroup("meter")
	mux.Get("/", Handle(l, cf))
	return mux
}

// First draft. Later, this will be a scheduled background job.
func Handle(logger *slog.Logger, cf *client.Client) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readings, err := ReadUsage(r.Context(), cf.ServiceInstances, cf.Spaces)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(readings)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

// next step: POST a ReadMeter job or something. Starts a job which finishes when services are read and result is stored in the database.
