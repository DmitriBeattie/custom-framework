package apierror

import (
	"fmt"
	"delivery-report/framework/interfaces/translator"
	"strconv"
	"strings"
)

type ErrorArguments map[string]string

//APIError - структура для описания ошибки приложения
type APIError struct {
	//msg - сообщение с шаблонами для замены из Args
	msg string

	//extError ошибка полученная с уровня ниже
	extError error

	//component характеризует место в приложении, в котором
	//формируется ошибка
	component string

	//code код ошибки
	code string

	//Args для подстановки аргументов в текст ошибки
	args ErrorArguments

	//lang определяет язык сообщеня
	lang translator.Language

	//tr определяет правила перевода
	tr translator.Translator
}

func (e APIError) ID() string {
	//return e.component + ":" + e.code

	return e.code
}

func (e APIError) Error() string {
	var msg string
	var isTranslated bool

	if e.tr != nil {
		msg, isTranslated = e.tr.TranslateOK(e.ID(), "", e.lang)
	}

	if !isTranslated {
		msg = e.msg
	}

	if e.extError != nil {
		var err error

		baseErr, ok := e.extError.(APIError)
		if ok {
			baseErr.lang = e.lang
			baseErr.tr = e.tr

			err = baseErr
		} else {
			err = e.extError
		}

		msg = strings.Replace(msg, "{err}", err.Error(), -1)
	}

	for pattern, value := range e.args {
		msg = strings.Replace(msg, "{"+pattern+"}", value, -1)
	}

	if msg == "" {
		msg = e.ID()
		if e.extError != nil {
			msg += ": " + e.extError.Error()
		}
	}

	return msg
}

func New() APIError {
	return APIError{}
}

func (e APIError) FromMsg(msg string) APIError {
	e.msg = msg

	return e
}

func (e APIError) FromErr(err error) APIError {
	e.msg = err.Error()

	return e
}

func (e APIError) WithError(err error) APIError {
	e.extError = err

	return e
}

func (e APIError) Code(code string) APIError {
	e.code = code

	return e
}

func (e APIError) Component(cmp string) APIError {
	e.component = cmp

	return e
}

func (e APIError) Args(args ErrorArguments) APIError {
	e.args = args

	return e
}

func (e APIError) Language(lang translator.Language) APIError {
	e.lang = lang

	return e
}

func (e APIError) TranslateRule(trRule translator.Translator) APIError {
	e.tr = trRule

	return e
}

func (e APIError) Arguments(args ...interface{}) APIError {
	a := make(ErrorArguments, len(args)/2)

	var key, val string

	for i := range args {
		if i == 0 || i%2 == 0 {
			var ok bool

			key, ok = args[i].(string)
			if !ok {
				key = fmt.Sprintf("Arg%d", i/2)
			}

			continue
		}

		switch valInterface := args[i].(type) {
		case string:
			val = valInterface
		case int64:
			val = strconv.FormatInt(valInterface, 10)
		case uint64:
			val = strconv.FormatUint(valInterface, 10)
		case bool:
			val = strconv.FormatBool(valInterface)
		case float64:
			val = strconv.FormatFloat(valInterface, 'f', 2, 64)
		default:
			val = "unknown format"
		}

		a[key] = val
	}

	return e.Args(a)
}
