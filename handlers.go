package main

import (
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
    "strings"
    "slices"
    "cmp"
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

    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("You picked %q", data.Options[0].StringValue()),
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
            fmt.Println(scores)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[1].Focused:
        players := bot.cache.getPlayersByPos("C")

        log.Println(players)

        choices = createPlayerChoices(players)

        if data.Options[1].StringValue() != "" {
            scores := getClosestPlayers(data.Options[1].StringValue(), players)
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
        s := computeSmithWaterman(strings.ToLower(src), strings.ToLower(player.Name))

        scores[i] = playerScore{
            player: player.Name,
            score: s,
            dollarValue: player.DollarValue,
        }
    }

    c := 0

    for _, s := range scores {
        if s.score != 0 {
            scores[c] = s
            c++
        }
    }

    scores = scores[:c]

    if len(scores) > 25 {
        scores = scores[:25]
    }

    slices.SortFunc(scores, func(a, b playerScore) int {
        return -1 * cmp.Compare(a.score, b.score)
    })

    return scores

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
