package app

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
)

type TaskConfig struct {
	Name    string
	context map[interface{}]interface{}
	events Events
	ackResult  AcknowledgmentResult
}


type EventProcessor struct {
	f         EventAdapter
	eRero     EventRepository
	eCons     EventConsumer
	eventName string
	log       Logger
}

type ID interface{}

type Events map[ID]interface{}

type AcknowledgmentResult map[ID]error

type EventAdapter func(srcEvents Events) error

var WithoutAdapting EventAdapter = func(srcEvents Events) error { return nil }

type EventRepository interface {
	GetNew(task *TaskConfig, consumerName string) error
	ConfirmAck(task *TaskConfig, consumerName string) error
}

type EventConsumer interface {
	Consume(taskConfig *TaskConfig) error
	Name() string
}

func NewEventProcessor(eRepo EventRepository, eCons EventConsumer, name string, adapter EventAdapter, log Logger) *EventProcessor {
	return &EventProcessor{
		f:         adapter,
		eRero:     eRepo,
		eCons:     eCons,
		eventName: name,
		log:       log,
	}
}

func (evp *EventProcessor) error(err error) error {
	if evp.log != nil {
		evp.log.Error(err)
	}

	return err
}

func (tConf *TaskConfig) AllocateMemForEvents(cap int) {
	tConf.events = make(Events, cap)
	tConf.ackResult = make(AcknowledgmentResult, cap)
}

func (tConf *TaskConfig) WriteEvent(id ID, data interface{}) {
	tConf.events[id] = data
	tConf.ackResult[id] = nil
}

func (tConf *TaskConfig) WriteEventsFromMap(mapParam interface{}) error {
	t := reflect.ValueOf(mapParam)

	if t.Kind() != reflect.Map {
		return errors.New("Param is not a map")
	}

	tConf.AllocateMemForEvents(t.Len())

	iter := t.MapRange()
	for iter.Next() {
		tConf.WriteEvent(iter.Key().Interface(), iter.Value().Interface())
	}

	return nil
}

func (tConf *TaskConfig) WriteEventsFromStructSlice(structSlice interface{}, keyFieldName string) error {
	if reflect.Slice != reflect.TypeOf(structSlice).Kind() {
		return errors.New("Param is not a slice")
	}

	val := reflect.ValueOf(structSlice)

	tConf.AllocateMemForEvents(val.Len())

	for i := 0; i < val.Len(); i++ {
		if val.Index(i).Kind() != reflect.Struct {
			return errors.New("Slice contains type differs from struct")
		}

		keyFieldVal := val.Index(i).FieldByName(keyFieldName)

		if !keyFieldVal.IsValid() {
			return errors.New("Not found specified field in struct")
		}

		tConf.WriteEvent(keyFieldVal.Interface(), val.Index(i).Interface())
	}

	return nil
}

func (tConf *TaskConfig) BadInput(id ID, dstTypeName string) {
	tConf.SetError(id, fmt.Errorf("Error while trying to convert source msg to %s", dstTypeName))
}

func (tConf *TaskConfig) WriteEventsFromSliceOfPtrToStruct(sliceOfPtrToStruct interface{}, keyFieldName string) error {
	if reflect.Slice != reflect.TypeOf(sliceOfPtrToStruct).Kind() {
		return errors.New("Param is not a slice")
	}

	val := reflect.ValueOf(sliceOfPtrToStruct)

	tConf.AllocateMemForEvents(val.Len())

	for i := 0; i < val.Len(); i++ {
		if val.Index(i).Kind() != reflect.Ptr {
			return errors.New("Slice contains type differs from ptr")
		}

		if val.Index(i).Elem().Kind() != reflect.Struct {
			return errors.New("Ptr isn't points to struct")
		}

		keyFieldVal := val.Index(i).Elem().FieldByName(keyFieldName)

		if !keyFieldVal.IsValid() {
			return errors.New("Not found specified field in struct")
		}

		tConf.WriteEvent(keyFieldVal.Interface(), val.Index(i).Interface())
	}

	return nil
}

