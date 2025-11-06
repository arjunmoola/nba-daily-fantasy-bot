package client

import (
	//"maps"
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
    typespkg "github.com/arjunmoola/nba-daily-fantasy-bot/internal/types"
)

var ErrRosterLocked = errors.New("Roster is Locked")

type NbaFantasyClient struct {
    baseUrl *url.URL
    client *http.Client

	Activity *Activity
	Roster *Roster
}

type Activity struct {
	client *NbaFantasyClient
}

func (a *Activity) newURLConstructor() *urlConstructor {
	cons := newUrlConstructor(a.client.baseUrl)
	cons.join("activity")
	return cons
}

type Roster struct {
	client *NbaFantasyClient
}

func (r *Roster) newUrlConstructor() *urlConstructor {
	cons := newUrlConstructor(r.client.baseUrl)
	cons.join("activity", "roster")
	return cons
}

type urlConstructor struct {
	base *url.URL
	method string
	body io.Reader
	values url.Values
}

func newUrlConstructor(base *url.URL) *urlConstructor {
	return &urlConstructor{
		base: base,
		values: make(url.Values),
	}
}

func (c *urlConstructor) join(s ...string) {
	c.base = c.base.JoinPath(s...)
}

func (c *urlConstructor) set(key, value string) {
	c.values.Set(key, value)
}

func (c *urlConstructor) setMethod(method string) {
	c.method = method
}

func (c *urlConstructor) String() string {
	c.base.RawQuery = c.values.Encode()
	return c.base.String()
}

func (c *urlConstructor) NewRequestWithContext(ctx context.Context) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, c.method, c.String(), c.body)
}

func doRequest(ctx context.Context, client *http.Client, cons *urlConstructor, v any) error {
	req, err := cons.NewRequestWithContext(ctx)

	if err != nil {
		return err
	}

	return fetchRequest(client, req, v)
}

func fetchRequest(client *http.Client, req *http.Request, v any) error {
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func NewNbaFantasyClient() (*NbaFantasyClient, error) {
	serverURL := os.Getenv("URL")
	baseUrl, err := url.Parse(serverURL)

	if err != nil {
		return nil, fmt.Errorf("server url incorrect. unable to parse url %v", err)
	}

	client := &NbaFantasyClient{
        baseUrl: baseUrl,
        client: &http.Client{},
    }

	activity := &Activity{
		client: client,
	}

	roster := &Roster{
		client: client,
	}

	client.Activity = activity
	client.Roster = roster



	return client, nil
}

func (a *Activity) GetTodaysPlayers(ctx context.Context) ([]typespkg.NbaPlayer, error) {
	cons := a.newURLConstructor()
	cons.join("todays-players")
	cons.setMethod(http.MethodGet)

	var players []typespkg.NbaPlayer

	if err := doRequest(ctx, a.client.client, cons, &players); err != nil {
		return nil, err
	}

	return players, nil
}

func (a *Activity) GetLockTime(ctx context.Context) (*typespkg.LockTime, error) {
	cons := a.newURLConstructor()
	cons.join("lock-time")
	cons.setMethod(http.MethodGet)

	var lockTime typespkg.LockTime

	if err := doRequest(ctx, a.client.client, cons, &lockTime); err != nil {
		return nil, err
	}

	return &lockTime, nil
}

func (r *Roster) GetGlobalRoster(ctx context.Context, date string) ([]typespkg.DiscordPlayer, error) {
	cons := r.newUrlConstructor()
	cons.set("date", date)
	cons.setMethod(http.MethodGet)

	var players []typespkg.DiscordPlayer

	if err := doRequest(ctx, r.client.client, cons, &players); err != nil {
		return nil, err
	}

	return players, nil
}

func (r *Roster) GetWeeklyRoster(ctx context.Context, guidId string, date string) ([]typespkg.DiscordPlayer, error) {
	cons := r.newUrlConstructor()
	cons.join(guidId, "weekly")
	cons.set("date", date)
	
	var players []typespkg.DiscordPlayer

	if err := doRequest(ctx, r.client.client, cons, &players); err != nil {
		return nil, err
	}

	return players, nil
}

func (c *NbaFantasyClient) GetLockTime(ctx context.Context) (*typespkg.LockTime, error) {
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

    var lock typespkg.LockTime

    if err := json.NewDecoder(resp.Body).Decode(&lock); err != nil {
        return nil, err
    }

    return &lock, nil
}

func (c *NbaFantasyClient) GetGlobalRoster(ctx context.Context, date string) ([]typespkg.DiscordPlayer, error) {
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

    var players []typespkg.DiscordPlayer 

    if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
        return nil, err
    }

    fmt.Println(players)

    return players, nil
}

func (c *NbaFantasyClient) GetWeeklyRoster(ctx context.Context, guildId string, date string) ([]typespkg.DiscordPlayer, error) {
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

    var players []typespkg.DiscordPlayer 

    if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
        return nil, err
    }

    fmt.Println(players)

    return players, nil
}

func (c *NbaFantasyClient) SetMyRoster(ctx context.Context, payload typespkg.MyRosterPayload) error {
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

func (c *NbaFantasyClient) DeleteMyRoster(ctx context.Context, payload typespkg.MyRosterDeletePayload) error {
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

func (c *NbaFantasyClient) GetMyRosterGuild(ctx context.Context, guildId string, discordPlayerId string, date string) ([]typespkg.DiscordPlayer, error) {
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

    var discordPlayers []typespkg.DiscordPlayer

    if err := json.NewDecoder(resp.Body).Decode(&discordPlayers); err != nil {
        return nil, err
    }

    return discordPlayers, nil
}

func savePlayers(players []typespkg.NbaPlayer, path string) error {
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
