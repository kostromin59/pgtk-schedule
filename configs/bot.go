package configs

type Bot struct {
	BotToken string `envconfig:"BOT_TOKEN" required:"true"`
	AdminID  int64  `envconfig:"ADMIN_ID" required:"true"`
	DB_CONN  string `envconfig:"DB_CONN" required:"true"`
}
