package vlsp

import (
	"os"
	"strings"

	"github.com/chrehall68/vls/internal/lang"
	"go.lsp.dev/protocol"
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

func (h Handler) GetSymbolsForFile(fname string) {
	vlexer := lang.NewVLexer(h.state.log)
	parser := lang.NewParser()

	// lex
	tokens, err := vlexer.Lex(h.state.files[fname].GetContents())
	if err != nil {
		h.state.log.Sugar().Errorf("error lexing file %s: %s", fname, err)
		return
	}

	// parse
	results, err := parser.Parse(tokens)
	if err != nil {
		h.state.log.Sugar().Errorf("error parsing file %s: %s", fname, err)
		h.state.defines[fname] = []lang.Define{}
		h.state.modules[fname] = []lang.Module{}
	} else {
		h.state.modules[fname] = results.Modules
		h.state.defines[fname] = results.Defines

		for _, module := range results.Modules {
			h.state.symbolMap[module.Name] = protocol.Location{
				URI: protocol.DocumentURI("file://" + fname),
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(module.Token.Line()), Character: uint32(module.Token.StartCharacter())},
					End:   protocol.Position{Line: uint32(module.Token.Line()), Character: uint32(module.Token.EndCharacter())}},
			}
		}
		for _, define := range results.Defines {
			h.state.symbolMap[define.Name] = protocol.Location{
				URI: protocol.DocumentURI("file://" + fname),
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(define.Token.Line()), Character: uint32(define.Token.StartCharacter())},
					End:   protocol.Position{Line: uint32(define.Token.Line()), Character: uint32(define.Token.EndCharacter())}},
			}
		}
	}
}

func (h Handler) GetSymbols() {
	// first, reset state
	h.state.defines = map[string][]lang.Define{}
	h.state.modules = map[string][]lang.Module{}
	h.state.symbolMap = map[string]protocol.Location{}

	// then, get the files to parse
	files := h.getFileFullPaths(h.state.workspace)

	for _, file := range files {
		if strings.HasSuffix(file, ".v") {
			// create the file object
			h.state.files[file] = NewFile(file)

			// parse the file
			h.GetSymbolsForFile(file)
		}
	}
}
