-- name: GetAllDailyRoster :many
select guild_id, discord_player_id from public.daily_roster;

-- name: GetDistinctDailyRoster :many
select distinct guild_id, discord_player_id from public.daily_roster;

-- name: GetDiscordPlayerGuilds :many
select * from public.discord_player_guilds;

-- name: InsertDiscordPlayerGuilds :copyfrom
insert into public.discord_player_guilds
    (guild_id, discord_player_id)
values (
    sqlc.arg(guild_id),
    sqlc.arg(discord_player_id)
);
