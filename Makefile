build:
	go build -o bin/ github.com/arjunmoola/nba-daily-fantasy-bot/cmd/server
	go build -o bin/ github.com/arjunmoola/nba-daily-fantasy-bot/cmd/spring-extractor

clean:
	rm -rf bin/
