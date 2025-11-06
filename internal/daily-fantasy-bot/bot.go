package dailyfantasybot

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
    clientpkg "github.com/arjunmoola/nba-daily-fantasy-bot/internal/client"
    typespkg "github.com/arjunmoola/nba-daily-fantasy-bot/internal/types"
)

type NbaFantasyBot struct {
    client *clientpkg.NbaFantasyClient
    session *discordgo.Session
    cache *PlayerCache

    cmds []*discordgo.ApplicationCommand
    cmdHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
}

func NewNbaFantasyBot() (*NbaFantasyBot, error) {
    token := os.Getenv("TOKEN")

    if token == "" {
        return nil, fmt.Errorf("discord token is not set")
    }

    session, err := discordgo.New("Bot " + token)

    if err != nil {
		return nil, fmt.Errorf("failed to get session %v", err)
    }
	
	client, err := clientpkg.NewNbaFantasyClient()

	if err != nil {
		return nil, fmt.Errorf("unable to setup http client %v", err)
	}

    bot := NbaFantasyBot{
        session: session,
		client: client,
        cmdHandlers: make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate)),
    }

	fmt.Println("new session created")

    return &bot, nil
}

func playersFromFile(path string) ([]typespkg.NbaPlayer, error) {
    file, err := os.Open(path)

    if err != nil {
        return nil, err
    }

    defer file.Close()

    var players []typespkg.NbaPlayer

    if err := json.NewDecoder(file).Decode(&players); err != nil {
        return nil, err
    }

    return players, nil
}

func (b *NbaFantasyBot) Init(ctx context.Context) error {
    lockTime, err := b.client.Activity.GetLockTime(ctx)

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

    players, err := b.client.Activity.GetTodaysPlayers(ctx)

    if errors.Is(err, clientpkg.ErrRosterLocked) {
        log.Println("Using cached players")

        players, err = playersFromFile("cache/players.json")

        if err != nil {
            return err
        }
    } else if err != nil {
        return err
    }

	fmt.Println(players)

    b.cache = newPlayerCache(players)
    b.addCommand(createSetRosterCommand())
    b.registerHandler("set-roster", setRosterHandler(b))

    b.addCommand(createGetGlobalRosterCommand())
    b.registerHandler("global-roster", getGlobalRosterCommandHandler(b))

    b.addCommand(createGetMyRosterCommand())
    b.registerHandler("my-roster", getMyRosterHandler(b))

    b.addCommand(createResetRosterCommand())
    b.registerHandler("reset-roster", resetRosterHandler(b))

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

