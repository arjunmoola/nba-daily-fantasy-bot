package main

import (
    "log"
    //"os"
    "github.com/joho/godotenv"
    "context"
)


func main() {
    if err := godotenv.Load(); err != nil {
        log.Panic(err)
    }

    bot, err := newNbaFantasyBot()

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
