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
                Name: "pg",
                Description: "Choose PG",
                Type: discordgo.ApplicationCommandOptionString,
                Required: true,
                Autocomplete: true,
            },
            {
                Name: "c",
                Description: "Choose Center",
                Type: discordgo.ApplicationCommandOptionString,
                Required: true,
                Autocomplete: true,
            },
        },
    }   
}
