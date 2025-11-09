package bot

import (
    "cmp"
    "slices"
    typespkg "github.com/arjunmoola/nba-daily-fantasy-bot/internal/types"
)

func newGlobalRoster(players []typespkg.DiscordPlayer) map[string]*typespkg.DiscordPlayerRoster {
    rosters := make(map[string]*typespkg.DiscordPlayerRoster)

    for _, player := range players {
        if discordPlayer, ok := rosters[player.DiscordPlayerId]; !ok {
            rosters[player.DiscordPlayerId] = &typespkg.DiscordPlayerRoster{
                Nickname: player.Nickname,
                GuildId: player.GuildId,
                PlayerId: player.DiscordPlayerId,
                Date: player.Date,
                Players: []typespkg.NbaPlayerRoster{ player.GetRosterPlayer() },
            }
        } else {
            discordPlayer.Players = append(discordPlayer.Players, player.GetRosterPlayer())
        }
    }

    for _, player := range rosters {
        totalScore := float64(0)
        for _, p := range player.Players {
            totalScore += p.FantasyScore
        }
        player.TotalScore = totalScore

        slices.SortFunc(player.Players, func(a, b typespkg.NbaPlayerRoster) int {
            return -1 * cmp.Compare(a.FantasyScore, b.FantasyScore)
        })
    }

    return rosters
}

func constructLeaderBoard(rosters map[string]*typespkg.DiscordPlayerRoster) []*typespkg.DiscordPlayerRoster {
    var leaderboard []*typespkg.DiscordPlayerRoster

    for _, player := range rosters {
        leaderboard = append(leaderboard, player)
    }

    slices.SortFunc(leaderboard, func(a, b *typespkg.DiscordPlayerRoster) int {
        return -1*cmp.Compare(a.TotalScore, b.TotalScore)
    })

    return leaderboard
}
