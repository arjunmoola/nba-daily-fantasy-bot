package main

import (
    "fmt"
    "log"
    "os"
    "github.com/bwmarrin/discordgo"
    "github.com/joho/godotenv"
)

type NbaFantasyBot struct {
    //client *NbaFantasyClient
    session *discordgo.Session
}

func newNbaFantasyBot() (*NbaFantasyBot, error) {
    token := os.Getenv("TOKEN")

    if token == "" {
        return nil, fmt.Errorf("discord token is not set")
    }

    session, err := discordgo.New("Bot " + token)

    if err != nil {
        return nil, err
    }

    bot := NbaFantasyBot{
        session: session,
    }

    return &bot, nil
}

func (b *NbaFantasyBot) Open() error {
    return b.session.Open()
}

func (b *NbaFantasyBot) Close() error {
    return b.session.Close()
}

func main() {
    if err := godotenv.Load(); err != nil {
        log.Panic(err)
    }

    bot, err := newNbaFantasyBot()

    if err != nil {
        log.Panic(err)
    }

    if err := bot.Open(); err != nil {
        log.Panic(err)
    }

    defer bot.Close()

    fmt.Println("hello world")
}
