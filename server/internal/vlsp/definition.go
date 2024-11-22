package vlsp

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/chrehall68/vls/internal/lang"
	"go.lsp.dev/protocol"
)

type LocationDetails struct {
	token         lang.Token
	currentModule string
}

func (h Handler) getLocationDetails(fname string, line int, character int) (*LocationDetails, error) {
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
			h.state.log.Sugar().Info("lineTokens: ", tokens)
			if err == nil {
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
	tokenStart := 0

	for _, token := range tokens {
		tokenEnd := tokenStart + len(token.Value)
		if tokenStart <= int(character) && int(character) < tokenEnd {
			// this is the result
			return &LocationDetails{token: token, currentModule: curModule}, nil
		}
		tokenStart = tokenEnd
	}
	return nil, fmt.Errorf("no token at that position")
}

func (h Handler) jumpTo(fname string, line int, character int) ([]protocol.Location, error) {
	details, err := h.getLocationDetails(fname, line, character)

	if err != nil {
		return nil, err
	}

	// process this token
	result := []protocol.Location{}
	if details.token.Type == "identifier" {
		// see if it's a module or definition
		location, ok := h.state.symbolMap[details.token.Value]
		if ok {
			result = append(result, location)
		} else {
			// otherwise, maybe it's a variable
			moduleMap, ok := h.state.variableDefinitions[details.currentModule]
			if ok {
				// look for variable definition
				location, ok := moduleMap[details.token.Value]
				if ok {
					result = append(result, location)
				}
			}
		}
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
