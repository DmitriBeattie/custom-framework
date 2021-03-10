package config

import (
	"net/http"
	"net/url"
)

type ServiceDiscovery struct {
	Path        string      `json:"path"`
	Description string      `json:"description"`
	Method      string      `json:"method"`
	Query       url.Values  `json:"query,omitempty"`
	Header      http.Header `json:"header,omitempty"`
	Payload     *string     `json:"payload,omitempty"`
}

type Service struct {
	Url
	Discovery map[string]*ServiceDiscovery `json:"services"`
}

func (s *Service) InstanceKind() string {
	return "service"
}

func (s *Service) GetURL(serviceName string) (u *url.URL, isFound bool) {
	srv, found := s.Discovery[serviceName]
	if !found {
		return nil, false
	}

	return &url.URL{
		Scheme:   s.Scheme,
		Host:     s.Host,
		Path:     s.Path + srv.Path,
		RawQuery: srv.Query.Encode(),
	}, true
}

func (s *Service) GetRequestData(serviceName string) (sd *ServiceDiscovery, isFound bool) {
	sd, isFound = s.Discovery[serviceName]

	return sd, isFound
}
