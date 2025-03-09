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
            {
                Name: "sf",
                Description: "Choose Small Forward",
                Type: discordgo.ApplicationCommandOptionString,
                Required: true,
                Autocomplete: true,
            },
            {
                Name: "pf",
                Description: "Choose Power Forward",
                Type: discordgo.ApplicationCommandOptionString,
                Required: true,
                Autocomplete: true,
            },
            {
                Name: "sg",
                Description: "Choose shooting guard",
                Type: discordgo.ApplicationCommandOptionString,
                Required: true,
                Autocomplete: true,
            },
        },
    }   
}

func createSetPGCommand() *discordgo.ApplicationCommand {
    return &discordgo.ApplicationCommand{
        Name: "set-pg",
        Description: "Set PG",
        Type: discordgo.ChatApplicationCommand,
        Options: []*discordgo.ApplicationCommandOption{
            {
                Name: "pg",
                Description: "Choose PG",
                Type: discordgo.ApplicationCommandOptionString,
                Required: true,
                Autocomplete: true,
            },
        },
    }
}


func createGetGlobalRosterCommand() *discordgo.ApplicationCommand {
    return &discordgo.ApplicationCommand{
        Name: "global-roster",
        Description: "Get global roster",
        Type: discordgo.ChatApplicationCommand,
    }
}

func createGetPGCommand() *discordgo.ApplicationCommand {
    return &discordgo.ApplicationCommand{
        Name: "get-pg",
        Description: "Get all playing point guards",
        Type: discordgo.ChatApplicationCommand,
    }
}

func createGetMyRosterCommand() *discordgo.ApplicationCommand {
    return &discordgo.ApplicationCommand{
        Name: "my-roster",
        Description: "Get my roster today",
        Type: discordgo.ChatApplicationCommand,
    }
}
