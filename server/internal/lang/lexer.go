package lang

import (
	"errors"
	"regexp"

	"go.uber.org/zap"
)

type Token struct {
	Type  string
	Value string
}

type Lexer struct {
	regexps []*regexp.Regexp
	funcs   []func(string) (Token, error)
	logger  *zap.Logger
}

func NewLexer(logger *zap.Logger) *Lexer {
	return &Lexer{
		regexps: []*regexp.Regexp{},
		funcs:   []func(string) (Token, error){},
		logger:  logger,
	}
}

// AddMapping adds a mapping to the lexer
// the pattern should probably start with a ^ to indicate
// the start of the string
func (l *Lexer) AddMapping(pattern *regexp.Regexp, mapper func(string) (Token, error)) {
	l.regexps = append(l.regexps, pattern)
	l.funcs = append(l.funcs, mapper)
}

// helper to make adding a mapping easier when you don't need to capture
// the value
func (l *Lexer) AddMappingNoCapture(pattern *regexp.Regexp, Type string) {
	l.AddMapping(pattern, func(code string) (Token, error) {
		return Token{Type: Type, Value: code}, nil
	})
}

func (l *Lexer) Lex(code string) []Token {
	tokens := []Token{}
	for i := 0; i < len(code); {
		// figure out which of the tokens will consume
		// the most characters, and match that token
		// with the code
		maxLength := 0
		f := func(_ string) (Token, error) {
			return Token{}, errors.New("no token found")
		}

		// enforce order of precedence (mappings inserted first take precedence)
		for j := 0; j < len(l.funcs); j++ {
			length := len(l.regexps[j].FindString(code[i:]))
			if length > maxLength {
				maxLength = length
				f = l.funcs[j]
			}
		}

		// if no token was found
		if maxLength == 0 {
			l.logger.Sugar().Errorf("no token found!! Still have code: ", code[i:])
			l.logger.Sugar().Errorf("current char: ", code[i])
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
