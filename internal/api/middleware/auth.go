package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// ctxKey is the key for accessing Claims data stored in a Context. See Context.Value() docs for explanation.
var ctxKey struct{}

type Claims struct {
	Email  string   `json:"email"`
	Scopes []string `json:"scope"`
}

func NewHasScope(logger *slog.Logger, verifier *oidc.IDTokenVerifier, scope string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ah := r.Header.Get("Authorization")
			if !strings.HasPrefix(strings.ToLower(ah), "bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			raw := strings.TrimSpace(ah[7:]) // len("bearer ")

			idTok, err := verifier.Verify(r.Context(), raw)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			var c Claims
			err = idTok.Claims(&c)
			if err != nil {
				http.Error(w, "cannot parse claims", http.StatusUnauthorized)
				return
			}

			if scope != "" && !slices.Contains(c.Scopes, scope) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			rc := r.WithContext(context.WithValue(r.Context(), ctxKey, c))

			h.ServeHTTP(w, rc)
		})
	}
}

func ClaimsFrom(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(ctxKey).(Claims)
	return c, ok
}
