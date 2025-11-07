package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/arjunmoola/nba-daily-fantasy-bot/internal/database"
	//"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"context"
	"fmt"
	"log"
	"os"
	//"strings"
	"reflect"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)

	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Panic(err)
	}
	
	if err := migrate(context.Background(), pool); err != nil {
		log.Panic(err)
	}
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)

	if err != nil {
		return err
	}

	defer conn.Release()

	tx, err := conn.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	queries := database.New(tx)

	if queries == nil {
		fmt.Println("queries is nil")
	}

	dailyRosters, err := queries.GetAllDailyRoster(ctx)

	if err != nil {
		return err
	}

	fmt.Println("number of non distinct pairs", len(dailyRosters))

	type pair struct {
		GuildID string
		DiscordPlayerID string
	}

	distinctDailyRoster, err := queries.GetDistinctDailyRoster(ctx)

	if err != nil {
		return err
	}

	distinctDailies := make([]pair, 0, len(distinctDailyRoster))

	for _, value := range distinctDailyRoster {
		distinctDailies = append(distinctDailies, pair{
			GuildID: value.GuildID,
			DiscordPlayerID: value.DiscordPlayerID,
		})
	}

	fmt.Println(len(distinctDailyRoster))

	guilds, err := queries.GetDiscordPlayerGuilds(ctx)

	if err != nil {
		return err
	}

	guildPairs := make([]pair, 0, len(guilds))
	guildMap := make(map[pair]bool)
	dailiesMap := make(map[pair]bool)

	for _, guild := range guilds {
		guildPairs = append(guildPairs, pair{
			GuildID: guild.GuildID,
			DiscordPlayerID: guild.DiscordPlayerID,
		})
	}

	for _, value := range distinctDailies {
		dailiesMap[value] = true
	}

	for _, guild := range guildPairs {
		guildMap[guild] = true
	}

	var diff []pair
	var common []pair

	for _, value := range distinctDailies {
		if !guildMap[value] {
			diff = append(diff, value)
		} else {
			common = append(common, value)
		}
	}

	fmt.Println("number of elements not in existing database ", len(diff))
	fmt.Println("number of elements already in existing database", len(common))

	params := make([]database.InsertDiscordPlayerGuildsParams, 0, len(diff))

	for _, value := range diff {
		params = append(params, database.InsertDiscordPlayerGuildsParams{
			GuildID: value.GuildID,
			DiscordPlayerID: value.DiscordPlayerID,
		})
	}

	n, err := queries.InsertDiscordPlayerGuilds(ctx, params)

	if err != nil {
		return err
	}

	if int(n) != len(params) {
		return fmt.Errorf("incorrect number of rows inserted rolling")
	}

	allRows, err := queries.GetDiscordPlayerGuilds(ctx)

	if err != nil {
		return err
	}

	var allpairs []pair

	for _, row := range allRows {
		allpairs = append(allpairs, pair{
			GuildID: row.GuildID,
			DiscordPlayerID: row.DiscordPlayerID,
		})
	}

	var diff2 []pair

	for _, row := range allpairs {
		if !dailiesMap[row] {
			diff2 = append(diff2, row)
		}
	}

	if len(diff2) != 0 {
		return fmt.Errorf("incorrect number of rows inserted")
	}

	tx.Commit(context.Background())

	return nil
}

func getRows(ctx context.Context, conn *pgxpool.Conn, query string) ([][]any, error) {
	rows, err := conn.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var resultRows [][]any

	for rows.Next() {
		row, err := rows.Values()

		if err != nil {
			return nil, err
		}

		for _, val := range row {
			switch val.(type) {
			case string:
				fmt.Println(val)
			}
		}

		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resultRows, nil
}

type SchemaRow struct {
	Schema string
	TableName string
	OrdinalPosition string
	ColumnName string
	DataType string
	IsNullable string
	ColumnDefault string
}

func parseSchemaRows(rows [][]any) ([]SchemaRow, error) {
	var result []SchemaRow

	for _, row := range rows {
		var schema SchemaRow

		if err := setFieldsInOrder(&schema, row); err != nil {
			return nil, err
		}

		result = append(result, schema)
	}

	return result, nil
}

func setFieldsInOrder(ptr any, row []any) error {
	rv := reflect.ValueOf(ptr)

	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("ptr must be a pointer to a struct")
	}

	v := rv.Elem()
	t := v.Type()

	if len(row) > t.NumField() {
		return fmt.Errorf("struct and array must have the same number of fields")
	}

	for i := range row {
		if row[i] == nil {
			continue
		}

		field := v.Field(i)

		if field.CanSet() {
		}

		val := reflect.ValueOf(row[i])

		if val.Type().AssignableTo(field.Type()) {
			field.Set(val)
		} else if val.Type().ConvertibleTo(field.Type()) {
			field.Set(val.Convert(field.Type()))
		} else {
			return fmt.Errorf("error types are not assignable")
		}
	}

	return nil
}

func GetSchema(ctx context.Context, pool *pgxpool.Pool) ([][]any, error) {
	query := `
	SELECT
        c.table_schema,
        c.table_name,
        c.ordinal_position,
        c.column_name,
        c.data_type,
        c.is_nullable,
        c.column_default
	FROM information_schema.columns c
	JOIN information_schema.tables t
	ON t.table_schema = c.table_schema AND t.table_name = c.table_name
	WHERE t.table_type = 'BASE TABLE'
	AND c.table_schema NOT IN ('pg_catalog', 'information_schema')
	AND c.table_schema = 'public'
	ORDER BY c.table_schema, c.table_name, c.ordinal_position;
`

	conn, err := pool.Acquire(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Release()

	rows, err := getRows(ctx, conn, query)

	if err != nil {
		return nil, err
	}

	return rows, nil
}
