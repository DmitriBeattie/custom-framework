package request

import "net/http"

type Identifier interface {
	Identify(r *http.Request) (permissions []string, err error)
}
