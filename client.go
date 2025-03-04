package main

import (
    "fmt"
    "os"
    "net/http"
    "context"
    "encoding/json"
)

type NbaFantasyClient struct {
    baseUrl string
    client *http.Client
}

func newNbaFantasyClient() *NbaFantasyClient {
    url := os.Getenv("URL")
    return &NbaFantasyClient{
        baseUrl: url,
        client: &http.Client{},
    }
}

func (c *NbaFantasyClient) getTodaysPlayers(ctx context.Context) ([]NbaPlayer, error) {
    url := fmt.Sprintf("%s/api/activity/todays-players", c.baseUrl)
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := c.client.Do(req)

    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    var players []NbaPlayer

    if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
        return nil, err
    }

    if len(players) == 0 {
        return nil, fmt.Errorf("roster is locked")
    }

    if err := savePlayers(players, "cache/players.json"); err != nil {
        return nil, err
    }

    return players, nil
}

func savePlayers(players []NbaPlayer, path string) error {
    file, err := os.Create(path)

    if err != nil {
        return err
    }

    defer file.Close()

    data, err := json.Marshal(players)

    if err != nil {
        return err
    }

    if _, err := file.Write(data); err != nil {
        return err
    }

    return nil
}
