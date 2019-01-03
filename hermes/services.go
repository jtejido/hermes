package hermes

import (
	"net/http"
)

type service func(http.Handler) http.Handler

func loader(h http.Handler, svcs ...service) http.Handler {
	for _, svc := range svcs {
		h = svc(h)
	}
	return h
}
