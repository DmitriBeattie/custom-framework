package file

import (
	"encoding/json"
	"github.com/DmitriBeattie/custom-framework/interfaces/translator"
	"io/ioutil"
)

type JSONFileTranslation struct {
	Data map[string]map[string]string
}

func OpenTranslationFile(path string) *JSONFileTranslation {
	var data JSONFileTranslation

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(b, &data.Data); err != nil {
		return nil
	}

	return &data
}

func (f *JSONFileTranslation) TranslateOK(msg string, curLang, newLang translator.Language) (res string, isOk bool) {
	defer func() {
		if rec := recover(); rec != nil {
			res = msg
		}
	}()

	transSlice, ok := f.Data[msg]
	if !ok {
		return msg, false
	}

	res, isOk = transSlice[string(newLang)]
	if !ok {
		return msg, false
	}

	return res, true
}

func (f *JSONFileTranslation) Translate(msg string, curLang, newLang translator.Language) string {
	res, _ := f.TranslateOK(msg, curLang, newLang)

	return res
}
