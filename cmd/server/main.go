package main

import (
    "log"
    //"os"
    "github.com/joho/godotenv"
    "context"
    "github.com/arjunmoola/nba-daily-fantasy-bot/internal/bot"
	"github.com/arjunmoola/nba-daily-fantasy-bot/internal/config"
)


func main() {
    if err := godotenv.Load(); err != nil {
        log.Panic(err)
    }

	config := config.New()

    picknRollBot := bot.New(config)

    if err := picknRollBot.Init(context.Background()); err != nil {
        log.Panic(err)
    }

    if err := picknRollBot.Run(context.Background()); err != nil {
        log.Panic(err)
    }
}
