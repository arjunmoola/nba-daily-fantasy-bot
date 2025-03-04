package main

import (
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
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
            Content: fmt.Sprintf("You picked %q autocompletion", data.Options[0].StringValue()),
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
            choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
                Name: data.Options[0].StringValue(),
                Value: "choice_custom",
            })
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
