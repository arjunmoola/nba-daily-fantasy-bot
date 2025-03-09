package main

import (
    "time"
    "log"
    "fmt"
    "github.com/bwmarrin/discordgo"
    "strings"
    "slices"
    "cmp"
    "context"
)

func interactionAuthor(i *discordgo.Interaction) *discordgo.User {
    if i.Member != nil {
        return i.Member.User
    }
    return i.User
}


func setRosterHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        switch i.Type{
        case discordgo.InteractionApplicationCommand:
            if err := handleSetRosterInteractionApplicationCommand(bot, s, i); err != nil {
                log.Println(err)
                s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
                return
            }
        case discordgo.InteractionApplicationCommandAutocomplete:
            if err := handleSetRosterInteractionAutocomplete(bot, s, i); err != nil {
                log.Panic(err)
            }
        }
    }
}

func setPGHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        switch i.Type{
        case discordgo.InteractionApplicationCommand:
            if err := handleSetPGInteractionApplicationCommand(bot, s, i); err != nil {
                log.Println(err)
                s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
                return
            }
        case discordgo.InteractionApplicationCommandAutocomplete:
        if err := handleSetPGInteractionAutocomplete(bot, s, i); err != nil {
                log.Println(err)
                s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
                return
            }
        }
    }
}

func getPGHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        players := bot.cache.getPlayersByPos("PG")

        builder := strings.Builder{}

        for _, player := range players {
            builder.WriteString(fmt.Sprintf("%s %d\n", player.Name, player.DollarValue))
        }

        err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: builder.String(),
                Flags: discordgo.MessageFlagsEphemeral,
            },
        })

        if err != nil {
            log.Println(err)
            errInteractionRespond(s, i, "something went wrong")
            return
        }

        ch, err := s.UserChannelCreate(interactionAuthor(i.Interaction).ID)

        if err != nil {
            log.Println(err)

            _, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
                Content: fmt.Sprintf("Mission failed. Cannot send a message to this error: %v", err),
                Flags: discordgo.MessageFlagsEphemeral,
            })

            if err != nil {
                log.Println(err)
                return
            }
        }

        _, err = s.ChannelMessageSend(ch.ID, fmt.Sprintf("Hi this is a channel message"))

        if err != nil {
            log.Println(err)
            return
        }

        _, err = s.ChannelMessageSend(ch.ID, fmt.Sprintf("Hi this is another message"))

        if err != nil {
            log.Println(err)
            return
        }

        for i := range 3 {
            time.Sleep(time.Second)

            _, err = s.ChannelMessageSend(ch.ID, fmt.Sprint("Hi this is another message %d", i))

            if err != nil {
                log.Println(err)
                return
            }
        }
    }
}

func getMyRosterHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        author := interactionAuthor(i.Interaction)
        guildId := i.Interaction.GuildID
        discordPlayerId := author.ID

        date := time.Now().Format(time.DateOnly)

        discordPlayers, err := bot.client.getMyRosterGuild(context.Background(), guildId, discordPlayerId, date)

        if err != nil {
            log.Println(err)
            errInteractionRespond(s, i, "unable to get my roster. http error")
            return
        }

        builder := strings.Builder{}

        totalValue := 0

        for _, player := range discordPlayers {
            rosterPlayer := player.getRosterPlayer()

            totalValue += rosterPlayer.DollarValue
            builder.WriteString(fmt.Sprintf("%s %s %d\n", rosterPlayer.Name, rosterPlayer.Position, rosterPlayer.DollarValue))
        }

        remainingValue := 100 - totalValue

        builder.WriteString(fmt.Sprintf("remaining value: %d\n", remainingValue))

        err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: builder.String(),
                Flags: discordgo.MessageFlagsEphemeral,
            },
        })

        if err != nil {
            log.Println(err)
            return
        }
    }
}

func handleSetPGInteractionApplicationCommand(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("You picked %q", data.Options[0].StringValue()),
        },
    })
    return err
}

func handleSetPGInteractionAutocomplete(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    var choices []*discordgo.ApplicationCommandOptionChoice

    players := bot.cache.getPlayersByPos("PG")
    choices = createPlayerChoices(players)

    if data.Options[0].StringValue() != "" {
        scores := getClosestPlayers(data.Options[0].StringValue(), players)
        choices = createChoicesFromScores(scores)
    }

    return autocompleteInteractionRespond(s, i, choices)
}

func handleSetRosterInteractionApplicationCommand(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    builder := strings.Builder{}

    pos := []string{ "PG", "C", "SF", "PF", "SG" }

    var totalSum int64

    for j, option := range data.Options {
        player, ok := bot.cache.getPlayer(option.Value.(string))

        if !ok {
            continue
        }

        totalSum += player.DollarValue

        if totalSum > 100 {
            return s.InteractionRespond(i.Interaction, errResponseInteraction("exceeded maximum budget"))
        }

        payload := myRosterPayload{
            PlayerId: player.UID,
            Position: player.Position,
            GuildId: i.Interaction.GuildID,
            Nickname: interactionAuthor(i.Interaction).Username,
            DiscordPlayerId: interactionAuthor(i.Interaction).ID,
            Date: time.Now().Format(time.DateOnly),
        }

        if err := bot.client.setMyRoster(context.Background(), payload); err != nil {
            return errInteractionRespond(s, i, "http post error")
        }

        builder.WriteString(fmt.Sprintf("You picked %q for %s\n", player.Name, pos[j]))
    }

    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: builder.String(),
            Flags: discordgo.MessageFlagsEphemeral,
        },
    })

    return err
}

