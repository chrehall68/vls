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
				protocol.SemanticTokenType,      // 0
				protocol.SemanticTokenComment,   // 1
				protocol.SemanticTokenNumber,    // 2
				protocol.SemanticTokenMacro,     // 3
				protocol.SemanticTokenVariable,  // 4
				protocol.SemanticTokenClass,     // 5
				protocol.SemanticTokenParameter, // 6
				protocol.SemanticTokenFunction,  // 7
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
		"comment":         1,
		"type":            0,
		"direction":       0,
		"defparam":        0,
		"literal":         2,
		"module":          3,
		"endmodule":       3,
		"begin":           3,
		"end":             3,
		"case":            3,
		"endcase":         3,
		"generate":        3,
		"endgenerate":     3,
		"for":             3,
		"if":              3,
		"else":            3,
		"assign":          3,
		"initial":         3,
		"always":          3,
		"time":            3,
		"default":         3,
		"include":         3,
		"timescale":       3,
		"define":          3,
		"identifier":      4,
		"existing_module": 5,
		"port":            6,
		"funcliteral":     7,
		"signed":          7,
		"dollar":          7,
		"pound":           7,
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
	f := URIToPath(string(params.TextDocument.URI))
	contents := h.state.files[f].GetContents()

	// extract tokens
	lexer := lang.NewVLexer(h.state.log)
	tokens, _ := lexer.Lex(contents)

	// extract ast if possible
	ast, err := lang.NewParser().ParseFile(tokens)
	if err == nil {
		h.state.log.Sugar().Info("Getting statements for file: ", f)
		interiorNodes := lang.GetInteriorStatements(ast)
		tokensIdx := 0

		for _, interiorNode := range interiorNodes {
			if interiorNode.ModuleApplicationNode != nil {
				// get to the current module
				for tokens[tokensIdx] != interiorNode.ModuleApplicationNode.ModuleName && tokensIdx < len(tokens) {
					tokensIdx++
				}

				// then label it as a module name
				if tokensIdx < len(tokens) {
					tokens[tokensIdx].Type = "existing_module"
				}

				for _, argument := range interiorNode.ModuleApplicationNode.Arguments {
					if argument.Label != nil {
						// get to this label and label it as a port
						for tokens[tokensIdx] != *argument.Label && tokensIdx < len(tokens) {
							tokensIdx++
						}

						if tokensIdx < len(tokens) {
							tokens[tokensIdx].Type = "port"
						}
					}
				}
			}
		}

		// do similar thing for functions
		tokensIdx = 0
		functionNodes := lang.GetFunctionNodes(ast)

		for _, functionNode := range functionNodes {
			// get to the function name
			for tokens[tokensIdx] != functionNode.Function && tokensIdx < len(tokens) {
				tokensIdx++
			}

			if tokensIdx < len(tokens) {
				tokens[tokensIdx].Type = "funcliteral"
			}
		}
	}

	// encode
	result := &protocol.SemanticTokens{
		Data: Encode(tokens),
	}
	h.state.log.Sugar().Info("SemanticTokensFull result: ", result)

	return result, nil
}
