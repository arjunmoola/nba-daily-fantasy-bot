package main

type NbaPlayer struct {
    UID string `json:"nba_player_uid"`
    ID int `json:"nba_player_id"`
    Name string `json:"player_name"`
    Date string `json:"date"`
    Position string `json:"position"`
    AgainstTeam int64 `json:"against_team"`
    AgainstTeamName string `json:"against_team_name"`
    DollarValue int64 `json:"dollar_value"`
    FantasyScore *int64 `json:"fantasy_score"`
    TeamId int64 `json:"team_id"`
    TeamName string `json:"team_name"`
    Status string `json:"status"`
    JerseyNum string `json:"jersey_num"`
    AvgStl string `json:"avg_stl"`
    AvgTov string `json:"avg_tov"`
    AvgReb string `json:"avg_pts"`
    AvgAst string `json:"avg_ast"`
    AvgBlk string `json:"avg_blk"`
}

type DiscordPlayer struct {
    NbaPlayerUID string `json:"nbaPlayerUid"`
    NbaPlayerId int `json:"nbaPlayerId"`
    Nickname string `json:"nickname"`
    Name string `json:"name"`
    DollarValue int `json:"dollarValue"`
    FantasyScore *int `json:"fantasyScore"`
    GuildId string `json:"guildId"`
    DiscordPlayerId string `json:"discordPlayerId"`
    Position string `json:"position"`
}

type LockTime struct {
    Date string `json:"date"`
    Time string `json:"lockTime"`
}
