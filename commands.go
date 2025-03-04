package main

import (
    "github.com/bwmarrin/discordgo"
)

func createSetRosterCommand() *discordgo.ApplicationCommand {
    return &discordgo.ApplicationCommand{
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
    }   
}
