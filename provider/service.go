package provider

import (
	"delivery-report/framework/interfaces/request"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/context"
)

type ServiceRequest struct {
	U       *url.URL
	Payload *string
	Method  string
	Header  http.Header
	Dec     request.RequestDecorator
}

func (s *ServiceRequest) CreateRequest(body io.Reader, additionalQuery url.Values, additionalHeader http.Header, urlPatternReplacement map[string]string, ctx map[interface{}]interface{}) (*http.Request, error) {
	if s == nil {
		return nil, errors.New("Instance is nil")
	}

	sCopy := *s

	if sCopy.U == nil {
		return nil, errors.New("Url is not set")
	}

	if len(urlPatternReplacement) > 0 {
		urlRaw, err := url.PathUnescape(sCopy.U.String())
		if err != nil {
			urlRaw = s.U.String()
		}

		for pattern, replacement := range urlPatternReplacement {
			urlRaw = strings.Replace(urlRaw, "{"+pattern+"}", replacement, -1)
		}

		sCopy.U, _ = url.Parse(urlRaw)
	}

	q := sCopy.U.Query()
	for key, value := range additionalQuery {
		q[key] = value
	}
	sCopy.U.RawQuery = q.Encode()

	if sCopy.Payload != nil {
		body = strings.NewReader(*sCopy.Payload)
	}

	req, err := http.NewRequest(sCopy.Method, sCopy.U.String(), body)
	if err != nil {
		return req, err
	}

	header := make(http.Header)

	for defaultKey, defaultVal := range sCopy.Header {
		header[defaultKey] = defaultVal
	}

	for key, value := range additionalHeader {
		header[key] = value
	}

	if len(header) > 0 {
		req.Header = header
	}

	for ctxHeader, ctxValue := range ctx {
		context.Set(req, ctxHeader, ctxValue)
	}

	if sCopy.Dec != nil {
		req = sCopy.Dec.Decorate(req)
	}

	return req, nil
}

