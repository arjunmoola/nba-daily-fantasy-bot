package bot

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
	"log/slog"
    clientpkg "github.com/arjunmoola/nba-daily-fantasy-bot/internal/client"
    typespkg "github.com/arjunmoola/nba-daily-fantasy-bot/internal/types"
	"github.com/arjunmoola/nba-daily-fantasy-bot/internal/config"
	//db "github.com/arjunmoola/nba-daily-fantasy-bot/internal/daily-fantasy-bot-db"
	"github.com/arjunmoola/nba-daily-fantasy-bot/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"database/sql"
)

type DiscordEventHandler func(*discordgo.Session, any)

type ContextHandler func(context.Context, *discordgo.Session, *discordgo.InteractionCreate)
type InteractionHandler func(*discordgo.Session, *discordgo.InteractionCreate)
type ReadyHandler func(*discordgo.Session, *discordgo.Ready)
type Middleware func(InteractionHandler) InteractionHandler

type BotOption func(b *PicknRollBot)

func WithConfig(c *config.Config) BotOption {
	return func(b *PicknRollBot) {
		b.config = c
	}
}

func WithLogger(l *slog.Logger) BotOption {
	return func(b *PicknRollBot) {
		b.logger = l
	}
}

func WithPool(p *pgxpool.Pool) BotOption {
	return func(b *PicknRollBot) {
		b.pool = p
	}
}

func WithCache(c *sql.DB) BotOption {
	return func(b *PicknRollBot) {
		b.cache = c
	}
}

type PicknRollBot struct {
	pool *pgxpool.Pool 
	cache *sql.DB
	logger *slog.Logger
	session *discordgo.Session
	config *config.Config

	lockTime time.Time
	todaysPlayers []database.NbaPlayer

    cmds []*discordgo.ApplicationCommand
    cmdHandlers map[string]InteractionHandler

	registry *Registry
}

func New(config *config.Config, opts ...BotOption) *PicknRollBot {
	bot := new(PicknRollBot)
	bot.config = config

	for _, opt := range opts {
		opt(bot)
	}

	return bot
}

func (b *PicknRollBot) Init(ctx context.Context) error {
	if b.config == nil {
		return fmt.Errorf("config not set")
	}

	if b.pool == nil {
		pool, err := pgxpool.New(ctx, b.config.DatabaseURL)

		if err != nil {
			return err
		}

		if err := pool.Ping(ctx); err != nil {
			return err
		}

		b.pool = pool
	}

	if b.session == nil {
		session, err := discordgo.New("Bot " + b.config.Token)

		if err != nil {
			return err
		}

		b.session = session
	}

	if b.logger == nil {
		b.logger = slog.Default()
		b.logger.Info("logger set")
	}

	if err := getLockTime(ctx, b.pool, &b.lockTime); err != nil {
		return err
	}

	if err := getTodaysPlayers(ctx, b.pool, &b.todaysPlayers); err != nil {
		return err
	}

	registry := NewRegistry(
		NewCommand(createGetMyRosterCommand(), GetMyRosterHandler(b.logger, b.pool)),
	)

	b.registry = registry

    // b.addCommand(createSetRosterCommand())
    // b.registerHandler("set-roster", setRosterHandler(b))
    //
    // b.addCommand(createGetGlobalRosterCommand())
    // b.registerHandler("global-roster", getGlobalRosterCommandHandler(b))
    //
    // b.addCommand(createGetMyRosterCommand())
    // b.registerHandler("my-roster", getMyRosterHandler(b))
    //
    // b.addCommand(createResetRosterCommand())
    // b.registerHandler("reset-roster", resetRosterHandler(b))

    b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		b.logger.Info("bot is running")
    })

	b.session.AddHandler(registry.DispatchWithLogger(b.logger))

	//fmt.Println(b.lockTime.Date())

	return nil
}

