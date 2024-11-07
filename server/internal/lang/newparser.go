package lang

import (
	"fmt"
)

type FileNode struct {
	Statements []TopLevelStatement
}
type TopLevelStatement struct {
	Directive *DirectiveNode
	Module    *ModuleNode
}
type InteriorNode struct {
	DeclarationNode       *DeclarationNode
	AssignmentNode        *AssignmentNode
	ModuleApplicationNode *ModuleApplicationNode
}
type ModuleNode struct {
	Identifier Token        // name of module
	PortList   PortListNode // list of ports
	Interior   []InteriorNode
}
type PortListNode struct {
	Ports []Token // list of ports (identifiers)
}
type DefineNode struct {
	Identifier Token // name of the define
}
type DirectiveNode struct {
	DefineNode *DefineNode
}
type AssignmentNode struct {
	Identifier Token
	Index      *IndexNode
	Value      ValueNode
}
type IndexNode struct {
	Index Token // identifier, or literal
}
type SelectorNode struct {
	IndexNode *IndexNode
	RangeNode *RangeNode
}
type ValueNode struct {
	Identifiers []Token // technically, there could be literals too, but we don't care?
}
type DeclarationNode struct {
	Type      TypeNode
	Variables []VariableNode
	Value     *ValueNode
}
type VariableNode struct {
	Identifier Token
	Range      *RangeNode
}
type RangeNode struct {
	From Token
	To   Token
}
type TypeNode struct {
	Type        Token
	VectorRange *RangeNode
}
type ModuleApplicationNode struct {
	ModuleName Token  // name of the module
	GateName   *Token // name of this gate construct, could be nil
	Arguments  []ArgumentNode
}
type ArgumentNode struct {
	Label *Token    // label for argument name, could be nil
	Value ValueNode // value of the argument
}

func newErrorFrom(from string, expected []string, pos int, tokens []Token) error {
	return fmt.Errorf("parsing %s, expected %v, got: %v at position %d, remaining tokens: %v", from, expected, tokens[pos], pos, tokens[pos+1:])
}
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

// returned position is the position of the expected (or failed) token
func (p *Parser) checkToken(from string, expected []string, pos int, tokens []Token) (int, error) {
	// skip over any skippables
	pos = p.skip(tokens, p.skipTokens, pos)

	if pos >= len(tokens) {
		return -1, newErrorFrom(from, expected, len(tokens), append(tokens, Token{Type: "EOF"}))
	}

	for _, tp := range expected {
		if tp == tokens[pos].Type {
			return pos, nil
		}
	}
	return -1, newErrorFrom(from, expected, pos, tokens)
}

func (p *Parser) isEOF(tokens []Token, pos int) bool {
	pos = p.skip(tokens, p.skipTokens, pos)
	return pos >= len(tokens)
}

// ==============================
// Module Interior Section
// ==============================

