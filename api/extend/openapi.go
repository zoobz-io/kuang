// Package extend holds handlers that extend or override rocco's built-in
// behavior, wrapping the engine rather than registering as endpoints.
package extend

import (
	"net/http"

	"github.com/zoobz-io/rocco"

	"github.com/zoobzio/kuang/internal/auth"
)

// OpenAPIInterceptor wraps the rocco router to serve identity-filtered
// OpenAPI specs at GET /openapi. Rocco's built-in /openapi handler serves
// an unfiltered spec; this interceptor calls GenerateOpenAPI(identity) to
// filter endpoints by the requesting agent's scopes.
//
// It must wrap the router rather than register as an endpoint: rocco reserves
// GET /openapi on its own mux, so a same-path endpoint would collide. Shadowing
// the route here lets us override the default spec with a scoped one.
func OpenAPIInterceptor(engine *rocco.Engine) http.Handler {
	router := engine.Router()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/openapi" {
			var id rocco.Identity
			if identity := auth.IdentityFromContext(r.Context()); identity != nil {
				id = identity
			}
			spec := engine.GenerateOpenAPI(id)
			data, err := spec.ToJSON()
			if err != nil {
				http.Error(w, "failed to generate spec", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
			return
		}
		router.ServeHTTP(w, r)
	})
}
