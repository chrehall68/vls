package vlsp

import (
	"os"
	"strings"

	"github.com/chrehall68/vls/internal/lang"
)

func (h *Handler) getFileFullPaths(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		h.state.log.Sugar().Panic("while reading dir: %w", err)
	}
	paths := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			paths = append(paths, h.getFileFullPaths(dir+"/"+entry.Name())...)
		} else {
			paths = append(paths, dir+"/"+entry.Name())
		}
	}
	return paths
}

func (h Handler) GetSymbols() {
	vlexer := lang.NewVLexer(h.state.log)
	parser := lang.NewParser()
	files := h.getFileFullPaths(h.state.workspace)

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

			// add to map
			fileToTokensMap[file] = tokens
		}
	}

	// now, let's parse and print out results
	for _, tokens := range fileToTokensMap {
		results := parser.Parse(tokens)
		h.state.defines = append(h.state.defines, results.Defines...)
		h.state.modules = append(h.state.modules, results.Modules...)
	}
}
