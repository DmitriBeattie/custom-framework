package workers

import "fmt"

//WorkerError описывает ошибки в процессе выполнения воркеров
type WorkerError map[string]error

func (w WorkerError) Error() string {
	var errMsg string

	for workName, err := range w {
		errMsg += fmt.Sprintf("Ошибка воркера %s: %s\n", workName, err)
	}

	return errMsg
}
