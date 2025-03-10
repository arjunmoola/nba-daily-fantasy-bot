package main

import (
    "io"
    "bytes"
    "fmt"
    "os"
    "log"
    "net/http"
    "context"
    "encoding/json"
    "errors"
    "net/url"
)

var ErrRosterLocked = errors.New("Roster is Locked")

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
        return nil, ErrRosterLocked
    }

    if err := savePlayers(players, "cache/players.json"); err != nil {
        return nil, err
    }

    return players, nil
}

func (c *NbaFantasyClient) getLockTime(ctx context.Context) (*LockTime, error) {
    url := fmt.Sprintf("%s/api/activity/lock-time", c.baseUrl)
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

    if err != nil {
        return nil, err
    }

    resp, err := c.client.Do(req)

    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    var lock LockTime

    if err := json.NewDecoder(resp.Body).Decode(&lock); err != nil {
        return nil, err
    }

    return &lock, nil
}

func (c *NbaFantasyClient) getGlobalRoster(ctx context.Context, date string) ([]DiscordPlayer, error) {
    values := url.Values{}
    values.Add("date", date)

    url := fmt.Sprintf("%s/api/activity/rosters/global?%s", c.baseUrl, values.Encode())

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

    if err != nil {
        return nil, err
    }

    resp, err := c.client.Do(req)

    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    var players []DiscordPlayer 

    if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
        return nil, err
    }

    fmt.Println(players)

    return players, nil
}

func (c *NbaFantasyClient) getWeeklyRoster(ctx context.Context, guildId string, date string) ([]DiscordPlayer, error) {
    values := url.Values{}
    values.Add("date", date)

    url := fmt.Sprintf("%s/api/activity/rosters/%s/weekly?%s", c.baseUrl, guildId, values.Encode())

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

    if err != nil {
        return nil, err
    }

    resp, err := c.client.Do(req)

    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    var players []DiscordPlayer 

    if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
        return nil, err
    }

    fmt.Println(players)

    return players, nil
}

func (c *NbaFantasyClient) setMyRoster(ctx context.Context, payload myRosterPayload) error {
    url := fmt.Sprintf("%s/api/activity/my-roster", c.baseUrl)

    data, err := json.Marshal(payload)

    if err != nil {
        return err
    }

    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))

    if err != nil {
        return err
    }

    req.Header.Add("Content-Type", "application/json")


    resp, err := c.client.Do(req)

    if err != nil {
        return err
    }

    defer resp.Body.Close()

    log.Println(resp.Status, resp.StatusCode)

    data, err = io.ReadAll(resp.Body)

    if err != nil {
        return err
    }

    log.Println(string(data))


    return nil
}

func (c *NbaFantasyClient) deleteMyRoster(ctx context.Context, payload myRosterDeletePayload) error {
    data, err := json.Marshal(payload)

    if err != nil {
        return err
    }

    url := fmt.Sprintf("%s/api/activity/my-roster", c.baseUrl)

    req, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewBuffer(data))

    if err != nil {
        return err
    }

    req.Header.Add("Content-Type", "application/json")

    resp, err := c.client.Do(req)

    if err != nil {
        return err
    }

    defer resp.Body.Close()

    fmt.Println(resp.Status, resp.StatusCode)

    data, err = io.ReadAll(resp.Body)

    if err != nil {
        return err
    }

    fmt.Println(string(data))

    return nil
}

func (c *NbaFantasyClient) getMyRosterGuild(ctx context.Context, guildId string, discordPlayerId string, date string) ([]DiscordPlayer, error) {
    values := url.Values{}
    values.Add("date", date)
    url := fmt.Sprintf("%s/api/activity/my-roster/%s/%s?%s", c.baseUrl, guildId, discordPlayerId, values.Encode())

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

    if err != nil {
        return nil, err
    }

    resp, err := c.client.Do(req)

    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    var discordPlayers []DiscordPlayer

    if err := json.NewDecoder(resp.Body).Decode(&discordPlayers); err != nil {
        return nil, err
    }

    return discordPlayers, nil
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