func getLockTime(ctx context.Context, pool *pgxpool.Pool, lockTime *time.Time) error {
	conn, err := pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	param := pgtype.Date{
		Time: time.Now(),
		InfinityModifier: pgtype.Finite,
		Valid: true,
	}

	result, err := database.New(conn).IsLocked(ctx, param)

	if err != nil {
		return err
	}

	*lockTime = result.LockTime.Time

	return nil
	
}

func getTodaysPlayers(ctx context.Context, pool *pgxpool.Pool, players *[]database.NbaPlayer) error {
	conn, err := pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	result, err := database.New(conn).GetAllTodaysNbaPlayers(ctx)

	if err != nil {
		return err
	}

	*players = result

	return nil
}

func (b *PicknRollBot) Run(ctx context.Context) error {
	if err := b.session.Open(); err != nil {
		return err
	}

	defer b.session.Close()
	defer b.pool.Close()

	_, err := b.registry.BulkOverWriteSession(b.session)

    if err != nil {
        return err
    }

    stop := make(chan os.Signal, 1)

    signal.Notify(stop, os.Interrupt)
    <-stop

    log.Println("Gracefully shutting down")

    return nil
}

type NbaFantasyBot struct {
    client *clientpkg.NbaFantasyClient
    session *discordgo.Session
    cache *PlayerCache
	//db *db.DB

	logger *slog.Logger
    cmds []*discordgo.ApplicationCommand
    cmdHandlers map[string]InteractionHandler
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
        cmdHandlers: make(map[string]InteractionHandler),
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
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))

	if err != nil {
		return fmt.Errorf("unable to parse pgxpool.Config")
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		return fmt.Errorf("unable to set up pgx pool %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database %v", err)
	}

	//b.pool = pool

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

type Registry struct {
	commands map[string]Command
}

func NewRegistry(cmds ...Command) *Registry{
	reg := new(Registry)
	reg.commands = make(map[string]Command)

	for _, cmd := range cmds {
		reg.commands[cmd.AppCmd.Name] = cmd
	}

	return reg
}

func (r *Registry) Definitions() []*discordgo.ApplicationCommand {
	var defs []*discordgo.ApplicationCommand

	for _, cmd := range r.commands {
		defs = append(defs, cmd.AppCmd)
	}

	return defs
}

func (r *Registry) BulkOverWriteSession(s *discordgo.Session) ([]*discordgo.ApplicationCommand, error) {
	created, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", r.Definitions())

	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Registry) Dispatch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Name

	cmd, ok := r.commands[name]

	if !ok {
		return
	}

	cmd.Handle(context.Background(), s, i)
}

func (r *Registry) DispatchWithLogger(l *slog.Logger) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		l = l.With("component", "dispatcher")

		switch i.Type {
		case discordgo.InteractionPing:
			l.Info("received interaction ping", "interaction-type", i.Type.String())
		case discordgo.InteractionApplicationCommand:
			author := interactionAuthor(i.Interaction)
			name := i.ApplicationCommandData().Name
			appId := i.Interaction.AppID
			channelId := i.Interaction.ChannelID
			guildId := i.Interaction.GuildID

			if cmd, ok := r.commands[name]; ok {
				l.Info("handling interaction create request",
					"appId", appId,
					"guildId", guildId,
					"channelId", channelId,
					"userId", author.ID,
					"author", author,
					"interaction-type", i.Type.String(),
					"command", name,
				)
				cmd.Handle(context.Background(), s, i)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
		case discordgo.InteractionMessageComponent:
		case discordgo.InteractionModalSubmit:
		}
	}
}

type Command struct {
	Name string
	AppCmd *discordgo.ApplicationCommand
	Handle ContextHandler
}

func NewCommand(cmd *discordgo.ApplicationCommand, handler ContextHandler) Command {
	return Command{
		AppCmd: cmd,
		Handle: handler,
	}
}

func handleWithContext(ctx context.Context, handle InteractionHandler) InteractionHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handle(s, i)
	}
}

