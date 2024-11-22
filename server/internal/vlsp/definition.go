package vlsp

import (
	"bufio"
	"context"
	"strings"

	"github.com/chrehall68/vls/internal/lang"
	"go.lsp.dev/protocol"
)

func (h Handler) jumpTo(fname string, line int, character int) ([]protocol.Location, error) {
	h.state.log.Sugar().Info("opening file", fname)
	f := h.state.files[fname].GetContents()
	reader := bufio.NewReader(strings.NewReader(f))
	lexer := lang.NewVLexer(h.state.log)
	parser := lang.NewParser()
	lineString := ""
	curModule := ""
	for l := 0; l <= line; l++ {
		lineString, _ = reader.ReadString('\n')

		// keep track of which module we're inside
		if strings.Contains(lineString, "module") {
			tokens, err := lexer.Lex(lineString)
			h.state.log.Sugar().Info("lineTOkens: ", tokens)
			if err == nil {
				h.state.log.Sugar().Info("Ok so far: ")
				for i := range tokens {
					if tokens[i].Type == "module" {
						// new module?
						pos, err := parser.CheckToken("", []string{"identifier"}, i+1, tokens)
						if err == nil {
							// pos contains module name
							curModule = tokens[pos].Value
						}
					}
				}
			}
		}
	}
	tokens, _ := lexer.Lex(lineString)

	h.state.log.Sugar().Info("tokens: ", tokens, "curModule: ", curModule, "line: ", line, "character: ", character)
	tokenStart := 0
	result := []protocol.Location{}
	for _, token := range tokens {
		tokenEnd := tokenStart + len(token.Value)
		if tokenStart <= int(character) && int(character) < tokenEnd {
			// process this token
			if token.Type == "identifier" {
				// see if it's a module or definition
				h.state.log.Sugar().Info("now looking for", token.Value)
				location, ok := h.state.symbolMap[token.Value]
				if ok {
					result = append(result, location)
				} else {
					// otherwise, maybe it's a variable
					moduleMap, ok := h.state.variableDefinitions[curModule]
					if ok {
						h.state.log.Sugar().Info("moduleMap: ", moduleMap)
						// look for variable definition
						location, ok := moduleMap[token.Value]
						if ok {
							result = append(result, location)
						}
					}
				}
			}
		}
		tokenStart = tokenEnd
	}

	return result, nil
}

func (h Handler) Declaration(ctx context.Context, params *protocol.DeclarationParams) (result []protocol.Location, err error) {
	fname := URIToPath(string(params.TextDocument.URI))
	pos := params.TextDocumentPositionParams.Position

	return h.jumpTo(fname, int(pos.Line), int(pos.Character))
}
func (h Handler) Definition(ctx context.Context, params *protocol.DefinitionParams) (result []protocol.Location, err error) {
	fname := URIToPath(string(params.TextDocument.URI))
	pos := params.TextDocumentPositionParams.Position

	return h.jumpTo(fname, int(pos.Line), int(pos.Character))
}
func (h Handler) Implementation(ctx context.Context, params *protocol.ImplementationParams) (result []protocol.Location, err error) {
	fname := URIToPath(string(params.TextDocument.URI))
	pos := params.TextDocumentPositionParams.Position

	return h.jumpTo(fname, int(pos.Line), int(pos.Character))
}
