package api

import (
	"net/http"
)

type EndpointName string

type UseCaseName string

type UseCase interface {
	Handle(endpointName EndpointName) http.HandlerFunc
}

type UseCaseFactory interface {
	Build(useCaseName UseCaseName, r *http.Request) (UseCase, error)
}

func UseCaseHandle(useCaseName UseCaseName, endpointName EndpointName, useCaseFactory UseCaseFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		useCase, err := useCaseFactory.Build(useCaseName, r)
		if err != nil {
			panic(err)
		}

		useCase.Handle(endpointName)(w, r)
	}
}
