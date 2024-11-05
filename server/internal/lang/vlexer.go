package lang

import (
	"errors"
	"fmt"
	"regexp"
)

type VLexer struct {
	Lexer
}

func NewVLexer() *VLexer {
	vlexer := &VLexer{
		Lexer: *NewLexer(),
	}

	// add mappings
	// whitespace
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^[\t ]+`), "whitespace")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^[\r\n]+`), "newline")
	// comments
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\/\/.*`), "comment")
	// keywords
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^module`), "module")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^endmodule`), "endmodule")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^begin`), "begin")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^end`), "end")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^case`), "case")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^endcase`), "endcase")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^generate`), "generate")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^endgenerate`), "endgenerate")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^for`), "for")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^if`), "if")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^else`), "else")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^assign`), "assign")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^initial`), "initial")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^((negedge)|(posedge))`), "time")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^default`), "default")
	// comparisons/assignments
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^((\=\=)|(\!\=)|(\<\=)|(>\=)|\>|\<|(\=\=\=)|(\!\-\=))`), "comparator")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^((\&\&)|(\|\|))`), "logical_operator")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^[\+\-\*\/\|&]`), "operator")
	// symbols
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\(`), "lparen")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\)`), "rparen")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\[`), "lbracket")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\]`), "rbracket")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\{`), "lcurl")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\}`), "rcurl")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^:`), "colon")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\,`), "comma")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\.`), "dot")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\;`), "semicolon")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\?`), "question")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\@`), "at")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\=+`), "equal")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\~`), "tilde")
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^\#`), "pound")
	// other
	vlexer.AddMappingNoCapture(regexp.MustCompile("^`include.+"), "include")
	vlexer.AddMappingNoCapture(regexp.MustCompile("^`define"), "define")
	vlexer.AddMappingNoCapture(regexp.MustCompile("^`timescale"), "timescale")
	vlexer.AddMapping(regexp.MustCompile(`^\$.*`), func(code string) (Token, error) {
		return Token{Type: "unknown", Value: code}, nil
	})
	// variable-related
	vlexer.AddMappingNoCapture(regexp.MustCompile(`^((reg)|(wire)|(genvar)|(parameter)|(input)|(output)|(defparam))`), "type")
	vlexer.AddMapping(regexp.MustCompile("^`?[A-Za-z][a-zA-Z0-9_]*"), func(code string) (Token, error) {
		re := regexp.MustCompile("^`?(?P<IDENTIFIER>[A-Za-z][a-zA-Z0-9_]*)")
		matches := re.FindStringSubmatch(code)
		if len(matches) == 0 {
			fmt.Println("failed to parse identifier on ", code)
			return Token{}, errors.New("failed to parse identifier")
		}
		return Token{Type: "identifier", Value: matches[re.SubexpIndex("IDENTIFIER")]}, nil
	})
	vlexer.AddMapping(regexp.MustCompile(`^(([0-9]+)|([0-9]*\'[hbd][0-9xzXZA-Fa-f]+)|(\"[^\s]*\"))`), func(code string) (Token, error) {
		re := regexp.MustCompile(`^(?P<LITERAL>(([0-9]+)|([0-9]*\'[hbd][0-9xzXZA-Fa-f]+)|(\"[^\s]*\")))`)
		matches := re.FindStringSubmatch(code)
		if len(matches) == 0 {
			err := fmt.Sprintln("failed to parse literal on ", code)
			return Token{}, errors.New(err)
		}
		return Token{Type: "literal", Value: matches[re.SubexpIndex("LITERAL")]}, nil
	})

	return vlexer
}
