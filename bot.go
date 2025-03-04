package main

import (
    "fmt"
    "os"
    "github.com/bwmarrin/discordgo"
    "context"
)

var (
    commands = []*discordgo.ApplicationCommand{
        {
            Name: "set-roster",
            Description: "set roster",
            Type: discordgo.ChatApplicationCommand,
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Name: "PG",
                    Description: "Choose PG",
                    Type: discordgo.ApplicationCommandOptionString,
                    Required: true,
                    Autocomplete: true,
                },
            },
        },
    }

    commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        "set-roster": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

        },
    }
)

type NbaFantasyBot struct {
    client *NbaFantasyClient
    session *discordgo.Session
    cache *PlayerCache

    cmds []*discordgo.ApplicationCommand
    cmdHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
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
        client: newNbaFantasyClient(),
        cmdHandlers: make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate)),
    }

    return &bot, nil
}

func (b *NbaFantasyBot) Init(ctx context.Context) error {
    players, err := b.client.getTodaysPlayers(ctx)

    if err != nil {
        return err
    }
    b.cache = newPlayerCache(players)
    return nil
}

func (b *NbaFantasyBot) Run(ctx context.Context) error {
    if err := b.Open(); err != nil {
        return err
    }

    defer b.Close()

    return nil
}

func (b *NbaFantasyBot) Open() error {
    return b.session.Open()
}

func (b *NbaFantasyBot) Close() error {
    return b.session.Close()
}

func (b *NbaFantasyBot) addCommand(cmd *discordgo.ApplicationCommand) {
    b.cmds = append(b.cmds, cmd)
}

func (b *NbaFantasyBot) registerHandler(name string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
    b.cmdHandlers[name] = handler
}
