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
	lineString := ""
	for l := 0; l <= line; l++ {
		lineString, _ = reader.ReadString('\n')
	}
	lexer := lang.NewVLexer(h.state.log)
	tokens, _ := lexer.Lex(lineString)

	h.state.log.Sugar().Info("tokens: ", tokens)
	tokenStart := 0
	result := []protocol.Location{}
	for _, token := range tokens {
		tokenEnd := tokenStart + len(token.Value)
		if tokenStart <= int(character) && int(character) < tokenEnd {
			// process this token
			if token.Type == "identifier" {
				// see if it's a module or definition
				location, ok := h.state.symbolMap[token.Value]
				if ok {
					result = append(result, location)
				}
			}
		}
		tokenStart = tokenEnd
	}

	return result, nil
}

func (h Handler) Declaration(ctx context.Context, params *protocol.DeclarationParams) (result []protocol.Location, err error) {
	fname := params.TextDocument.URI.Filename()
	pos := params.TextDocumentPositionParams.Position

	return h.jumpTo(fname, int(pos.Line), int(pos.Character))
}
func (h Handler) Definition(ctx context.Context, params *protocol.DefinitionParams) (result []protocol.Location, err error) {
	fname := params.TextDocument.URI.Filename()
	pos := params.TextDocumentPositionParams.Position

	return h.jumpTo(fname, int(pos.Line), int(pos.Character))
}
func (h Handler) Implementation(ctx context.Context, params *protocol.ImplementationParams) (result []protocol.Location, err error) {
	fname := params.TextDocument.URI.Filename()
	pos := params.TextDocumentPositionParams.Position

	return h.jumpTo(fname, int(pos.Line), int(pos.Character))
}
