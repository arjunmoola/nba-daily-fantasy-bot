package bot

import (
	"log/slog"
    "time"
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
    "strings"
    "slices"
    "cmp"
    "context"
    textpkg "github.com/arjunmoola/nba-daily-fantasy-bot/internal/text"
    typespkg "github.com/arjunmoola/nba-daily-fantasy-bot/internal/types"
	"github.com/arjunmoola/nba-daily-fantasy-bot/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func interactionAuthor(i *discordgo.Interaction) *discordgo.User {
    if i.Member != nil {
        return i.Member.User
    }
    return i.User
}


func setRosterHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        switch i.Type{
        case discordgo.InteractionApplicationCommand:
            if err := handleSetRosterInteractionApplicationCommand(bot, s, i); err != nil {
                log.Println(err)
                s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
                return
            }
        case discordgo.InteractionApplicationCommandAutocomplete:
            if err := handleSetRosterInteractionAutocomplete(bot, s, i); err != nil {
                log.Panic(err)
            }
        }
    }
}


func resetRosterHandler(bot *NbaFantasyBot) func (s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        author := interactionAuthor(i.Interaction)
        guildId := i.Interaction.GuildID
        discordPlayerId := author.ID

        date := time.Now().Format(time.DateOnly)

        discordPlayers, err := bot.client.GetMyRosterGuild(context.Background(), guildId, discordPlayerId, date)

        if err != nil {
            log.Println(err)
            errInteractionRespond(s, i, "unable to get my roster. http error")
            return
        }

        if len(discordPlayers) == 0 {
            err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: "Your roster is empty",
                    Flags: discordgo.MessageFlagsEphemeral,
                },
            })

            if err != nil {
                log.Println(err)
                return
            }
        }

        for _, player := range discordPlayers {
            rosterPlayer := player.GetRosterPlayer()

            payload := typespkg.MyRosterDeletePayload{
                NbaPlayerUID: player.NbaPlayerUID,
                NbaPlayerID: player.NbaPlayerId,
                Nickname: player.Nickname,
                Name: player.Name,
                DollarValue: player.DollarValue,
                FantasyScore: rosterPlayer.FantasyScore,
                DiscordPlayerID: player.DiscordPlayerId,
                GuildID: player.GuildId,
                Position: player.Position,
                Date: date,
            }

            fmt.Println(payload)

            if err := bot.client.DeleteMyRoster(context.Background(), payload); err != nil {
                log.Println(err)
                errInteractionRespond(s, i, fmt.Sprintf("unable to delete player %s", player.Name))
                continue
            }
        }

        err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Roster has been reset",
                Flags: discordgo.MessageFlagsEphemeral,
            },
        })

        if err != nil {
            log.Println(err)
        }
    }
}

func GetMyRosterHandler(logger *slog.Logger, pool *pgxpool.Pool) ContextHandler {
    return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
        author := interactionAuthor(i.Interaction)
        //guildId := i.Interaction.GuildID
        discordPlayerId := author.ID

        //date := time.Now().Format(time.DateOnly)

		var todaysRoster []database.GetTodaysRosterByDiscordIdRow

		err := pool.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
			params := database.GetTodaysRosterByDiscordIdParams{
				Discordid: discordPlayerId,
				Date: pgtype.Date{
					Time: time.Now(),
					InfinityModifier: pgtype.Finite,
					Valid: true,
				},
			}

			results, err := database.New(conn).GetTodaysRosterByDiscordId(ctx, params)

			if err != nil {
				return err
			}

			todaysRoster = results

			return nil
		})

		if err != nil {
			log.Println(err)
			errInteractionRespond(s, i, "unable to get my roster. http error")
			return
		}

        builder := strings.Builder{}

        totalValue := 0

		for _, player := range todaysRoster {
			totalValue += int(player.DollarValue.Int32)
			fmt.Fprintf(&builder, "%s %s %d\n", player.Name, player.Position, int(player.DollarValue.Int32))
		}

        remainingValue := 100 - totalValue

        builder.WriteString(fmt.Sprintf("remaining value: %d\n", remainingValue))

        err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: builder.String(),
                Flags: discordgo.MessageFlagsEphemeral,
            },
        })

        if err != nil {
            log.Println(err)
            return
        }
    }
}

func getMyRosterHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        author := interactionAuthor(i.Interaction)
        guildId := i.Interaction.GuildID
        discordPlayerId := author.ID

        date := time.Now().Format(time.DateOnly)

        discordPlayers, err := bot.client.GetMyRosterGuild(context.Background(), guildId, discordPlayerId, date)

        if err != nil {
            log.Println(err)
            errInteractionRespond(s, i, "unable to get my roster. http error")
            return
        }

        builder := strings.Builder{}

        totalValue := 0

        for _, player := range discordPlayers {
            rosterPlayer := player.GetRosterPlayer()

            totalValue += rosterPlayer.DollarValue
            builder.WriteString(fmt.Sprintf("%s %s %d\n", rosterPlayer.Name, rosterPlayer.Position, rosterPlayer.DollarValue))
        }

        remainingValue := 100 - totalValue

        builder.WriteString(fmt.Sprintf("remaining value: %d\n", remainingValue))

        err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: builder.String(),
                Flags: discordgo.MessageFlagsEphemeral,
            },
        })

        if err != nil {
            log.Println(err)
            return
        }
    }
}

func handleSetRosterInteractionApplicationCommand(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    builder := strings.Builder{}

    pos := []string{ "PG", "C", "SF", "PF", "SG" }

    author := interactionAuthor(i.Interaction)
    guildId := i.Interaction.GuildID
    nickName := author.Username
    discordPlayerId := author.ID
    date := time.Now().Format(time.DateOnly)

    existingPlayers, _ := bot.client.GetMyRosterGuild(context.Background(), guildId, discordPlayerId, date)

    var totalSum int64

    for j, option := range data.Options {
        player, ok := bot.cache.getPlayer(option.Value.(string))


        if !ok {
            continue
        }

        if exits := containsPlayer(existingPlayers, player.ID); exits {
            continue
        }

        totalSum += player.DollarValue

        if totalSum > 100 {
            return s.InteractionRespond(i.Interaction, errResponseInteraction("exceeded maximum budget"))
        }

        payload := typespkg.MyRosterPayload{
            PlayerId: player.UID,
            Position: player.Position,
            GuildId: i.Interaction.GuildID,
            Nickname: nickName,
            DiscordPlayerId: discordPlayerId,
            Date: date,
        }

        err := bot.client.SetMyRoster(context.Background(), payload)

        if err != nil {
            return errInteractionRespond(s, i, "http post error")
        }

        builder.WriteString(fmt.Sprintf("You picked %q for %s\n", player.Name, pos[j]))
    }

    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: builder.String(),
            Flags: discordgo.MessageFlagsEphemeral,
        },
    })

    return err
}

func handleSetRosterInteractionAutocomplete(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    var choices []*discordgo.ApplicationCommandOptionChoice

    fmt.Println(data.Options)

    switch {
    case data.Options[0].Focused:
        players := bot.cache.getPlayersByPos("PG")
        choices = createPlayerChoices(players)
        if data.Options[0].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(),opt.Value)
            }
            scores := getClosestPlayers(data.Options[0].StringValue(), players)
            choices = createChoicesFromScores(scores)
            fmt.Println(choices)
        }
    case data.Options[1].Focused:
        players := bot.cache.getPlayersByPos("C")
        choices = createPlayerChoices(players)
        if data.Options[1].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(), opt.Value)
            }
            scores := getClosestPlayers(data.Options[1].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[2].Focused:
        players := bot.cache.getPlayersByPos("SF")
        choices = createPlayerChoices(players)
        if data.Options[2].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(), opt.Value)
            }
            scores := getClosestPlayers(data.Options[2].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[3].Focused:
        players := bot.cache.getPlayersByPos("PF")
        choices = createPlayerChoices(players)
        if data.Options[3].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(), opt.Value)
            }
            scores := getClosestPlayers(data.Options[3].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[4].Focused:
        players := bot.cache.getPlayersByPos("SG")
        choices = createPlayerChoices(players)
        if data.Options[4].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(), opt.Value)
            }
            scores := getClosestPlayers(data.Options[4].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    }


    return autocompleteInteractionRespond(s, i, choices)

}

