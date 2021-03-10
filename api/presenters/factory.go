package presenters

import (
	"fmt"
	"delivery-report/framework/interfaces/api"
	utilshttp "delivery-report/framework/utils/http"
	"net/http"
)

type RequestPresenter struct {
	parseRequestFunc func(r *http.Request) (interface{}, error)
	responseFunc func(w http.ResponseWriter, r *http.Request, data interface{})
}

type defaultPresenterFactory struct {
	def map[api.UseCaseName]map[api.EndpointName]*RequestPresenter
}

func DefaultPresenterFactory(def map[api.UseCaseName]map[api.EndpointName]*RequestPresenter) *defaultPresenterFactory {
	return &defaultPresenterFactory{def: def}
}

func NewDefaultPresenterFactory() *defaultPresenterFactory {
	return &defaultPresenterFactory{def: make(map[api.UseCaseName]map[api.EndpointName]*RequestPresenter)}
}

func (d *defaultPresenterFactory) AddDef(
	uCaseName api.UseCaseName,
	ePointName api.EndpointName,
	parseRequestFunc func(r *http.Request) (interface{}, error),
	responseFunc func(w http.ResponseWriter, r *http.Request, data interface{}),
	) {
		rp := &RequestPresenter{
			parseRequestFunc: parseRequestFunc,
			responseFunc:     responseFunc,
		}

		if _, ok := d.def[uCaseName]; !ok {
			d.def[uCaseName] = map[api.EndpointName]*RequestPresenter{
				ePointName: rp,
			}
		} else {
			d.def[uCaseName][ePointName] = rp
		}
}

var (
	NotFoundEndpoint = "There's no action for build endpoint %s of use case %s"
	NotFoundUseCase  = "There's no action for use case %s"
)

func (d *defaultPresenterFactory) ParseRequest(r *http.Request) (interface{}, error) {
	def, err := d.findDefByRequest(r)
	if err != nil {
		return nil, err
	}

	return def.parseRequestFunc(r)
}

func (d *defaultPresenterFactory) Response(w http.ResponseWriter, r *http.Request, data interface{}) {
	def, err := d.findDefByRequest(r)
	if err != nil {
		panic(err)
	}

	def.responseFunc(w, r, data)
}


func (d *defaultPresenterFactory) findDefByRequest(r *http.Request) (*RequestPresenter, error) {
	useCaseName, err := utilshttp.GetUseCaseNameFromRequestContext(r)
	if err != nil {
		return nil, err
	}

	endPointData, ok := d.def[useCaseName]
	if !ok {
		return nil, fmt.Errorf(NotFoundUseCase, useCaseName)
	}

	endPointName, err := utilshttp.GetEndpointNameFromRequestContext(r)
	if err != nil {
		return nil, err
	}

	pr, ok := endPointData[endPointName]
	if !ok {
		return nil, fmt.Errorf(NotFoundEndpoint, endPointName, useCaseName)
	}

	return pr, nil
}