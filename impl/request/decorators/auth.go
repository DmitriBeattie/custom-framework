package auth

import (
	"encoding/json"
	"delivery-report/framework/abstract/apierror"
	"delivery-report/framework/interfaces/app"
	"delivery-report/framework/provider"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/context"
)

type IdentityAuth struct {
	baseRequest *provider.ServiceRequest
	c           *http.Client
	mu          sync.RWMutex
	log         app.Logger
	token       string
}

func IdentityAuthInstance(_baseRequest *provider.ServiceRequest, log app.Logger) *IdentityAuth {
	return &IdentityAuth{
		baseRequest: _baseRequest,
		c:           http.DefaultClient,
		log:         log,
	}
}

func IdentityAuthError() apierror.APIError {
	return apierror.New().
		Component("delivery-report/framework/auth")
}

func (i *IdentityAuth) Decorate(r *http.Request) *http.Request {
	i.mu.RLock()

	token := i.token

	i.mu.RUnlock()

	var refreshHeader bool

	if refreshAuthHeaderSign, ok := context.GetOk(r, "refreshAuthHeader"); ok {
		refreshHeader, _ = refreshAuthHeaderSign.(bool)
	}

	if token == "" || refreshHeader {
		token = i.RefreshToken()
	}

	r.Header.Set("Authorization", "Bearer "+token)

	return r
}

func (i *IdentityAuth) RefreshToken() string {
	r, err := i.baseRequest.CreateRequest(nil, nil, nil, nil, nil)
	if r.Body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		i.log.Error(IdentityAuthError().WithError(err).Code("refreshTokenCode1").FromMsg("Ошибка при создании запроса на обновление токена: {err}"))

		return ""
	}

	resp, err := i.c.Do(r)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		i.log.Error(IdentityAuthError().WithError(err).Code("refreshTokenCode2").FromMsg("Ошибка при запросе обновления токена: {err}"))

		return ""
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)

		i.log.Error(IdentityAuthError().
			Code("refreshTokenCode3").
			FromMsg("Ответ при запросе токена: {resp} (статус {status})").
			Args(apierror.ErrorArguments{"resp": string(body), "status": strconv.Itoa(resp.StatusCode)}))

		return ""
	}

	token := struct {
		Value string `json:"access_token"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		i.log.Error(IdentityAuthError().WithError(err).Code("refreshTokenCode4").FromMsg("Не удалось получить токен из ответа: {err}"))

		return ""
	}

	i.mu.Lock()
	i.token = token.Value
	i.mu.Unlock()

	return i.token
}