func autocompleteInteractionRespond(s *discordgo.Session, i *discordgo.InteractionCreate, choices []*discordgo.ApplicationCommandOptionChoice) error {
    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionApplicationCommandAutocompleteResult,
        Data: &discordgo.InteractionResponseData{
            Choices: choices,
        },
    })
}

type playerScore struct {
    player string
    score int
    dollarValue int64
    id int
}

func getClosestPlayers(src string, players []typespkg.NbaPlayer) []playerScore {
    scores := make([]playerScore, len(players))

    for i, player := range players {
        s := textpkg.ComputeSmithWatermanFunc(strings.ToLower(src), strings.ToLower(player.Name), textpkg.DefaultScorer)

        scores[i] = playerScore{
            player: player.Name,
            score: s,
            dollarValue: player.DollarValue,
            id: player.ID,
        }
    }

    scores = filterFunc(scores, greaterThanThreshold)

    slices.SortFunc(scores, func(a, b playerScore) int {
        return -1 * cmp.Compare(a.score, b.score)
    })

    return scores

}

func filterFunc[S []T,T any](s S, f func(T) bool) S {
    idx := 0
    for _, x := range s {
        if f(x) {
            s[idx] = x
            idx++
        }
    }
    s = s[:idx]
    return s
}

func containsFunc[S []T, T any](s S, y T, f func(T, T) bool) bool {
    for _, x := range s {
        if f(x, y) {
            return true
        }

    }
    return false
}

func containsPlayer(players []typespkg.DiscordPlayer, id int) bool {
    for _, player := range players {
        if id == player.NbaPlayerId {
            return true
        }
    }

    return false
}

func greaterThanThreshold(x playerScore) bool {
    if x.score >= 4 {
        return true
    } else {
        return false
    }
}

func createPlayerChoicesByPos(bot *NbaFantasyBot, pos string) []*discordgo.ApplicationCommandOptionChoice {
    players := bot.cache.getPlayersByPos(pos)
    return createPlayerChoices(players)
}

func createPlayerChoices(players []typespkg.NbaPlayer) []*discordgo.ApplicationCommandOptionChoice {
    choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, 25)

    if len(players) > 25 {
        players = players[:25]
    }

    for _, player := range players {
        choice := &discordgo.ApplicationCommandOptionChoice{
            Name: fmt.Sprintf("%s %d", player.Name, player.DollarValue),
            Value: fmt.Sprintf("%d", player.ID),
        }

        choices = append(choices, choice)
    }

    return choices
}

func createChoicesFromScores(scores []playerScore) []*discordgo.ApplicationCommandOptionChoice {
    choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, 25)

    if len(scores) > 25 {
        scores = scores[:25]
    }

    for _, score := range scores {
        choice := &discordgo.ApplicationCommandOptionChoice {
            Name: fmt.Sprintf("%s %d", score.player, score.dollarValue),
            Value: fmt.Sprintf("%d", score.id),
        }

        choices = append(choices, choice)
    }

    return choices
}

func getGlobalRosterCommandHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {

        date := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)

		fmt.Println("date: ", date)

        discordPlayers, err := bot.client.Roster.GetGlobalRoster(context.Background(), date)

        if err != nil {
            log.Println(err)
            s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
            return
        }

        user := interactionAuthor(i.Interaction)

        fmt.Println(user.ID, i.GuildID)

        leaderBoard := constructLeaderBoard(newGlobalRoster(discordPlayers))

        builder := strings.Builder{}

        fmt.Fprintln(&builder, "### Global Leaderboard\n")


        for _, player := range leaderBoard {
            builder.WriteString(fmt.Sprintf("**%s** %.2f\n", player.Nickname, player.TotalScore))

            for _, nbaPlayer := range player.Players {
                builder.WriteString(fmt.Sprintf("\t%s %s %.2f\n", nbaPlayer.Position, nbaPlayer.Name, nbaPlayer.FantasyScore))
            }
        }

        err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: builder.String(),
            },
        })

        if err != nil {
            log.Println(err)
        }
    }
}

func errResponseInteraction(msg string) *discordgo.InteractionResponse {
    return &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: msg,
            Flags: discordgo.MessageFlagsEphemeral,
        },
    }
}

func errInteractionRespond(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
    return s.InteractionRespond(i.Interaction, errResponseInteraction(msg))
}
