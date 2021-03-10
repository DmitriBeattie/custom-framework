package loggers

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	yaml "gopkg.in/yaml.v2"

	"strings"

	"github.com/ivpusic/golog"
)

type TelegramBotLogger struct {
	ChatIDs     []int64
	bot         *tgbotapi.BotAPI
	AppID       string
	Environment string
	Instance    string
}

func InitTelegramBot(_token string, _channels []int64, _appID string, _environmentID string, _instance string) (*TelegramBotLogger, error) {
	bot, err := tgbotapi.NewBotAPI(_token)
	if err != nil {
		return nil, err
	}

	bot.Debug = false

	return &TelegramBotLogger{
		ChatIDs:     _channels,
		bot:         bot,
		AppID:       _appID,
		Environment: _environmentID,
		Instance:    _instance,
	}, nil
}

func splitMsg(input string) (result []string) {
	separator := "\n"
	msgLength := 4096

	for {
		if len(input) <= msgLength {
			result = append(result, input)

			break
		}

		indx := strings.LastIndex(input[:msgLength], separator)

		if indx == -1 {
			result = append(result, input[:msgLength])
			input = input[msgLength:]
		} else {
			result = append(result, input[:indx])
			input = input[indx+len(separator):]
		}
	}

	return result
}

func prepareMsg(message string, chatID int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "markdown"
	return msg
}

func (tb *TelegramBotLogger) Listen() (tgbotapi.UpdatesChannel, error) {
	var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60

	return tb.bot.GetUpdatesChan(ucfg)
}

func (tb *TelegramBotLogger) SendMsg(channel int64, msg string) {
	arrayMsgs := splitMsg(msg)

	for i := range tb.ChatIDs {
		if tb.ChatIDs[i] == channel {
			for i := range arrayMsgs {

				m := prepareMsg(arrayMsgs[i], channel)
				m.ParseMode = ""

				tb.bot.Send(m)
			}

			break
		}
	}
}

func (tb *TelegramBotLogger) Append(log golog.Log) {
	msgInfo := struct {
		AppID       string      `yaml:"*Приложение*"`
		Environment string      `yaml:"*Окружение*"`
		Instance    string      `yaml:"*Экземпляр*"`
		State       golog.Level `yaml:"*Уровень*"`
		Time        string      `yaml:"*Время*"`
		Msg         string      `yaml:"*Cообщение*"`
	}{
		tb.AppID,
		tb.Environment,
		tb.Instance,
		log.Level,
		log.Time.Format("2006-01-02 15:04:05"),
		log.Message,
	}

	msgByte, _ := yaml.Marshal(msgInfo)

	msgStr := string(msgByte)

	arrayMsgs := splitMsg(msgStr)

	telegramMsgs := make([]tgbotapi.Chattable, 0, len(arrayMsgs)*len(tb.ChatIDs))

	for i := range tb.ChatIDs {
		for j := range arrayMsgs {
			telegramMsgs = append(telegramMsgs, prepareMsg(arrayMsgs[j], tb.ChatIDs[i]))
		}
	}

	for i := range telegramMsgs {
		tb.bot.Send(telegramMsgs[i])
	}
}

func (kb *TelegramBotLogger) Id() string {
	return "bot"
}
