package loggers

import (
	"encoding/json"
	"fmt"

	"github.com/ivpusic/golog"
)

type StdoutLogger struct{}

func NewStdoutLogger() StdoutLogger {
	return StdoutLogger{}
}

func (s StdoutLogger) Append(log golog.Log) {
	b, _ := json.Marshal(log)

	fmt.Println(string(b))
}

func (s StdoutLogger) Id() string {
	return "Stdout"
}
