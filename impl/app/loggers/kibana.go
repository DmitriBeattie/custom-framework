package loggers

import (
	"github.com/DmitriBeattie/custom-framework/provider"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ivpusic/golog"
)

type KibanaLogger struct {
	s           *provider.ServiceRequest
	AppName     string
	Environment string
	AppID       string
}

type LoggerData struct {
	Fields EnviromentInfo `json:"fields"`
	Data   golog.Log      `json:"data"`
}

type EnviromentInfo struct {
	ApplicationName string `json:"ApplicationName,omitempty"`
	Environment     string `json:"Environment,omitempty"`
	Level           string `json:"Level,omitempty"`
}

func InitKibanaLogger(_s *provider.ServiceRequest, appName string, environment string, appID string) *KibanaLogger {
	return &KibanaLogger{
		s:           _s,
		AppName:     appName,
		Environment: environment,
		AppID:       appID,
	}
}

func (kb *KibanaLogger) Append(log golog.Log) {
	kibanaLog := LoggerData{
		Fields: EnviromentInfo{
			ApplicationName: kb.AppName,
			Environment:     fmt.Sprintf("%s %s", kb.Environment, kb.AppID),
			Level:           log.Level.Name,
		},
		Data: log,
	}

	kibanaLogJSON, _ := json.Marshal(kibanaLog)

	client := &http.Client{}

	req, _ := kb.s.CreateRequest(bytes.NewBuffer(kibanaLogJSON), nil, nil, nil, nil)
	if req != nil && req.Body != nil {
		defer req.Body.Close()
	}

	/*Сами пишите в эту сраную польскую кибану*/
	if req == nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, _ := client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
}

func (kb *KibanaLogger) Id() string {
	return "kibana"
}
