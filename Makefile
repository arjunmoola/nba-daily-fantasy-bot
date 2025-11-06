build:
	go build -o bin/ github.com/arjunmoola/nba-daily-fantasy-bot/cmd/server

clean:
	rm -rf bin/
