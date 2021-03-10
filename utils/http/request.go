package http

import (
	"github.com/DmitriBeattie/custom-framework/interfaces/api"
	"github.com/DmitriBeattie/custom-framework/interfaces/translator"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sarulabs/di"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

var DefaultLanguage translator.Language = "ru"

func GetUseCaseNameFromRequestContext(r *http.Request) (api.UseCaseName, error) {
	useCase, ok := context.GetOk(r, api.UseCaseName("use-case"))
	if !ok {
		return "", fmt.Errorf("Use case is not found in request context")
	}

	useCaseName, ok := useCase.(api.UseCaseName)
	if !ok {
		return "", fmt.Errorf("Use case is not type api.UseCaseName")
	}

	return useCaseName, nil
}

func GetEndpointNameFromRequestContext(r *http.Request) (api.EndpointName, error) {
	endPoint, ok := context.GetOk(r, api.EndpointName("endpoint"))
	if !ok {
		return "", fmt.Errorf("Endpoint is not found in request context")
	}

	endPointName, ok := endPoint.(api.EndpointName)
	if !ok {
		return "", fmt.Errorf("Endpoint is not api.EndpointName")
	}

	return endPointName, nil
}

func GetDIContainerFromRequestContext(r *http.Request) (di.Container, error) {
	ctn, ok := context.GetOk(r, di.ContainerKey("di"))
	if !ok {
		return nil, errors.New("Not found di container in request")
	}

	appCtn, ok := ctn.(di.Container)
	if !ok {
		return nil, errors.New("Di is not di.Container")
	}

	return appCtn, nil
}

func GetLocaleFromRequest(r *http.Request) translator.Language {
	lang, ok := context.GetOk(r, "lang")
	if !ok {
		return DefaultLanguage
	}

	l, ok := lang.(translator.Language)
	if !ok {
		return DefaultLanguage
	}

	return l
}

func GetIDFromPath(r *http.Request, keyName string) (int, error) {
	vars := mux.Vars(r)

	officeid, err := strconv.ParseInt(vars[keyName], 10, 32)
	if err != nil {
		return -1, errors.New("Код должен быть числом")
	}

	return int(officeid), nil
}

func GetStringFromPath(r *http.Request, keyName string) string {
	vars := mux.Vars(r)

	return vars[keyName]
}

func GetEmployeeFromContext(req *http.Request) (int, error) {
	employeeID := context.Get(req, "employeeID")
	if employeeID == nil {
		return -1, errors.New("Не удалось получить сотрудника!")
	}

	employeeIDInt64, ok := isValueNumeric(employeeID.(string))
	if !ok {
		return -1, fmt.Errorf("Некорректный код сотрудника: %s", employeeID.(string))
	}

	return int(employeeIDInt64), nil
}

func isValueNumeric(value string) (int64, bool) {
	valInt64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return -1, false
	}

	return valInt64, true
}

func GetNumValueFromQuery(req *http.Request, param string) (int64, bool) {
	data, ok := GetValueFromQuery(req, param)
	if !ok {
		return -1, false
	}

	val, isNumeric := isValueNumeric(data)
	if !isNumeric {
		return -2, false
	}

	return val, true
}

func GetValueFromQuery(req *http.Request, param string) (string, bool) {
	query := req.URL.Query()

	data, ok := query[param]
	if !ok {
		return "", false
	}

	if len(data) < 1 {
		return "", false
	}

	return data[0], true
}

func GetBoolValueFromQuery(req *http.Request, param string) (bool, bool) {
	data, ok := GetValueFromQuery(req, param)
	if !ok {
		return false, false
	}

	if data == "true" {
		return true, true
	}

	if data == "false" {
		return false, true
	}

	return false, false
}