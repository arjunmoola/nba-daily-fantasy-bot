package main

import (
    "log"
    //"os"
    "github.com/joho/godotenv"
    "context"
    dailyfantasybot "github.com/arjunmoola/nba-daily-fantasy-bot/internal/daily-fantasy-bot"
)


func main() {
    if err := godotenv.Load(); err != nil {
        log.Panic(err)
    }

    bot, err := dailyfantasybot.NewNbaFantasyBot()

    if err != nil {
        log.Panic(err)
    }

    if err := bot.Init(context.Background()); err != nil {
        log.Panic(err)
    }

    if err := bot.Run(context.Background()); err != nil {
        log.Panic(err)
    }
}
