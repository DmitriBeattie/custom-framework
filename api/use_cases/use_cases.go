package use_cases

import (
	"fmt"
	"delivery-report/framework/abstract/apierror"
	"delivery-report/framework/interfaces/api"
	"delivery-report/framework/interfaces/app"
	"delivery-report/framework/interfaces/translator"
	utilshttp "delivery-report/framework/utils/http"
	"github.com/gorilla/context"
	"net/http"
)

var useCases map[api.UseCaseName]map[api.EndpointName]HandleFunc

type HandleFunc func(useCase api.UseCase,  parsedRequest interface{}, r *http.Request) (responseData interface{}, statusCode int, err error)

func Register(uCase api.UseCaseName, endPointName api.EndpointName, handler HandleFunc) {
	if useCases == nil {
		useCases = map[api.UseCaseName]map[api.EndpointName]HandleFunc{
			uCase: map[api.EndpointName]HandleFunc{
				endPointName: handler,
			},
		}

		return
	}

	data, ok := useCases[uCase]
	if !ok {
		useCases[uCase] = map[api.EndpointName]HandleFunc{
			endPointName: handler,
		}

		return
	}

	data[endPointName] = handler
}

type CommonUseCaseData struct {
	log       app.Logger
	lang      translator.Language
	trRule    translator.Translator
	presenter api.HTTPPresenter
}

func NewCommonUseCaseData(_log app.Logger, _lang translator.Language, _trRule translator.Translator, _presenter api.HTTPPresenter) *CommonUseCaseData {
	return &CommonUseCaseData{
		log:       _log,
		lang:      _lang,
		trRule:    _trRule,
		presenter: _presenter,
	}
}

func registerUseCase(useCaseName api.UseCaseName, r *http.Request) {
	context.Set(r, api.UseCaseName("use-case"), useCaseName)
}

func registerEndpoint(endpointName api.EndpointName, r *http.Request) {
	context.Set(r, api.EndpointName("endpoint"), endpointName)
}

func requestError(cmn *CommonUseCaseData, err error, w http.ResponseWriter, r *http.Request) {
	var useCaseName, _ = utilshttp.GetUseCaseNameFromRequestContext(r)

	endpointName, _ := utilshttp.GetEndpointNameFromRequestContext(r)

	newErr := apierror.New().
		FromMsg(fmt.Sprintf("Некорректный запрос для пользовательского случая {usecase} ({endpoint}): %s", err)).
		Arguments("usecase", string(useCaseName), "endpoint", string(endpointName)).
		Component("refund-api/api/v1/usecases").
		Code("RequestError").
		TranslateRule(cmn.trRule)

	if cmn.log != nil {
		cmn.log.Error(newErr)
	}

	cmn.presenter.Error(w, r, newErr.Language(cmn.lang), http.StatusBadRequest, nil)
}

var BadExpression = func(useCaseName string, endPointName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("Bad Expression (%s: %s)", useCaseName, endPointName))
	}
}

var UseCaseNotFound = func(uCaseName api.UseCaseName) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("Не удалось найти пользовательский случай %s", uCaseName))
	}
}

var EndpointNameNotFound = func(uCaseName api.UseCaseName, endPointName api.EndpointName) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("Не удалось найти конечную точку %s для пользовательского случая %s", endPointName, uCaseName))
	}
}

func Handle(useCaseName api.UseCaseName, endPointName api.EndpointName, uCaseInstance api.UseCase, cmn *CommonUseCaseData) http.HandlerFunc {
	useCaseData, isUseCaseExists := useCases[useCaseName]
	if !isUseCaseExists {
		return UseCaseNotFound(useCaseName)
	}

	handler, isHandlerExists := useCaseData[endPointName]
	if !isHandlerExists {
		return EndpointNameNotFound(useCaseName, endPointName)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		registerUseCase(useCaseName, r)
		registerEndpoint(endPointName, r)

		data, err := cmn.presenter.ParseRequest(r)
		if err != nil {
			requestError(cmn, err, w, r)

			return
		}

		data, statusCode, err := handler(uCaseInstance, data, r)
		if err != nil {
			cmn.presenter.Error(w, r, err, statusCode, data)

			return
		}

		cmn.presenter.Response(w, r, data)
	}
}