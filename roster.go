package main

import (
    "sync"
)

type DiscordPlayerRoster struct {
    Nickname string
    PlayerId string
    GuildId string
    Date string
    players []nbaPlayerRoster
}

type nbaPlayerRoster struct {
    Uid string
    Id int
    Name string
    DollarValue int
    FantasyScore *float64
    Position string

}

type GlobalRoster struct {
    rosterMx sync.RWMutex
    rosters map[string]*DiscordPlayerRoster
}

func newGlobalRoster(players []DiscordPlayer) map[string]*DiscordPlayerRoster {
    rosters := make(map[string]*DiscordPlayerRoster)

    for _, player := range players {
        if discordPlayer, ok := rosters[player.DiscordPlayerId]; !ok {
            rosters[player.DiscordPlayerId] = &DiscordPlayerRoster{
                Nickname: player.Nickname,
                GuildId: player.GuildId,
                PlayerId: player.DiscordPlayerId,
                Date: player.Date,
                players: []nbaPlayerRoster{ player.getRosterPlayer() },
            }
        } else {
            discordPlayer.players = append(discordPlayer.players, player.getRosterPlayer())
        }
    }

    return rosters

}
