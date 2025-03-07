package main

import (
    "fmt"
    "time"
    "log"
    "os"
    "os/signal"
    "github.com/bwmarrin/discordgo"
    "context"
    "encoding/json"
    "errors"
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

func playersFromFile(path string) ([]NbaPlayer, error) {
    file, err := os.Open(path)

    if err != nil {
        return nil, err
    }

    defer file.Close()

    var players []NbaPlayer

    if err := json.NewDecoder(file).Decode(&players); err != nil {
        return nil, err
    }

    return players, nil
}

func (b *NbaFantasyBot) Init(ctx context.Context) error {
    lockTime, err := b.client.getLockTime(ctx)

    if err != nil {
        return err
    }

    log.Println(lockTime)

    t, err := time.Parse(time.RFC3339, lockTime.Time)

    if err != nil {
        return err
    }

    log.Println(t)

    date := time.Now().Format(time.DateOnly)

    fmt.Println(date)

    players, err := b.client.getTodaysPlayers(ctx)

    if errors.Is(err, ErrRosterLocked) {
        log.Println("Using cached players")

        players, err = playersFromFile("cache/players.json")

        if err != nil {
            return err
        }
    } else if err != nil {
        return err
    }

    b.cache = newPlayerCache(players)
    b.addCommand(createSetRosterCommand())
    b.registerHandler("set-roster", setRosterHandler(b))

    b.addCommand(createGetGlobalRosterCommand())
    b.registerHandler("global-roster", getGlobalRosterCommandHandler(b))


    b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        log.Println("Bot is up")
    })

    b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if h, ok := b.cmdHandlers[i.ApplicationCommandData().Name]; ok {
            h(s, i)
        }
    })

    return nil
}

func (b *NbaFantasyBot) Run(ctx context.Context) error {
    if err := b.Open(); err != nil {
        return err
    }

    defer b.Close()

    s := b.session

    _, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", b.cmds)

    if err != nil {
        return err
    }

    stop := make(chan os.Signal, 1)

    signal.Notify(stop, os.Interrupt)
    <-stop

    log.Println("Gracefully shutting down")

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

