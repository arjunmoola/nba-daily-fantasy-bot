package main

import (
    "sync"
    "slices"
    "cmp"
)

type DiscordPlayerRoster struct {
    Nickname string
    PlayerId string
    GuildId string
    Date string
    players []nbaPlayerRoster
    TotalScore float64
}

type nbaPlayerRoster struct {
    Uid string
    Id int
    Name string
    DollarValue int
    FantasyScore float64
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

    for _, player := range rosters {
        player.setTotalScore()

        slices.SortFunc(player.players, func(a, b nbaPlayerRoster) int {
            return -1 * cmp.Compare(a.FantasyScore, b.FantasyScore)
        })
    }

    return rosters
}

func constructLeaderBoard(rosters map[string]*DiscordPlayerRoster) []*DiscordPlayerRoster {
    var leaderboard []*DiscordPlayerRoster

    for _, player := range rosters {
        leaderboard = append(leaderboard, player)
    }

    slices.SortFunc(leaderboard, func(a, b *DiscordPlayerRoster) int {
        return -1*cmp.Compare(a.TotalScore, b.TotalScore)
    })

    return leaderboard
}

func (p *DiscordPlayerRoster) setTotalScore() {
    totalScore := float64(0)
    for _, player := range p.players {
        totalScore += player.FantasyScore
    }
    p.TotalScore = totalScore
}
