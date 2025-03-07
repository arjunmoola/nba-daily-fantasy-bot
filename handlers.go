package main

import (
    "time"
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
    "strings"
    "slices"
    "cmp"
    "context"
)

func setRosterHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        switch i.Type{
        case discordgo.InteractionApplicationCommand:
            if err := handleSetRosterInteractionApplicationCommand(bot, s, i); err != nil {
                log.Panic(err)
            }
        case discordgo.InteractionApplicationCommandAutocomplete:
            if err := handleSetRosterInteractionAutocomplete(bot, s, i); err != nil {
                log.Panic(err)
            }
        }
    }
}

func handleSetRosterInteractionApplicationCommand(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    builder := strings.Builder{}

    pos := []string{ "PG", "C", "SF", "PF", "SG" }

    for i, option := range data.Options {
        builder.WriteString(fmt.Sprintf("You picked %q for %s\n", option.StringValue(), pos[i]))
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

    switch {
    case data.Options[0].Focused:
        players := bot.cache.getPlayersByPos("PG")
        choices = createPlayerChoices(players)
        if data.Options[0].StringValue() != "" {
            fmt.Println(data.Options[0].StringValue())
            scores := getClosestPlayers(data.Options[0].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[1].Focused:
        players := bot.cache.getPlayersByPos("C")
        choices = createPlayerChoices(players)
        if data.Options[1].StringValue() != "" {
            scores := getClosestPlayers(data.Options[1].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[2].Focused:
        players := bot.cache.getPlayersByPos("SF")
        choices = createPlayerChoices(players)
        if data.Options[2].StringValue() != "" {
            scores := getClosestPlayers(data.Options[2].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[3].Focused:
        players := bot.cache.getPlayersByPos("PF")
        choices = createPlayerChoices(players)
        if data.Options[3].StringValue() != "" {
            scores := getClosestPlayers(data.Options[3].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[4].Focused:
        players := bot.cache.getPlayersByPos("SG")
        choices = createPlayerChoices(players)
        if data.Options[4].StringValue() != "" {
            scores := getClosestPlayers(data.Options[4].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    }

    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionApplicationCommandAutocompleteResult,
        Data: &discordgo.InteractionResponseData{
            Choices: choices,
        },
    })

    return err
}

type playerScore struct {
    player string
    score int
    dollarValue int64
}

func getClosestPlayers(src string, players []NbaPlayer) []playerScore {
    scores := make([]playerScore, len(players))

    for i, player := range players {
        s := computeSmithWatermanFunc(strings.ToLower(src), strings.ToLower(player.Name), defaultScorer)

        scores[i] = playerScore{
            player: player.Name,
            score: s,
            dollarValue: player.DollarValue,
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
            Value: player.Name,
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
            Value: score.player,
        }

        choices = append(choices, choice)
    }

    return choices
}

func getGlobalRosterCommandHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        date := time.Now().Format(time.DateOnly)
        discordPlayers, err := bot.client.getGlobalRoster(context.Background(), date)

        if err != nil {
            log.Println(err)
            s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
            return
        }

        rosterMap := newGlobalRoster(discordPlayers)

        builder := strings.Builder{}

        for _, discordPlayer := range rosterMap {
            builder.WriteString(discordPlayer.Nickname + "\n")
            for _, nbaPlayer := range discordPlayer.players {
                builder.WriteString(fmt.Sprintf("\t%s %s\n", nbaPlayer.Name, nbaPlayer.Position))
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
        },
    }
}
