package identifiers

import (
	"github.com/gorilla/context"
	"net/http"
)

type Dummy struct {
}

func (d Dummy) Identify(r *http.Request) ([]string, error) {
	context.Set(r, "employeeID", "-74807")

	return nil, nil
}
