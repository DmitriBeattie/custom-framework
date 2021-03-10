package middlewares

import (
	"github.com/DmitriBeattie/custom-framework/interfaces/api"
	"net/http"

	"github.com/gorilla/context"
)

func ContextClear() api.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if r.Body != nil {
					r.Body.Close()
				}

				context.Clear(r)
			}()

			next(w, r)
		}
	}
}