func (tConf *TaskConfig) WriteCtx(key, value interface{}) {
	tConf.context[key] = value
}

func (tConf *TaskConfig) SetError(id ID, err error) {
	tConf.ackResult[id] = err
}

func (tConf *TaskConfig) SetErrorOK(id ID, err error) bool {
	_, ok := tConf.ackResult[id]
	if !ok {
		return false
	}

	tConf.SetError(id, err)

	return true
}

func (tConf *TaskConfig) Len() int {
	return len(tConf.events)
}

func (tConf *TaskConfig) ShowEvents() Events {
	return tConf.events
}

func (tConf *TaskConfig) ShowConsumigResult() AcknowledgmentResult {
	return tConf.ackResult
}

func (tConf *TaskConfig) SetErrorToAll(err error) {
	for id := range tConf.ackResult {
		tConf.ackResult[id] = err
	}
}

func (tConf *TaskConfig) GetContext(id interface{}) interface{} {
	return tConf.context[id]
}

func (tConf *TaskConfig) GetContextOK(id interface{}) (interface{}, bool) {
	val, ok := tConf.context[id]

	return val, ok
}

func (evp *EventProcessor) logNotProcessedEvents(config *TaskConfig) {
	if evp.log == nil {
		return
	}

	var errText string

	errMap := make(map[string][]ID)

	for id, err := range config.ShowConsumigResult() {
		if err != nil {
			errMap[err.Error()] = append(errMap[err.Error()], id)
		}
	}

	if len(errMap) == 0 {
		return
	}

	for errStr, ids := range errMap {
		errText += ". " + errStr + ": "

		for i := range ids {
			var divider string

			if i > 0 {
				divider = "; "
			}

			errText += fmt.Sprintf("%s%v", divider, ids[i])
		}
	}

	evp.log.Info(errText)
}

func (evp *EventProcessor) Process(needLogNotProcessedEvents bool) error {
	defer func() {
		if rec := recover(); rec != nil {
			err := fmt.Sprint(rec)

			evp.log.Error(fmt.Errorf("Panic while executing %s: %s. %s", evp.eventName, err, string(debug.Stack())))
		}
	}()

	conf := &TaskConfig{
		Name:    evp.eventName,
		context: make(map[interface{}]interface{}),
	}

	if err := evp.eRero.GetNew(conf, evp.eCons.Name()); err != nil {
		return evp.error(err)
	}

	if len(conf.events) == 0 {
		return nil
	}

	if evp.f == nil {
		return evp.error(errors.New("No adapter provided"))
	}

	if err := evp.f(conf.events); err != nil {
		return evp.error(err)
	}

	if err := evp.eCons.Consume(conf); err != nil {
		return evp.error(err)
	}

	if needLogNotProcessedEvents {
		evp.logNotProcessedEvents(conf)
	}

	if err := evp.eRero.ConfirmAck(conf, evp.eCons.Name()); err != nil {
		return evp.error(err)
	}

	return nil
}

func WrapperFunc(f func(msg interface{}) (interface{}, bool, error)) EventAdapter {
	return func(srcEvents Events) error {
		for key, val := range srcEvents {
			adaptedMsg, isMsgFit, err := f(val)
			if err != nil {
				return err
			}

			if !isMsgFit {
				return errors.New("Bad input")
			}

			srcEvents[key] = adaptedMsg
		}

		return nil
	}
}

type EventConsumerExtension struct {
	EventConsumer
	IsResultMatter bool
}

type EventConsumerArr []EventConsumerExtension

func (cons EventConsumerArr) Consume(conf *TaskConfig) error {
	copyConfig := *conf

	for i := range cons {
		var consConf *TaskConfig

		if cons[i].IsResultMatter {
			consConf = conf
		} else {
			consConf = new(TaskConfig)
			*consConf = copyConfig
		}

		if err := cons[i].Consume(conf); err != nil {
			if cons[i].IsResultMatter {
				return err
			}
		}
	}

	return nil
}

func (cons EventConsumerArr) Name() string {
	var nm string

	for i := range cons {
		nm += "Name:" + cons[i].Name() + ". "
	}

	return nm
}
