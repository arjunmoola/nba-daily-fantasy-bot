package main

import (
    "sync"
    "cmp"
    "slices"
)

type PlayerCache struct {
    playersMx, positionsMx sync.RWMutex
    players map[int]NbaPlayer
    positions map[string][]NbaPlayer
}

func newPlayerCache(players []NbaPlayer) *PlayerCache {
    c := PlayerCache{}
    c.players = make(map[int]NbaPlayer)
    c.positions = make(map[string][]NbaPlayer)

    for _, player := range players {
        c.players[player.ID] = player
        c.positions[player.Position] = append(c.positions[player.Position], player)
    }

    for _, players := range c.positions {
        slices.SortFunc(players, func(a, b NbaPlayer) int {
            return -1*cmp.Compare(a.DollarValue, b.DollarValue)
        })
    }

    return &c
}

func (c *PlayerCache) getPlayersByPos(pos string) []NbaPlayer {
    c.positionsMx.RLock()
    defer c.positionsMx.RUnlock()

    players := c.positions[pos]

    return slices.Clone(players)
}
