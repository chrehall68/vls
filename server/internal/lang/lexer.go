package lang

import (
	"errors"
	"fmt"
	"regexp"
)

type Token struct {
	Type  string
	Value string
}

type Lexer struct {
	mappings map[*regexp.Regexp]func(string) (Token, error)
}

func NewLexer() *Lexer {
	return &Lexer{
		mappings: map[*regexp.Regexp]func(string) (Token, error){},
	}
}

// AddMapping adds a mapping to the lexer
// the pattern should probably start with a ^ to indicate
// the start of the string
func (l *Lexer) AddMapping(pattern *regexp.Regexp, mapper func(string) (Token, error)) {
	l.mappings[pattern] = mapper
}

// helper to make adding a mapping easier when you don't need to capture
// the value
func (l *Lexer) AddMappingNoCapture(pattern *regexp.Regexp, Type string) {
	l.AddMapping(pattern, func(code string) (Token, error) {
		return Token{Type: Type, Value: code}, nil
	})
}

func (l *Lexer) Lex(code string) []Token {
	i := 0
	tokens := []Token{}
	for i < len(code) {
		// figure out which of the tokens will consume
		// the most characters, and match that token
		// with the code
		maxLength := 0
		f := func(_ string) (Token, error) {
			return Token{}, errors.New("no token found")
		}
		for pattern, mapping := range l.mappings {
			length := len(pattern.FindString(code[i:]))
			if length > maxLength {
				maxLength = length
				f = mapping
			}
		}

		// if no token was found
		if maxLength == 0 {
			fmt.Println("no token found!! Still have code: ", code[i:])
			fmt.Println("current char: ", code[i])
			break
		}

		// now, match the token with the code
		token, err := f(code[i : i+maxLength])
		if err == nil { // don't add empty tokens
			tokens = append(tokens, token)
		} else {
			panic(err)
		}
		i += maxLength
	}

	return tokens
}
