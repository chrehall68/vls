package main

/*
import (
	"github.com/chrehall68/vls/internal/vlsp"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopmentConfig().Build()

	// Start the server
	vlsp.StartServer(logger)
}*/
import (
	"fmt"
	"github.com/chrehall68/vls/internal/lang"
	"os"
	"strings"
)

func getFileFullPaths(dir string) []string {
	entries, _ := os.ReadDir(dir)
	paths := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			paths = append(paths, getFileFullPaths(dir+"/"+entry.Name())...)
		} else {
			paths = append(paths, dir+"/"+entry.Name())
		}
	}
	return paths
}

func main() {
	vlexer := lang.NewVLexer()
	dir := os.Args[1]

	files := getFileFullPaths(dir)
	fileToTokensMap := map[string][]lang.Token{}
	for _, file := range files {
		if strings.HasSuffix(file, ".v") {
			// read file if it's a verilog file
			b, err := os.ReadFile(file)
			if err != nil {
				panic(err)
			}
			text := string(b)

			// get tokens
			tokens := vlexer.Lex(text)
			if len(tokens) == 0 {
				fmt.Println("no tokens found for file: ", file)
			}

			// add to map
			fileToTokensMap[file] = tokens
		}
	}

	// now, let's parse and print out results
	fmt.Println("Done lexing")
	parser := lang.NewParser()
	for file, tokens := range fileToTokensMap {
		results := parser.Parse(tokens)
		fmt.Println("Parse results for file: ", file)
		fmt.Println(results)
	}
}
