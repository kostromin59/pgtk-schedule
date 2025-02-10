package main

import (
	"log"
	"pgtk-schedule/configs"
	"pgtk-schedule/internal/apps/bot"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err.Error())
	}

	var cfg configs.Bot
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err.Error())
	}

	if err := bot.Run(cfg); err != nil {
		log.Fatal(err.Error())
	}
}
