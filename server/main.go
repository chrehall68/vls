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
	"bufio"
	"fmt"
	"github.com/chrehall68/vls/internal/lang"
	"os"
	"strings"
)

func main() {
	vlexer := lang.NewVLexer()
	reader := bufio.NewReader(os.Stdin)

	fileToParse, err := reader.ReadString('\n')
	for err == nil {
		fileToParse = strings.TrimSuffix(fileToParse, "\n")
		fmt.Println(fileToParse)
		b, err := os.ReadFile(fileToParse)
		if err != nil {
			panic(err)
		}
		text := string(b)
		tokens := vlexer.Lex(text)
		if len(tokens) == 0 {
			fmt.Println("no tokens found")
		}
		fileToParse, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
