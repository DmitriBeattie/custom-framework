package workers

import (
	"fmt"
	"delivery-report/framework/interfaces/app"
	"runtime/debug"
	"time"
)

type action func() error

var workers = map[string]*workerSetting{}

//WorkerSetting содержит информацию о воркере
type workerSetting struct {
	//action описывает действие воркера
	action

	//s - расписание воркера
	s *Schedule

	//restart для сигнализации о том, что
	//worker следует запустить заново
	restart chan interface{}

	//repair - канал для информировании о
	//том, что были исправлены ошибки и воркер
	//следует запустить заново
	repair chan interface{}

	//err сохраняет ошибки в ходе выполнения
	//воркеров
	err chan error
}

//RegisterWork регистрирутет воркер
func RegisterWork(name string, f func() error, s *Schedule) {
	workers[name] = &workerSetting{
		action:  action(f),
		restart: make(chan interface{}),
		repair:  make(chan interface{}),
		s:       s,
		err:     make(chan error, 1),
	}
}

func GetVersion(name string) int {
	if s, ok := workers[name]; ok {
		return s.s.Version
	}

	return -1
}

func checkIsWorkerExists(workerName string) (*workerSetting, error) {
	val, ok := workers[workerName]
	if !ok {
		return nil, fmt.Errorf("Не найден воркер %s", workerName)
	}

	return val, nil
}

func updSchedule(name string, s *Schedule) (*workerSetting, error) {
	val, err := checkIsWorkerExists(name)
	if err != nil {
		return nil, err
	}

	val.s = s

	return val, nil
}

//UpdScheduleWithTimeout сохраняет измененное расписание воркера
//и пытается уведомить об измерении в течении timeout
func UpdScheduleWithTimeout(name string, s *Schedule, timeout time.Duration) error {
	val, err := updSchedule(name, s)
	if err != nil {
		return err
	}

	select {
	case val.restart <- true:
	case <-time.After(timeout):
	}

	return nil
}

//UpdSchedule сохраняет измененное расписание.
func UpdSchedule(name string, s *Schedule) error {
	val, err := updSchedule(name, s)
	if err != nil {
		return err
	}

	val.restart <- true

	return nil
}

//RenewScheduleWithTimeout обновляет сразу несколько расписаний
func RenewScheduleWithTimeout(newData map[string]*Schedule, timeout time.Duration) {
	for i := range newData {
		if _, ok := workers[i]; ok {
			workers[i].s = newData[i]

			select {
			case workers[i].restart <- true:
			case <-time.After(timeout):
			}
		}
	}
}

//RenewSchedule обновляет сразу несколько расписаний
func RenewSchedule(newData map[string]*Schedule, timeout time.Duration) {
	for i := range newData {
		if _, ok := workers[i]; ok {
			workers[i].s = newData[i]

			workers[i].restart <- true
		}
	}
}

//RepairWorkWithTimeout посылает сигнал о том, что воркеры поччинены
func RepairWorkWithTimeout(name string, timeout time.Duration) error {
	val, err := checkIsWorkerExists(name)
	if err != nil {
		return err
	}

	select {
	case val.repair <- true:
	case <-time.After(timeout):
	}

	return nil
}

func getWorkerSetting(name string) (*workerSetting, error) {
	val, err := checkIsWorkerExists(name)
	if err != nil {
		return nil, err
	}

	return val, nil
}

//ExecWorkers запускает процесс выполнения воркеров
func ExecWorkers(logger app.Logger) {
	for name := range workers {
		go func(n string) {
			defer func() {
				if rec := recover(); rec != nil {
					err := fmt.Sprint(rec)

					logger.Error(fmt.Errorf("Паника при выполнении %s: %s. %s", n, err, string(debug.Stack())))

					panic(err)
				}
			}()

			for {
				execWorkers(n)
			}
		}(name)
	}
}

//execWorkers вызывает воркер с имененем workerName
//в соответствие с его настройками
func execWorkers(workerName string) {
	wS, _ := getWorkerSetting(workerName)

	if wS.action == nil {
		wS.err <- fmt.Errorf("Не найдено действия для %s", workerName)

		return
	}

	if wS.s == nil {
		wS.err <- fmt.Errorf("Не найдено расписания для %s", workerName)

		<-wS.restart

		return
	}

	if wS.s.IsActive == false {
		wS.err <- fmt.Errorf("Воркер %s не активен. Ждем обновлений", workerName)

		<-wS.restart

		return
	}

	//Расчитываем задержку до запуска воркера
	delay, isActive := wS.s.calculateDelay()

	//Не удалось рассчитать задержку, ждем обновлений
	if isActive == false {
		<-wS.restart

		return
	}

	//Запускаем воркер
	select {
	case <-wS.restart:
		return
	case <-time.After(delay):
		if err := wS.action(); err != nil {
			//Если предыдущая ошибка не прочитана, то воркер перестает быть активным.
			//Поэтому рекомендуется регистрировать обработчик ошибок RegisterGetErrorAction
			wS.err <- err

			select {
			case <-wS.repair:
			case <-time.After(wS.s.FailureTimeOut):
			}
		}
	}

	select {
	case <-wS.restart:
	case <-time.After(time.Duration(wS.s.DelaySeconds) * time.Second):
	}
}

//GetError чтение ошибок воркера
func GetError() error {
	err := WorkerError{}

	for i := range workers {
		select {
		case e := <-workers[i].err:
			err[i] = e
		default:
		}
	}

	if len(err) == 0 {
		return nil
	}

	return err
}

func RegisterGetErrorAction(f func(err error), delaySec uint64) error {
	s := EveryDay().WithDelay(delaySec).SetIsActive(true)

	RegisterWork("errors", getErrorAction(f), s)

	return nil
}

func getErrorAction(f func(err error)) action {
	return func() error {
		err := GetError()
		if err != nil {
			f(err)
		}

		return nil
	}
}

func GetWorkerInfo() interface{} {
	var data []interface{}

	for key, value := range workers {
		/*if key == "errors" {
			continue
		}*/

		s := struct {
			WorkerName     string        `json:"workerName"`
			WorkerVersion  int           `json:"workerVersion"`
			IsActive       bool          `json:"isActive"`
			DelaySeconds   uint64        `json:"delaySeconds"`
			FailureTimeOut time.Duration `json:"failureTimeOut"`
			Schedule       map[time.Weekday]struct {
				HourFrom *uint8 `json:"hourFrom"`
				HourTo   *uint8 `json:"hourTo"`
			}
		}{}

		s.WorkerName = key

		if value.s != nil {
			s.WorkerVersion = value.s.Version
			s.IsActive = value.s.IsActive
			s.DelaySeconds = value.s.DelaySeconds
			s.FailureTimeOut = value.s.FailureTimeOut

			s.Schedule = map[time.Weekday]struct {
				HourFrom *uint8 `json:"hourFrom"`
				HourTo   *uint8 `json:"hourTo"`
			}{}

			for d, h := range value.s.Days {
				s.Schedule[d] = struct {
					HourFrom *uint8 `json:"hourFrom"`
					HourTo   *uint8 `json:"hourTo"`
				}{HourFrom: h.HourFrom, HourTo: h.HourTo}
			}
		}

		data = append(data, s)
	}
	return data
}