// returned position is the position after the rbracket
func (p *Parser) ParseRangeNode(tokens []Token, pos int) (result RangeNode, newPos int, err error) {
	pos, err = p.checkToken("range node", []string{"lbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// now take the from
	pos, err = p.checkToken("range node", []string{"identifier", "literal"}, pos, tokens)
	if err != nil {
		return
	}
	result.From = tokens[pos]
	pos++

	// double check for colon
	pos, err = p.checkToken("range node", []string{"colon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// take the to
	pos, err = p.checkToken("range node", []string{"identifier", "literal"}, pos, tokens)
	if err != nil {
		return
	}
	result.To = tokens[pos]
	pos++

	// double check for rbracket
	pos, err = p.checkToken("range node", []string{"rbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

// returned position is the position after the rbracket
func (p *Parser) ParseIndexNode(tokens []Token, pos int) (result IndexNode, newPos int, err error) {
	pos, err = p.checkToken("index node", []string{"lbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	pos, err = p.checkToken("index node", []string{"identifier", "literal"}, pos, tokens)
	if err != nil {
		return
	}
	result.Index = tokens[pos]
	pos++
	pos, err = p.checkToken("index node", []string{"rbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

// <selector> -> LBRACKET <ident_or_int> [COLON <ident_or_int>] RBRACKET
// <ident_or_int> -> <identifier> | LITERAL
func (p *Parser) ParseSelectorNode(tokens []Token, pos int) (result SelectorNode, newPos int, err error) {
	// check for lbracket
	pos, err = p.checkToken("selector node", []string{"lbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	pos, err = p.checkToken("selector node", []string{"identifier", "literal"}, pos, tokens)
	if err != nil {
		return
	}
	firstToken := tokens[pos]
	pos++

	// check for colon
	potentialPos, e := p.checkToken("selector node", []string{"colon"}, pos, tokens)
	if e == nil {
		// had a colon, so extract the second
		pos = potentialPos
		pos++

		pos, err = p.checkToken("selector node", []string{"identifier", "literal"}, pos, tokens)
		if err != nil {
			return
		}
		secondToken := tokens[pos]
		pos++
		result = SelectorNode{RangeNode: &RangeNode{From: firstToken, To: secondToken}}
	} else {
		result = SelectorNode{IndexNode: &IndexNode{Index: firstToken}}
	}

	// check for rbracket
	pos, err = p.checkToken("selector node", []string{"rbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	newPos = pos
	return
}

// returned position is the position after the value node
func (p *Parser) ParseValueNode(tokens []Token, pos int) (result ValueNode, newPos int, err error) {
	pos, err = p.checkToken("value node", []string{"identifier", "literal", "lcurl"}, pos, tokens)
	if err != nil {
		return
	}
	// now, determine what it is
	if tokens[pos].Type == "identifier" {
		result.Identifiers = []Token{tokens[pos]}
		pos++
		_, potentialPos, e := p.ParseSelectorNode(tokens, pos)
		if e == nil {
			// there was a selector
			// TODO - maybe store it?
			pos = potentialPos
		}
	} else if tokens[pos].Type == "literal" {
		// literal, so no need to store it
		// result.Identifiers = []Token{tokens[pos]}
		pos++
		newPos = pos

		// this literal might just be a size
		potentialPos, e := p.checkToken("value node", []string{"lcurl"}, pos, tokens)
		if e == nil {
			// this was just the size, so parse the inner value
			// (copied code from the lcurl case below)
			pos = potentialPos
			pos++
			inner, potentialPos, e := p.ParseValueNode(tokens, pos)
			if e == nil {
				// success!
				result.Identifiers = inner.Identifiers
				pos = potentialPos
			}
			// now, find the rest of the subparts
			potentialPos, e = p.checkToken("value node", []string{"comma"}, pos, tokens)
			for e == nil {
				inner, potentialPos, e = p.ParseValueNode(tokens, potentialPos+1) // potentialPos was position of comma
				if e == nil {
					result.Identifiers = append(result.Identifiers, inner.Identifiers...)
					pos = potentialPos
				} else {
					// TODO - something went wrong, should we stop here or try to keep going?
				}
				potentialPos, e = p.checkToken("value node", []string{"comma"}, pos, tokens)
			}
			// finally, the rcurl
			pos, err = p.checkToken("value node", []string{"rcurl"}, pos, tokens)
			if err != nil {
				return
			}
			pos++
			newPos = pos
			return
		}

		// might also be a selector
		_, potentialPos, e = p.ParseSelectorNode(tokens, pos)
		if e == nil {
			// there was a selector
			// TODO - maybe store it?
			pos = potentialPos
		}
	} else if tokens[pos].Type == "lcurl" {
		// some sort of concatenation of things...
		// TODO - maybe store it?
		// for now, just find all the identifiers in the subpart
		pos++
		inner, potentialPos, e := p.ParseValueNode(tokens, pos)
		if e == nil {
			// success!
			result.Identifiers = inner.Identifiers
			pos = potentialPos
		}
		// now, find the rest of the subparts
		potentialPos, e = p.checkToken("value node", []string{"comma"}, pos, tokens)
		for e == nil {
			inner, potentialPos, e = p.ParseValueNode(tokens, potentialPos+1) // potentialPos was position of comma
			if e == nil {
				result.Identifiers = append(result.Identifiers, inner.Identifiers...)
				pos = potentialPos
			} else {
				// TODO - something went wrong, should we stop here or try to keep going?
			}
			potentialPos, e = p.checkToken("value node", []string{"comma"}, pos, tokens)
		}
		// finally, the rcurl
		pos, err = p.checkToken("value node", []string{"rcurl"}, pos, tokens)
		if err != nil {
			return
		}
		pos++
	}

	newPos = pos
	return
}

func (p *Parser) ParseArgument(tokens []Token, pos int) (result ArgumentNode, newPos int, err error) {
	// dot for named parameter, identifier/lcurcly/literal for value
	pos, err = p.checkToken("argument", []string{"dot", "identifier", "lcurl", "literal"}, pos, tokens)
	if err != nil {
		return
	}
	if tokens[pos].Type == "dot" {
		// named parameter
		pos++
		pos, err = p.checkToken("argument", []string{"identifier"}, pos, tokens)
		if err != nil {
			return
		}
		result.Label = &tokens[pos] // store parameter name
		pos++

		// check for lparen
		pos, err = p.checkToken("argument", []string{"lparen"}, pos, tokens)
		if err != nil {
			return
		}
		pos++

		// get value
		value, potentialPos, e := p.ParseValueNode(tokens, pos)
		if e == nil {
			// success!
			result.Value = value
			pos = potentialPos
		} else {
			// something went wrong, for now just pretend everything's ok
			// TODO - error handling
		}

		// check for rparen
		pos, err = p.checkToken("argument", []string{"rparen"}, pos, tokens)
		if err != nil {
			return
		}
		pos++
		newPos = pos
		return
	} else {
		// just a value
		value, potentialPos, e := p.ParseValueNode(tokens, pos)
		if e == nil {
			// success!
			result.Value = value
			pos = potentialPos
		} else {
			// something went wrong, for now just pretend everything's ok
			// TODO - error handling
		}
		newPos = pos
		return
	}
}
func (p *Parser) ParseArguments(tokens []Token, pos int) (result []ArgumentNode, newPos int, err error) {
	// arguments is just a bunch of argument nodes
	// so we take our first argument and then take any extras separated by commas

	// get the first argument
	argument, potentialPos, e := p.ParseArgument(tokens, pos)
	if e == nil {
		// success!
		result = append(result, argument)
		pos = potentialPos
	}

	// now take the rest
	potentialPos, e = p.checkToken("arguments", []string{"comma"}, pos, tokens)
	for e == nil {
		argument, potentialPos, e = p.ParseArgument(tokens, potentialPos+1) // potentialPos was position of comma
		if e == nil {
			// success!
			result = append(result, argument)
			pos = potentialPos
		} else {
			// TODO - something went wrong, should we stop here or try to keep going?
		}
		potentialPos, e = p.checkToken("arguments", []string{"comma"}, pos, tokens)
	}
	newPos = pos
	return
}
func (p *Parser) ParseModuleApplication(tokens []Token, pos int) (result ModuleApplicationNode, newPos int, err error) {
	// module name
	pos, err = p.checkToken("module application", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.ModuleName = tokens[pos]
	pos++

	// might have a gate name
	potentialPos, e := p.checkToken("module application", []string{"identifier"}, pos, tokens)
	if e == nil {
		result.GateName = &tokens[potentialPos]
		pos = potentialPos + 1
	}

	// check for lparen
	pos, err = p.checkToken("module application", []string{"lparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get arguments
	arguments, potentialPos, e := p.ParseArguments(tokens, pos)
	if e == nil {
		// success!
		result.Arguments = arguments
		pos = potentialPos
	} else {
		// something went wrong, for now just pretend everything's ok
		// TODO - error handling
	}

	// check for rparen
	pos, err = p.checkToken("module application", []string{"rparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// check for semicolon
	pos, err = p.checkToken("module application", []string{"semicolon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

func (p *Parser) ParseVariableNode(tokens []Token, pos int) (result VariableNode, newPos int, err error) {
	// identifier optionally followed by a range
	pos, err = p.checkToken("variable", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Identifier = tokens[pos]
	pos++

	// now try taking the range; it's ok if it fails since it's optional
	rangeNode, potentialPos, e := p.ParseRangeNode(tokens, pos)
	if e == nil {
		// success!
		result.Range = &rangeNode
		pos = potentialPos
	}
	newPos = pos
	return
}
func (p *Parser) ParseAssignmentNode(tokens []Token, pos int) (result AssignmentNode, newPos int, err error) {
	// <identifier> [<index>] EQUALS <value> SEMICOLON
	pos, err = p.checkToken("assignment", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Identifier = tokens[pos]
	pos++

	// now try taking the index; it's ok if it fails since it's optional
	indexNode, potentialPos, e := p.ParseIndexNode(tokens, pos)
	if e == nil {
		// success!
		result.Index = &indexNode
		pos = potentialPos
	}

	// check for equals
	pos, err = p.checkToken("assignment", []string{"equals"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get the value
	valueNode, potentialPos, e := p.ParseValueNode(tokens, pos)
	if e == nil {
		// success!
		result.Value = valueNode
		pos = potentialPos
	} else {
		// this is really an error since the value is required
		// TODO - error handling
	}

	// check for semicolon
	pos, err = p.checkToken("assignment", []string{"semicolon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

func (p *Parser) ParseTypeNode(tokens []Token, pos int) (result TypeNode, newPos int, err error) {
	// TYPE [<range>]
	pos, err = p.checkToken("type", []string{"type"}, pos, tokens)
	if err != nil {
		return
	}
	result.Type = tokens[pos]
	pos++

	// now try taking the range; it's ok if it fails since it's optional
	rangeNode, potentialPos, e := p.ParseRangeNode(tokens, pos)
	if e == nil {
		// success!
		result.VectorRange = &rangeNode
		pos = potentialPos
	}
	newPos = pos
	return
}

func (p *Parser) ParseDeclarationNode(tokens []Token, pos int) (result DeclarationNode, newPos int, err error) {
	//<declaration> -> <type> <single_var> [EQUALS <value>] SEMICOLON
	//    | <type> <single_var> { COMMA <single_var> } SEMICOLON
	typeNode, potentialPos, e := p.ParseTypeNode(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Type = typeNode
	pos = potentialPos

	// get the variable
	variableNode, potentialPos, e := p.ParseVariableNode(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Variables = []VariableNode{variableNode}
	pos = potentialPos

	// see if it's an equals
	potentialPos, e = p.checkToken("declaration", []string{"equals"}, pos, tokens)
	if e == nil {
		// it's an equals, so there's a value
		valueNode, potentialPos, e := p.ParseValueNode(tokens, potentialPos+1)
		if e == nil {
			// success!
			result.Value = &valueNode
			pos = potentialPos
		} else {
			err = e
			return
		}
	} else {
		// see if it's a comma
		potentialPos, e = p.checkToken("declaration", []string{"comma"}, pos, tokens)
		if e == nil {
			pos = potentialPos
			// it's a comma, so there's more variables
			for e == nil {
				variableNode, potentialPos, e = p.ParseVariableNode(tokens, potentialPos+1)
				if e != nil {
					err = e
					return
				}
				result.Variables = append(result.Variables, variableNode)
				pos = potentialPos
				potentialPos, e = p.checkToken("declaration", []string{"comma"}, pos, tokens)
			}
		}
	}

	// check for semicolon
	pos, err = p.checkToken("declaration", []string{"semicolon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

func (p *Parser) ParseInteriorStatement(tokens []Token, pos int) (result InteriorNode, newPos int, err error) {
	// it could be either a declaration or module_application or assignment

	// check if it's a declaration
	declarationNode, potentialPos, e := p.ParseDeclarationNode(tokens, pos)
	if e == nil {
		// success!
		result.DeclarationNode = &declarationNode
		pos = potentialPos
	} else {
		// check if it's a module application
		moduleApplicationNode, potentialPos, e := p.ParseModuleApplication(tokens, pos)
		if e == nil {
			// success!
			result.ModuleApplicationNode = &moduleApplicationNode
			pos = potentialPos
		} else {
			// check if it's an assignment
			assignmentNode, potentialPos, e := p.ParseAssignmentNode(tokens, pos)
			if e == nil {
				// success!
				result.AssignmentNode = &assignmentNode
				pos = potentialPos
			} else {
				err = e
				return
			}
		}
	}

	newPos = pos
	return
}
func (p *Parser) ParseModuleInterior(tokens []Token, pos int) (result []InteriorNode, newPos int, err error) {
	for {
		nestedStatement, potentialPos, e := p.ParseInteriorStatement(tokens, pos)
		if e != nil {
			return
		}
		result = append(result, nestedStatement)
		pos = potentialPos
		newPos = pos
	}
}

// ==============================
// Module Definition Section
// ==============================

// Returns a list of ports, and newPos is the position after the list
func (p *Parser) ParsePorts(tokens []Token, pos int) (result []Token, newPos int, err error) {
	// ports is just the list of identifiers

	// get the first identifier
	pos, err = p.checkToken("ports", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result = append(result, tokens[pos])
	pos++

	// now take the rest
	potentialPos, e := p.checkToken("ports", []string{"comma"}, pos, tokens)
	for e == nil {
		pos, err = p.checkToken("ports", []string{"identifier"}, potentialPos+1, tokens)
		if err != nil {
			return
		}
		result = append(result, tokens[pos])
		pos++
		potentialPos, e = p.checkToken("ports", []string{"comma"}, pos, tokens)
	}
	newPos = pos
	return
}

// Returns a list of ports, and newPos is the position after the list
func (p *Parser) ParsePortList(tokens []Token, pos int) (result PortListNode, newPos int, err error) {
	// take lparen
	pos, err = p.checkToken("port list", []string{"lparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get ports if any
	ports, potentialPos, e := p.ParsePorts(tokens, pos)
	if e == nil {
		// got ports successfully
		result.Ports = ports
		pos = potentialPos
	}

	// then get the rparen
	pos, err = p.checkToken("port list", []string{"rparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	newPos = pos
	return
}

func (p *Parser) ParseModule(tokens []Token, pos int) (result ModuleNode, newPos int, err error) {
	// MODULE <identifier> [<port_list>] SEMICOLON <interior> ENDMODULE [SEMICOLON]
	pos, err = p.checkToken("module", []string{"module"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get the identifier
	pos, err = p.checkToken("module", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Identifier = tokens[pos]
	pos++

	// get the port list if any
	portList, potentialPos, e := p.ParsePortList(tokens, pos)
	if e == nil {
		// success!
		result.PortList = portList
		pos = potentialPos
	}

	// get the semicolon
	pos, err = p.checkToken("module", []string{"semicolon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get the interior
	interior, potentialPos, e := p.ParseModuleInterior(tokens, pos)
	if e == nil {
		// success!
		result.Interior = interior
		pos = potentialPos
	} else {
		err = e
		return
	}

	// get the endmodule
	pos, err = p.checkToken("module", []string{"endmodule"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get the semicolon, optionally
	potentialPos, e = p.checkToken("module", []string{"semicolon"}, pos, tokens)
	if e == nil {
		pos = potentialPos
	}
	pos++
	newPos = pos
	return
}

// ==============================
// Directive Section
// ==============================
func (p *Parser) ParseDefine(tokens []Token, pos int) (result *DefineNode, newPos int, err error) {
	// DEFINE <identifier> ... NEWLINE
	pos, err = p.checkToken("define", []string{"define"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get the identifier
	pos, err = p.checkToken("define", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result = &DefineNode{Identifier: tokens[pos]}
	pos++

	// skip to newline
	for tokens[pos].Type != "newline" {
		pos++
	}
	newPos = pos
	return
}
func (p *Parser) SkipTimescale(tokens []Token, pos int) (newPos int, err error) {
	// TIMESCALE
	pos, err = p.checkToken("timescale", []string{"timescale"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// skip to newline
	for tokens[pos].Type != "newline" {
		pos++
	}

	newPos = pos
	return
}
func (p *Parser) SkipInclude(tokens []Token, pos int) (newPos int, err error) {
	// INCLUDE
	pos, err = p.checkToken("include", []string{"include"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}
func (p *Parser) ParseDirective(tokens []Token, pos int) (result *DefineNode, newPos int, err error) {
	// directive is just a define, timescale, or include
	result, newPos, err = p.ParseDefine(tokens, pos)
	if err == nil {
		return // success with define
	}

	newPos, err = p.SkipTimescale(tokens, pos)
	if err == nil {
		return // success with timescale
	}

	newPos, err = p.SkipInclude(tokens, pos)
	if err == nil {
		return // success with include
	}
	return // failure with all three
}

func (p *Parser) ParseFile(tokens []Token) (result FileNode, err error) {
	pos := 0

	for !p.isEOF(tokens, pos) {
		// it's either a directive or a module
		// try directive
		directive, newPos, e := p.ParseDirective(tokens, pos)
		if e != nil {
			// try module
			module, newPos, e := p.ParseModule(tokens, pos)
			if e != nil {
				err = e
				return
			}
			result.Statements = append(result.Statements, TopLevelStatement{
				Module: &module})
			pos = newPos
		} else {
			result.Statements = append(result.Statements, TopLevelStatement{
				Directive: &DirectiveNode{DefineNode: directive},
			})
			pos = newPos
		}
	}

	return
}

/**
Grammar:

// top level grammar
<file> -> <statement> { <statement> }
<statement> -> <module> | <directive> | <skippable>

// sub grammars
<skippable> -> WHITESPACE | COMMENT | NEWLINE

<directive> -> <include> | <timescale> | <define>
<include> -> INCLUDE
<timescale> -> TIMESCALE <non-newline> NEWLINE
<define> -> DEFINE <identifier> <non-newline> NEWLINE
<identifier> -> IDENTIFIER


<module> -> MODULE <identifier> [<portlist>] SEMICOLON <interior> ENDMODULE [SEMICOLON]
<portlist> -> LPAREN [<ports>] RPAREN
<ports> -> <identifier> { COMMA <identifier> }

<interior> -> { <interior_statement> }
<interior_statement>  -> <declaration> | <module_application> | <assignment>

<assignment> -> <identifier> [<index>] EQUALS <value> SEMICOLON
<single_var> -> <identifier> [<range>]

<declaration> -> <type> <single_var> [EQUALS <value>] SEMICOLON
| <type> <single_var> { COMMA <single_var> } SEMICOLON
<type> -> TYPE [<range>]
<index> -> LBRACKET <identifier> RBRACKET | LBRACKET <integer> RBRACKET
<range> -> LBRACKET <integer> COLON <integer> RBRACKET
<integer> -> LITERAL | DEFINE

<module_application> -> <identifier> [<identifier>] LPAREN <arguments> RPAREN SEMICOLON
<arguments> -> <argument> { COMMA <argument> }
<argument> -> DOT <identifier> LPAREN  <value>  RPAREN | <value>
<value> -> [LITERAL] LCURL <value> {COMMA <value>} RCURL | LITERAL [<selector>] | <identifier> [<selector]
<selector> -> LBRACKET <ident_or_int> [COLON <ident_or_int>] RBRACKET
<ident_or_int> -> <identifier> | LITERAL
*/
