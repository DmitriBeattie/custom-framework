package middlewares

import (
	"delivery-report/framework/abstract/apierror"
	"delivery-report/framework/interfaces/api"
	"delivery-report/framework/interfaces/request"
	"net/http"
)

func AuthError() apierror.APIError {
	return apierror.New().Component("AUTHMiddleware")
}

func AuthMiddleware(rIdent request.Identifier, checkPermissions []string, pr api.ErrorPresenter) api.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			perm, err := rIdent.Identify(r)

			if err != nil {
				pr.Error(
					w,
					r,
					AuthError().
						FromErr(err).
						Code("PERMISSIONNOTFOUND"),
					http.StatusUnauthorized,
					nil,
				)

				return
			}

			for i := range checkPermissions {
				var found bool

				for j := range perm {
					if perm[j] == checkPermissions[i] {
						found = true

						break
					}
				}

				if !found {
					pr.Error(
						w,
						r,
						AuthError().
							FromMsg("Недостаточно прав. Требуется {permission}").
							Args(apierror.ErrorArguments{"permission": checkPermissions[i]}).
							Code("PERMISSIONDENIED"),
						http.StatusUnauthorized,
						nil,
					)

					return
				}
			}

			next(w, r)
		}
	}
}
