
## NBA Daily Fantasy Discord Bot

A Discord bot that lets users manage and view NBA daily fantasy rosters via slash commands. It talks to a backend service for lock times, player data, and roster CRUD.

### Features
- Global slash command registration on startup
- Autocomplete for player selection by position
- Commands:
  - `set-roster`: Pick your PG, C, SF, PF, SG within a 100 budget
  - `my-roster`: View your roster and remaining budget for today
  - `global-roster`: View yesterday’s global leaderboard
  - `reset-roster`: Clear your roster for today

### Requirements
- Go 1.23+
- A Discord bot token
- Backend base URL exposing:
  - `GET /api/activity/lock-time`
  - `GET /api/activity/todays-players`
  - `GET /api/activity/rosters/global?date=YYYY-MM-DD`
  - `GET /api/activity/my-roster/{guildId}/{discordPlayerId}?date=YYYY-MM-DD`
  - `POST /api/activity/my-roster`
  - `DELETE /api/activity/my-roster`

### Environment
A `.env` file is required at runtime.

```bash
TOKEN=your_discord_bot_token
URL=https://your-backend-base-url
```

### Build
```bash
make build      # builds the server binary into ./bin/server
make clean      # removes ./bin/
```

### Run
After building:
```bash
./run.sh        # runs ./bin/server
# or
./bin/server
```

### How it works
- Loads `.env` using `github.com/joho/godotenv`
- Creates a Discord session using `TOKEN`
- Initializes a client with `URL` for all HTTP calls
- On `GetTodaysPlayers` empty response (roster locked), the bot falls back to `cache/players.json` if present
- Registers global slash commands on startup and listens for interactions until interrupt

### Project layout
- `cmd/server/main.go`: Application entrypoint (loads env, initializes and runs the bot)
- `internal/daily-fantasy-bot/`:
  - `bot.go`: Session setup, command registration, lifecycle (`Init`, `Run`)
  - `handlers.go`: Slash command handlers and responses
  - `commands.go`: Slash command definitions
  - `cache.go`: In-memory player cache by position and id
  - `roster.go`: Aggregation and leaderboard helpers
- `internal/client/client.go`: HTTP client for backend endpoints
- `internal/types/models.go`: Data models exchanged with the backend
- `internal/text/similarity.go`: String similarity (used for autocomplete)
- `Makefile`: Build and clean targets
- `go.mod`, `go.sum`: Module configuration

### Testing
```bash
go test ./...
```

### Module
- Module path: `github.com/arjunmoola/nba-daily-fantasy-bot`
