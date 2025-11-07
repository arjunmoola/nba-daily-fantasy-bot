package main

import (
	//"errors"
	"strings"
	"github.com/joho/godotenv"
	"unicode"
	"bufio"
	"context"
	"bytes"
	"os"
	"fmt"
	"flag"
	"log"
	"io/fs"
	"path/filepath"
	"golang.org/x/sync/errgroup"
	"sync"
	"runtime"
)

func walk(ctx context.Context, dir string, out chan <-string) error {
	return filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.Type().IsRegular() && filepath.Ext(path) == ".java" {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case out <- path:
			}

			return nil
		}

		return nil
	})
}

type extractedFileResult struct {
	path string
	stmts []string
	err error
}

func sqlSplitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	leftTag := []byte("@Query")
	rightTag := []byte(");")
	//tag := []byte(`"""`)

	idx := bytes.Index(data, leftTag)

	if idx < 0 {
		if !atEOF {
			return 0, nil, nil
		}
		return 0, token, bufio.ErrFinalToken
	}
	
	leftIdx := idx

	advance += leftIdx

	idx = bytes.Index(data[leftIdx:], rightTag)

	if idx < 0 {
		if !atEOF {
			return advance, nil, nil
		}
		return 0, token, bufio.ErrFinalToken
	}

	rightIdx := leftIdx+idx

	token = data[leftIdx:rightIdx+2]
	advance += idx + len(rightTag)

	return advance, token, nil
}

func extractStatements(path string) ([]string, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(sqlSplitter)

	var stmts []string

	for scanner.Scan() {
		stmts = append(stmts, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return stmts, nil
}

func worker(ctx context.Context, in <-chan string, out chan <-extractedFileResult) {
	for {
		var path string

		select {
		case <-ctx.Done():
			return
		case p, ok := <-in:
			if !ok {
				return
			}

			path = p
		}

		stmts, err := extractStatements(path)

		result := extractedFileResult{
			path: path,
			stmts: stmts,
			err: err,
		}

		select {
		case <-ctx.Done():
			return
		case out <- result:
		}
	}
}

func extractSqlFromStatement(stmt []byte) string {
	tag := []byte(`"""`)

	idx := bytes.Index(stmt, tag)

	if idx < 0 {
		panic("expected to have tag")
	}

	stmt = stmt[idx+3:]

	idx = bytes.Index(stmt, tag)

	stmt = stmt[:idx]

	return string(stmt)
}

func cleanupStatement(stmt string) string {
	stmt = string(bytes.TrimSpace([]byte(stmt)))

	var buf bytes.Buffer

	isSpace := false

	for _, r := range stmt {
		if unicode.IsSpace(r) {
			if !isSpace {
				isSpace = true
				continue
			}
		} else {
			if isSpace {
				isSpace = false
				buf.WriteRune(' ')
				buf.WriteRune(r)
				continue
			}

			buf.WriteRune(r)
		}
	}

	return buf.String()
}

type springArg struct {
	pos int
	name string
	cast string
}

func isValidChar(r rune) bool {
	return unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_'
}

func extractArgsFromStatement(stmt []byte) []springArg {
	var args []springArg

	for idx := 0; idx < len(stmt);  {
		//fmt.Println(string(stmt[idx:]))
		leftIdx := bytes.IndexRune(stmt[idx:], ':')

		if leftIdx < 0 {
			break
		}

		idx += leftIdx+1

		start := idx

		for idx < len(stmt) {
			if !isValidChar(rune(stmt[idx])) {
				break
			}
			idx++
		}

		arg := springArg{
			pos: start,
			name: string(stmt[start:idx]),
		}

		var startCast int
		var hasCast bool

		if idx+1 < len(stmt) && bytes.HasPrefix(stmt[idx:], []byte("::")) {
			hasCast = true
			idx += 2
			startCast = idx

			for idx < len(stmt) {
				if !isValidChar(rune(stmt[idx])) {
					break
				}
				idx++
			}
		}

		if hasCast {
			arg.cast = string(stmt[startCast:idx])
		}

		args = append(args, arg)
	}

	return args
}

type replacementPair struct {
	old, new string
}

func createReplacementPairs(args []springArg) []replacementPair {
	var pairs []replacementPair

	for _, arg := range args {
		old := arg.name

		var idx int

		var parts []string

		for i, r := range old {
			if unicode.IsUpper(r) {
				part := old[idx:i]
				parts = append(parts, strings.ToLower(part))
				idx = i
			}
		}

		part := old[idx:]
		parts = append(parts, strings.ToLower(part))

		joinedStr := strings.Join(parts, "_")

		newStr := fmt.Sprintf("sqlc.arg(%s)", joinedStr)

		pairs = append(pairs, replacementPair{
			old: ":"+old,
			new: newStr,
		})
	}

	return pairs
}

func replacementPairsToList(pairs []replacementPair) []string {
	var args []string
	for _, pair := range pairs {
		args = append(args, pair.old, pair.new)
	}

	return args
}

func extract(ctx context.Context, springDir string, outputFile string) error {
	var g errgroup.Group

	pathsCh := make(chan string)

	g.Go(func() error {
		defer close(pathsCh)

		return walk(ctx, springDir, pathsCh)
	})
	
	out := make(chan extractedFileResult)

	var wg sync.WaitGroup

	for range runtime.NumCPU() {
		wg.Add(1)

		go func() {
			defer wg.Done()
			worker(ctx, pathsCh, out)
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	var allStmts []extractedFileResult

Loop:
	for {
		var result extractedFileResult

		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-out:
			if !ok {
				break Loop
			}

			result = res
		}

		if result.err != nil {
			log.Printf("error processing file %s %v", result.path, result.err)
			continue
		}

		allStmts = append(allStmts, result)
	}

	if err := g.Wait(); err != nil {
		return err
	}

	for _, res := range allStmts {
		if len(res.stmts) == 0 {
			continue
		}

		for _, stmt := range res.stmts {
			sqlStmt := extractSqlFromStatement([]byte(stmt))
			sqlStmt = cleanupStatement(sqlStmt)
			fmt.Println("old: ", sqlStmt)

			args := extractArgsFromStatement([]byte(sqlStmt))

			pairs := createReplacementPairs(args)

			replacementArgs := replacementPairsToList(pairs)

			newSql := strings.NewReplacer(replacementArgs...).Replace(sqlStmt)

			fmt.Println(newSql)
		}
	}

	if outputFile == "" {
		return nil
	}

	return nil
}

func main() {
	var springDir string
	var outputFile string

	godotenv.Load()

	defaultSpring := os.Getenv("DEFAULT_DIR")

	flag.StringVar(&springDir, "spring-dir", defaultSpring, "path to the directory to extract sql commands from")
	flag.StringVar(&outputFile, "out", "", "path to save the output sql statements")
	flag.Parse()

	if springDir == "" {
		fmt.Println("must specify what dir to use")
		os.Exit(1)
	}

	if err := extract(context.Background(), springDir, outputFile); err != nil {
		log.Fatal(err)
	}
}
