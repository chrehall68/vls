package vlsp

import (
	"context"

	"github.com/chrehall68/vls/internal/lang"
	"go.lsp.dev/protocol"
)

type SemanticTokensOptions struct {
	protocol.WorkDoneProgressOptions
	/**
	 * The legend used by the server
	 */
	Legend protocol.SemanticTokensLegend `json:"legend"`

	/**
	 * Server supports providing semantic tokens for a full document.
	 */
	Full bool `json:"full,omitempty"`
	/**
	 * Server supports providing semantic tokens for a full document range.
	 */
	Range bool `json:"range,omitempty"`
}

func GetSemanticTokensOptions() SemanticTokensOptions {
	return SemanticTokensOptions{
		Legend: protocol.SemanticTokensLegend{
			TokenTypes: []protocol.SemanticTokenTypes{
				protocol.SemanticTokenType,     // 0
				protocol.SemanticTokenComment,  // 1
				protocol.SemanticTokenNumber,   // 2
				protocol.SemanticTokenMacro,    // 3
				protocol.SemanticTokenVariable, // 4
			},
			TokenModifiers: []protocol.SemanticTokenModifiers{},
		},
		Full:  true,
		Range: false,
	}
}

func Encode(tokens []lang.Token) []uint32 {
	result := []uint32{}
	prevLine := 0
	prevCharacter := 0

	// the bool doesn't matter
	// if it's in ignoreTokens, then ignore it
	// maps token type to int
	// the tokens to ignore are not in here
	tokenTypeToInt := map[string]uint32{
		"comment":     1,
		"type":        0,
		"literal":     2,
		"module":      3,
		"endmodule":   3,
		"begin":       3,
		"end":         3,
		"case":        3,
		"endcase":     3,
		"generate":    3,
		"endgenerate": 3,
		"for":         3,
		"if":          3,
		"else":        3,
		"assign":      3,
		"initial":     3,
		"time":        3,
		"default":     3,
		"identifier":  4,
	}

	addToken := func(token lang.Token) {
		// [deltaLine, deltaStart, length, tokenType, tokenModifiers]
		val, ok := tokenTypeToInt[token.Type]
		if ok {
			// if it's a new line, don't worry about character
			if token.Line() != prevLine {
				prevCharacter = 0
			}

			// add into result
			result = append(result, uint32(token.Line()-prevLine), uint32(token.StartCharacter()-prevCharacter), uint32(len(token.Value)), val, 0)

			// update line and character
			prevLine = token.Line()
			prevCharacter = token.StartCharacter()
		}
	}

	for _, token := range tokens {
		addToken(token)
	}
	return result

}

func (h Handler) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	h.state.log.Sugar().Info("SemanticTokensFull called")

	// get contents
	f := params.TextDocument.URI.Filename()
	contents := h.state.files[f].GetContents()

	// extract tokens
	lexer := lang.NewVLexer(h.state.log)
	tokens, _ := lexer.Lex(contents)

	// encode
	result := &protocol.SemanticTokens{
		Data: Encode(tokens),
	}
	h.state.log.Sugar().Info("SemanticTokensFull result: ", result)

	return result, nil
}
