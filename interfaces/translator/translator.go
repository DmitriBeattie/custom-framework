package translator

//Language - язык (локаль) перевода
type Language string

//Translator предоставляет интерфейс перевода сообщения c одного языка на другой
type Translator interface {
	//Translate переводит msg c языка curLang на newlang и возвращает итоговый результат
	Translate(msg string, curLang, newlang Language) string

	//TranslateOK переводит msg c языка curLang на newlang и возвращает итоговый результат
	//с пометкой перевод удался или не удался
	TranslateOK(msg string, curLang, newlang Language) (string, bool)
}
