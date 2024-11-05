package lang

type Module struct {
	Name  string
	Ports []string
}

type GlobalResults struct {
	Modules []Module // defined modules
	Defines []string // defined identifiers
}

type Parser struct {
	skipTokens     []string
	topLevelTokens []string
}

func NewParser() *Parser {
	return &Parser{
		skipTokens:     []string{"whitespace", "comment", "newline"},
		topLevelTokens: []string{"module", "define", "include", "timescale"},
	}
}

/**
Grammar:

// top level grammar
<start> -> <statement> { <statement> }
<statement> -> <module> | <directive> | <skippable>

// sub grammars
<skippable> -> WHITESPACE | COMMENT | NEWLINE

<directive> -> <include> | <timescale> | <define>
<include> -> INCLUDE
<timescale> -> TIMESCALE <non-newline> NEWLINE
<define> -> DEFINE <identifier> <non-newline> NEWLINE
<identifier> -> IDENTIFIER


<module> -> MODULE <identifier> [<portlist>] SEMICOLON <interior> ENDMODULE
<portlist> -> LPAREN [<ports>] RPAREN
<ports> -> <identifier> { COMMA <identifier> }
*/

func (p *Parser) skip(tokens []Token, skippables []string, pos int) int {
	i := pos
	for ; i < len(tokens); i++ {
		skippable := false
		for j := 0; j < len(skippables); j++ {
			if tokens[i].Type == skippables[j] {
				skippable = true
				break
			}
		}
		if !skippable {
			break
		}
	}
	return i
}

// start with the skippable parts that we don't care about
func (p *Parser) skipInclude(tokens []Token, pos int) int {
	// skip over the include
	return p.skip(tokens, []string{"include"}, pos)
}
func (p *Parser) skipTimescale(tokens []Token, pos int) int {
	// skip over the timescale
	i := p.skip(tokens, []string{"timescale"}, pos)
	for (i < len(tokens)) && (tokens[i].Type != "newline") {
		i++
	}
	return i
}
func (p *Parser) skipModuleInterior(tokens []Token, pos int) int {
	// skip over the module interior
	i := pos
	for (i < len(tokens)) && (tokens[i].Type != "endmodule") {
		i++
	}
	return i
}

// now, the parts that we care about

// returns the name of the identifier
func (p *Parser) parseDefine(tokens []Token, pos int) (string, int) {
	// double check that the first token is a define
	if tokens[pos].Type != "define" {
		panic("expected a define for parseDefine")
	}

	i := p.skip(tokens, p.skipTokens, pos+1) // skip over any skippable tokens
	if tokens[i].Type != "identifier" {
		panic("expected an identifier for parseDefine")
	}
	identifier := tokens[i].Value
	// get to the end of the line
	for tokens[i].Type != "newline" {
		i++
	}
	return identifier, i
}
func (p *Parser) parsePorts(tokens []Token, pos int) ([]string, int) {
	identifiers := []string{}
	i := p.skip(tokens, p.skipTokens, pos) // get to first non-space
	for i < len(tokens) {
		// may be the end
		if tokens[i].Type == "rparen" {
			break
		}

		// check that it's an identifier before taking it
		if tokens[i].Type != "identifier" {
			panic("expected an identifier for parsePorts")
		}
		identifiers = append(identifiers, tokens[i].Value)

		// advance to either the comma or the rparen
		i = p.skip(tokens, p.skipTokens, i+1)
		if tokens[i].Type == "comma" {
			i = p.skip(tokens, p.skipTokens, i+1)
		} else if tokens[i].Type == "rparen" {
			break
		} else {
			panic("expected a comma or a rparen for parsePorts, got: " + tokens[i].Type)
		}
	}
	return identifiers, i
}
func (p *Parser) parsePortList(tokens []Token, pos int) ([]string, int) {
	if tokens[pos].Type != "lparen" {
		panic("expected a lparen for parsePortList")
	}
	ports, i := p.parsePorts(tokens, pos+1)
	if tokens[i].Type != "rparen" {
		panic("expected a rparen for parsePortList")
	}
	return ports, i + 1
}
func (p *Parser) parseModule(tokens []Token, pos int) (string, []string, int) {
	if tokens[pos].Type != "module" {
		panic("expected a module token for parseModule")
	}
	// extract name
	i := p.skip(tokens, p.skipTokens, pos+1) // skip over any skippable tokens
	if tokens[i].Type != "identifier" {
		panic("expected an identifier for parseModule")
	}
	name := tokens[i].Value

	// extract ports
	i = p.skip(tokens, p.skipTokens, i+1) // skip over any skippable
	ports := []string{}
	if tokens[i].Type == "lparen" {
		ports, i = p.parsePortList(tokens, i)
	}

	// check for semicolon
	i = p.skip(tokens, p.skipTokens, i) // skip over any skippable tokens
	if tokens[i].Type != "semicolon" {
		panic("expected a semicolon for parseModule")
	}
	i = p.skip(tokens, p.skipTokens, i+1) // skip over any skippable tokens

	// skip the interior
	i = p.skipModuleInterior(tokens, i)

	// double check for endmodule
	if tokens[i].Type != "endmodule" {
		panic("expected an endmodule for parseModule")
	}
	return name, ports, i + 1
}

func (p *Parser) isModule(tokens []Token, pos int) bool {
	if pos >= len(tokens) {
		return false
	}
	return tokens[pos].Type == "module"
}
func (p *Parser) isDirective(tokens []Token, pos int) bool {
	if pos >= len(tokens) {
		return false
	}
	return tokens[pos].Type == "include" || tokens[pos].Type == "timescale" || tokens[pos].Type == "define"
}
func (p *Parser) isSkippable(tokens []Token, pos int) bool {
	if pos >= len(tokens) {
		return false
	}
	skippable := false
	for j := 0; j < len(p.skipTokens); j++ {
		if tokens[pos].Type == p.skipTokens[j] {
			skippable = true
			break
		}
	}
	return skippable
}

func (p *Parser) Parse(tokens []Token) GlobalResults {
	modules := []Module{}
	defines := []string{}
	i := 0
	for i < len(tokens) {
		if p.isDirective(tokens, i) {
			if tokens[i].Type == "include" {
				i = p.skipInclude(tokens, i)
			} else if tokens[i].Type == "timescale" {
				i = p.skipTimescale(tokens, i)
			} else {
				name, i2 := p.parseDefine(tokens, i)
				i = i2
				defines = append(defines, name)
			}
		} else if p.isModule(tokens, i) {
			name, ports, i2 := p.parseModule(tokens, i)
			i = i2
			modules = append(modules, Module{
				Name:  name,
				Ports: ports,
			})
		} else {
			if !p.isSkippable(tokens, i) {
				if tokens[i].Type == "semicolon" {
					// exception for extraneous semicolons
					i = i + 1
				} else {
					panic("unexpected token on top level: " + tokens[i].Type)
				}
			}
			i = p.skip(tokens, p.skipTokens, i)
		}
	}
	return GlobalResults{
		Defines: defines,
		Modules: modules,
	}
}