func handleSetRosterInteractionAutocomplete(bot *NbaFantasyBot, s *discordgo.Session, i *discordgo.InteractionCreate) error {
    data := i.ApplicationCommandData()

    var choices []*discordgo.ApplicationCommandOptionChoice

    fmt.Println(data.Options)

    switch {
    case data.Options[0].Focused:
        players := bot.cache.getPlayersByPos("PG")
        choices = createPlayerChoices(players)
        if data.Options[0].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(),opt.Value)
            }
            scores := getClosestPlayers(data.Options[0].StringValue(), players)
            choices = createChoicesFromScores(scores)
            fmt.Println(choices)
        }
    case data.Options[1].Focused:
        players := bot.cache.getPlayersByPos("C")
        choices = createPlayerChoices(players)
        if data.Options[1].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(), opt.Value)
            }
            scores := getClosestPlayers(data.Options[1].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[2].Focused:
        players := bot.cache.getPlayersByPos("SF")
        choices = createPlayerChoices(players)
        if data.Options[2].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(), opt.Value)
            }
            scores := getClosestPlayers(data.Options[2].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[3].Focused:
        players := bot.cache.getPlayersByPos("PF")
        choices = createPlayerChoices(players)
        if data.Options[3].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(), opt.Value)
            }
            scores := getClosestPlayers(data.Options[3].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    case data.Options[4].Focused:
        players := bot.cache.getPlayersByPos("SG")
        choices = createPlayerChoices(players)
        if data.Options[4].StringValue() != "" {
            for _, opt := range data.Options {
                fmt.Println(opt.StringValue(), opt.Value)
            }
            scores := getClosestPlayers(data.Options[4].StringValue(), players)
            choices = createChoicesFromScores(scores)
        }
    }


    return autocompleteInteractionRespond(s, i, choices)

}

func autocompleteInteractionRespond(s *discordgo.Session, i *discordgo.InteractionCreate, choices []*discordgo.ApplicationCommandOptionChoice) error {
    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionApplicationCommandAutocompleteResult,
        Data: &discordgo.InteractionResponseData{
            Choices: choices,
        },
    })
}

type playerScore struct {
    player string
    score int
    dollarValue int64
    id int
}

func getClosestPlayers(src string, players []NbaPlayer) []playerScore {
    scores := make([]playerScore, len(players))

    for i, player := range players {
        s := computeSmithWatermanFunc(strings.ToLower(src), strings.ToLower(player.Name), defaultScorer)

        scores[i] = playerScore{
            player: player.Name,
            score: s,
            dollarValue: player.DollarValue,
            id: player.ID,
        }
    }

    scores = filterFunc(scores, greaterThanThreshold)

    slices.SortFunc(scores, func(a, b playerScore) int {
        return -1 * cmp.Compare(a.score, b.score)
    })

    return scores

}

func filterFunc[S []T,T any](s S, f func(T) bool) S {
    idx := 0
    for _, x := range s {
        if f(x) {
            s[idx] = x
            idx++
        }
    }
    s = s[:idx]
    return s
}

func greaterThanThreshold(x playerScore) bool {
    if x.score >= 4 {
        return true
    } else {
        return false
    }
}

func createPlayerChoicesByPos(bot *NbaFantasyBot, pos string) []*discordgo.ApplicationCommandOptionChoice {
    players := bot.cache.getPlayersByPos(pos)
    return createPlayerChoices(players)
}

func createPlayerChoices(players []NbaPlayer) []*discordgo.ApplicationCommandOptionChoice {
    choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, 25)

    if len(players) > 25 {
        players = players[:25]
    }

    for _, player := range players {
        choice := &discordgo.ApplicationCommandOptionChoice{
            Name: fmt.Sprintf("%s %d", player.Name, player.DollarValue),
            Value: fmt.Sprintf("%d", player.ID),
        }

        choices = append(choices, choice)
    }

    return choices
}

func createChoicesFromScores(scores []playerScore) []*discordgo.ApplicationCommandOptionChoice {
    choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, 25)

    if len(scores) > 25 {
        scores = scores[:25]
    }

    for _, score := range scores {
        choice := &discordgo.ApplicationCommandOptionChoice {
            Name: fmt.Sprintf("%s %d", score.player, score.dollarValue),
            Value: fmt.Sprintf("%d", score.id),
        }

        choices = append(choices, choice)
    }

    return choices
}

func getGlobalRosterCommandHandler(bot *NbaFantasyBot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return func(s *discordgo.Session, i *discordgo.InteractionCreate) {

        date := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)

        discordPlayers, err := bot.client.getGlobalRoster(context.Background(), date)

        if err != nil {
            log.Println(err)
            s.InteractionRespond(i.Interaction, errResponseInteraction("something went wrong"))
            return
        }

        user := interactionAuthor(i.Interaction)

        fmt.Println(user.ID, i.GuildID)

        leaderBoard := constructLeaderBoard(newGlobalRoster(discordPlayers))

        builder := strings.Builder{}

        fmt.Fprintln(&builder, "### Global Leaderboard\n")


        for _, player := range leaderBoard {
            builder.WriteString(fmt.Sprintf("**%s** %.2f\n", player.Nickname, player.TotalScore))

            for _, nbaPlayer := range player.players {
                builder.WriteString(fmt.Sprintf("\t%s %s %.2f\n", nbaPlayer.Position, nbaPlayer.Name, nbaPlayer.FantasyScore))
            }
        }

        err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: builder.String(),
            },
        })

        if err != nil {
            log.Println(err)
        }
    }
}

func errResponseInteraction(msg string) *discordgo.InteractionResponse {
    return &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: msg,
        },
    }
}

func errInteractionRespond(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
    return s.InteractionRespond(i.Interaction, errResponseInteraction(msg))
}
