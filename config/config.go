package config

type Config struct {
	ApiID          int32    `envconfig:"API_ID" required:"true"`
	ApiHash        string   `envconfig:"API_HASH" required:"true"`
	BotToken       string   `envconfig:"BOT_TOKEN" required:"true"`
	ChannelID      int64    `envconfig:"CHANNEL_ID" required:"true"`
	StringSessions []string `envconfig:"STRING_SESSIONS" required:"true"`
}
