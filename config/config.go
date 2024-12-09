package config

type config struct {
	ApiID          int32    `envconfig:"API_ID" required:"true"`
	ApiHash        string   `envconfig:"API_HASH" required:"true"`
	BotToken       string   `envconfig:"BOT_TOKEN" required:"true"`
	StringSessions []string `envconfig:"STRING_SESSIONS" required:"true"`
}
