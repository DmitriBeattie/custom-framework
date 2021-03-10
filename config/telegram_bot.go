package config

type TelegramBotConfig struct {
	Token    string  `json:"token"`
	Channels []int64 `json:"channels"`
}

func (tb *TelegramBotConfig) InstanceKind() string {
	return "telegram-bot"
}
