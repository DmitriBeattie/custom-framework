package app

type Logger interface {
	Info(msg interface{}, data ...interface{})
	Error(msg interface{}, data ...interface{})
}
