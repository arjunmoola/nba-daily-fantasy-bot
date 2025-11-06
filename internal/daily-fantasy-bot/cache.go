package dailyfantasybot

import (
    "cmp"
    "slices"
    "strconv"
    "sync"
    typespkg "github.com/arjunmoola/nba-daily-fantasy-bot/internal/types"
)

type PlayerCache struct {
    playersMx, positionsMx sync.RWMutex
    players map[int]typespkg.NbaPlayer
    positions map[string][]typespkg.NbaPlayer
}

func newPlayerCache(players []typespkg.NbaPlayer) *PlayerCache {
    c := PlayerCache{}
    c.players = make(map[int]typespkg.NbaPlayer)
    c.positions = make(map[string][]typespkg.NbaPlayer)

    for _, player := range players {
        c.players[player.ID] = player
        c.positions[player.Position] = append(c.positions[player.Position], player)
    }

    for pos := range c.positions {
        slices.SortFunc(c.positions[pos], func(a, b typespkg.NbaPlayer) int {
            return -1*cmp.Compare(a.DollarValue, b.DollarValue)
        })
    }

    return &c
}

func (c *PlayerCache) getPlayersByPos(pos string) []typespkg.NbaPlayer {
    c.positionsMx.RLock()
    defer c.positionsMx.RUnlock()

    players := c.positions[pos]

    return slices.Clone(players)
}

func (c *PlayerCache) getPlayer(id string) (typespkg.NbaPlayer, bool) {
    i, _ := strconv.Atoi(id)

    c.playersMx.RLock()
    defer c.playersMx.RUnlock()

    player, ok := c.players[i]

    return player, ok
}
