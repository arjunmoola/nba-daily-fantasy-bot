-- name: SaveRosterChoice :exec
insert into daily_roster (discord_player_id, nba_player_uid, date, nickname, position, guild_id) values (sqlc.arg(discordplayerid), sqlc.arg(nbaplayeruid), sqlc.arg(date), sqlc.arg(nickname), sqlc.arg(position)::daily_roster_position, 'null') on conflict(discord_player_id, date, position) do update set nba_player_uid = sqlc.arg(nbaplayeruid), position = sqlc.arg(position)::daily_roster_position;

-- name: DeleteRosterPlayerByDateAndDiscordIdAndPlayerName :exec
delete from daily_roster dr where dr.date = sqlc.arg(date) and dr.discord_player_id = sqlc.arg(discordid) and dr.nba_player_uid = sqlc.arg(nbaplayeruid);

-- name: GetTodaysRosterByDiscordId :many
select dr.*, np.nba_player_uid, np.nba_player_id, np.name, np.dollar_value from daily_roster dr join nba_players np on np.nba_player_uid = dr.nba_player_uid where dr.discord_player_id = sqlc.arg(discordid) and dr.date = sqlc.arg(date);

-- name: GetTodaysRosterPrice :many
select np.dollar_value from daily_roster dr join nba_players np on np.nba_player_uid = dr.nba_player_uid where dr.discord_player_id = sqlc.arg(discordid) and dr.date = sqlc.arg(date) and dr.position <> sqlc.arg(position)::daily_roster_position;

-- name: GetTodaysGlobalLeaderboard :many
with roster_totals as ( select dr.discord_player_id, sum(np.fantasy_score) as total_score from daily_roster dr join nba_players np on np.nba_player_uid = dr.nba_player_uid where dr.date = sqlc.arg(date) group by dr.discord_player_id order by total_score desc limit 100 ) select dr.discord_player_id, np.nba_player_id, dr.nba_player_uid, dr.date, dr.nickname, dr.position as position, np.name, np.dollar_value, np.fantasy_score from daily_roster dr join nba_players np on np.nba_player_uid = dr.nba_player_uid join roster_totals rr on dr.discord_player_id = rr.discord_player_id where dr.date = sqlc.arg(date) order by rr.total_score desc, np.fantasy_score desc;

-- name: GetWeeksLeaderboardByGuildId :many
select dr.discord_player_id, dr.nickname, sum(np.fantasy_score) as fantasy_score from daily_roster dr join nba_players np on np.nba_player_uid = dr.nba_player_uid join discord_player_guilds dpg on dpg.discord_player_id = dr.discord_player_id where dpg.guild_id = sqlc.arg(guildid) and dr.date < sqlc.arg(endday) and dr.date > sqlc.arg(startday) group by dr.discord_player_id, dr.nickname order by fantasy_score desc limit 100;

-- name: GetTodaysLeaderboardByGuildId :many
select dr.discord_player_id, np.nba_player_id, dr.nba_player_uid, dr.date, dr.nickname, dr.position as position, np.name, np.dollar_value, np.fantasy_score from daily_roster dr join nba_players np on np.nba_player_uid = dr.nba_player_uid join discord_player_guilds dpg on dpg.discord_player_id = dr.discord_player_id where dr.date = sqlc.arg(date) and dpg.guild_id = sqlc.arg(guildid) order by np.fantasy_score desc limit 1000;

-- name: GetTeamByPlayerUid :many
select t.* from nba_players np join teams t on t.team_id = np.team_id where np.nba_player_uid = sqlc.arg(playerid);

-- name: FindNbaPlayerByUid :one
select * from nba_players where nba_player_uid = sqlc.arg(nbaplayeruid);

-- name: FindNbaPlayerByName :one
select * from nba_players where name = sqlc.arg(name);

-- name: GetTodaysNbaPlayersByPosition :many
select np.* from nba_players np where np.position = sqlc.arg(position) order by np.dollar_value desc limit 5;

-- name: GetAllTodaysNbaPlayers :many
select np.* from nba_players np order by np.dollar_value desc limit 25;

-- name: GetNbaPlayersWithTeam :many
select np.name as player_name, t.abbr as team_name, at.abbr as against_team_name, np.* from nba_players np join teams t on t.team_id = np.team_id join teams at on np.against_team = at.team_id where np.date = sqlc.arg(date) order by np.dollar_value desc;

-- name: IsTodayLocked :one
select il.date, il.lock_time from is_locked il where il.date = current_date;

-- name: IsLocked :one
select il.date, il.lock_time from is_locked il where il.date = sqlc.arg(date);

-- name: InsertGuildForPlayerId :exec
insert into discord_player_guilds (discord_player_id, guild_id) values (sqlc.arg(discordplayerid), sqlc.arg(guildid)) on conflict (discord_player_id, guild_id) do nothing;

