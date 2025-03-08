package main

import (
    //"time"
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
    "strings"
    "slices"
    "cmp"
    "context"
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

func setPGHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        switch i.Type{
        case discordgo.InteractionApplicationCommand:
            if err := handleSetPGInteractionApplicationCommand(bot, s, i); err != nil {
                log.Println(err)
                s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
                return
            }
        case discordgo.InteractionApplicationCommandAutocomplete:
        if err := handleSetPGInteractionAutocomplete(bot, s, i); err != nil {
                log.Println(err)
                s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
                return
            }
        }
    }
}

func handleSetPGInteractionApplicationCommand(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("You picked %q", data.Options[0].StringValue()),
        },
    })
    return err
}

func handleSetPGInteractionAutocomplete(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    var choices []*discordgo.ApplicationCommandOptionChoice

    players := bot.cache.getPlayersByPos("PG")
    choices = createPlayerChoices(players)

    if data.Options[0].StringValue() != "" {
        scores := getClosestPlayers(data.Options[0].StringValue(), players)
        choices = createChoicesFromScores(scores)
    }

    return autocompleteInteractionRespond(s, i, choices)
}

func handleSetRosterInteractionApplicationCommand(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    builder := strings.Builder{}

    pos := []string{ "PG", "C", "SF", "PF", "SG" }

    var totalSum int64

    for j, option := range data.Options {
        player, ok := bot.cache.getPlayer(option.Value.(string))

        if !ok {
            continue
        }

        totalSum += player.DollarValue

        if totalSum > 100 {
            return s.InteractionRespond(i.Interaction, errResponseInteraction("exceeded maximum budget"))
        }

        builder.WriteString(fmt.Sprintf("You picked %q for %s\n", option.StringValue(), pos[j]))
    }

    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: builder.String(),
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

func getClosestPlayers(src string, players []NbaPlayer) []playerScore {
    scores := make([]playerScore, len(players))

    for i, player := range players {
        s := computeSmithWatermanFunc(strings.ToLower(src), strings.ToLower(player.Name), defaultScorer)

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

func createPlayerChoices(players []NbaPlayer) []*discordgo.ApplicationCommandOptionChoice {
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
        date := "2025-03-06"
        discordPlayers, err := bot.client.getGlobalRoster(context.Background(), date)

        if err != nil {
            log.Println(err)
            s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
            return
        }

        user := interactionAuthor(i.Interaction)

        fmt.Println(user.ID, i.GuildID)

        rosterMap := newGlobalRoster(discordPlayers)

        str := ""


        for _, discordPlayer := range rosterMap {
            playerName := discordPlayer.Nickname
            roster := ""
            playerSum := 0.0
            for _, nbaPlayer := range discordPlayer.players {
                score := nbaPlayer.FantasyScore
                fmt.Println(score)
                if score != nil {
                    roster += fmt.Sprintf("\t%s %s %.2f\n", nbaPlayer.Name, nbaPlayer.Position, *score)
                    playerSum += *score
                } else {
                    roster += fmt.Sprintf("\t%s %s\n", nbaPlayer.Name, nbaPlayer.Position)
                }
            }

            str += fmt.Sprintf("**%s** **%.2f**\n", playerName, playerSum) + roster
        }

        err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: str,
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
        },
    }
}
