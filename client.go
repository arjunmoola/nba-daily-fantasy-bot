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

    return players, nil
}

