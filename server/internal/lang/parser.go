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
	GenerateNode          *GenerateNode
	AlwaysNode            *AlwaysNode
	DefParamNode          *DefParamNode
	InitialNode           *InitialNode
	DirectiveNode         *DefineNode
	TaskNode              *TaskNode
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
	Identifier      Token
	Index           *IndexNode
	Value           ExprNode
	IsAssign        bool
	IsDelayedAssign bool // true if used <= instead of =
}
type IndexNode struct {
	Index ExprNode
}
type SelectorNode struct {
	IndexNode *IndexNode
	RangeNode *RangeNode
}
type ValueNode struct {
	Value     []Token
	Selectors []SelectorNode
}
type SizedValueNode struct {
	Size   *Token
	Values []ValueNode
}
type DeclarationNode struct {
	Type      TypeNode
	Variables []VariableNode
	Values    []ExprNode
}
type VariableNode struct {
	Identifier Token
	Ranges     []RangeNode
}
type RangeNode struct {
	From ExprNode
	To   ExprNode
}
type TypeNode struct {
	Type   Token
	Ranges []RangeNode
}
type ModuleApplicationNode struct {
	ModuleName Token  // name of the module
	GateName   *Token // name of this gate construct, could be nil
	Range      *RangeNode
	Arguments  []ArgumentNode
}
type ArgumentNode struct {
	Label *Token   // label for argument name, could be nil
	Value ExprNode // value of the argument
}
type ExprNode struct {
	Value      SizedValueNode
	Combinator *Token
	Right      *ExprNode
	ExprTrue   *ExprNode
	ExprFalse  *ExprNode
}
type GenerateNode struct {
	Statements []AlwaysStatement
}
type BeginBlockNode struct {
	Statements []AlwaysStatement
}
type ForBlockNode struct {
	Initializer *AssignmentNode
	Condition   *ExprNode
	Incrementor *AssignmentNode
	Body        AlwaysStatement
}
type IfBlockNode struct {
	Expr ExprNode
	Body AlwaysStatement
	Else *AlwaysStatement
}
type AlwaysNode struct {
	Times     []TimeNode
	Statement AlwaysStatement
}
type AlwaysStatement struct {
	DelayNode    *DelayNode
	BeginBlock   *BeginBlockNode
	ForBlock     *ForBlockNode
	IfBlock      *IfBlockNode
	InteriorNode *InteriorNode
	FunctionNode *FunctionNode
	CaseNode     *CaseBlock
}
type TimeNode struct {
	Time       *Token // negedge, posedge, or nil
	Identifier Token
}
type DelayNode struct {
	Amount Token
}
type FunctionNode struct {
	Function    Token
	Expressions []ExprNode
}
type DefParamNode struct {
	Identifiers []Token
	Value       ExprNode
}
type InitialNode struct {
	Statement AlwaysStatement
}
type CaseBlock struct {
	Expr    ExprNode
	Cases   []CaseNode
	Default *AlwaysStatement
}
type CaseNode struct {
	Conditions []ExprNode
	Statement  AlwaysStatement
}
type TaskNode struct {
	Identifier Token
	Statements []TaskStatement
}
type TaskStatement struct {
	InteriorNode *InteriorNode
	BeginBlock   *BeginBlockNode
}
type Parser struct {
	skipTokens            []string
	FarthestErrorPosition int
	FarthestError         *error
}

func NewParser() *Parser {
	return &Parser{
		skipTokens:            []string{"whitespace", "comment", "newline"},
		FarthestErrorPosition: -1,
		FarthestError:         nil,
	}
}

func (p *Parser) newErrorFrom(from string, expected []string, pos int, tokens []Token) error {
	err := fmt.Errorf("parsing %s, expected %v, got: %v at position %d", from, expected, tokens[pos], pos)
	if pos > p.FarthestErrorPosition {
		p.FarthestErrorPosition = pos
		p.FarthestError = &err
	}
	return err
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
		return -1, p.newErrorFrom(from, expected, len(tokens), append(tokens, Token{Type: "EOF"}))
	}

	for _, tp := range expected {
		if tp == tokens[pos].Type {
			return pos, nil
		}
	}
	return -1, p.newErrorFrom(from, expected, pos, tokens)
}

func (p *Parser) isEOF(tokens []Token, pos int) bool {
	pos = p.skip(tokens, p.skipTokens, pos)
	return pos >= len(tokens)
}

// ==============================
// Module Interior Section
// ==============================

// returned position is the position after the rbracket
func (p *Parser) parseRangeNode(tokens []Token, pos int) (result RangeNode, newPos int, err error) {
	pos, err = p.checkToken("range node", []string{"lbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// now take the from
	fromNode, pos, err := p.parseExpression(tokens, pos)
	if err != nil {
		return
	}
	result.From = fromNode

	// double check for colon
	pos, err = p.checkToken("range node", []string{"colon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// take the to
	toNode, pos, err := p.parseExpression(tokens, pos)
	if err != nil {
		return
	}
	result.To = toNode

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
func (p *Parser) parseIndexNode(tokens []Token, pos int) (result IndexNode, newPos int, err error) {
	pos, err = p.checkToken("index node", []string{"lbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	indexNode, pos, err := p.parseExpression(tokens, pos)
	if err != nil {
		return
	}
	result.Index = indexNode

	pos, err = p.checkToken("index node", []string{"rbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

// <selector> -> LBRACKET <expr> [COLON <expr>] RBRACKET
func (p *Parser) parseSelectorNode(tokens []Token, pos int) (result SelectorNode, newPos int, err error) {
	// check for lbracket
	pos, err = p.checkToken("selector node", []string{"lbracket"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	firstNode, pos, err := p.parseExpression(tokens, pos)
	if err != nil {
		return
	}

	// check for colon
	potentialPos, e := p.checkToken("selector node", []string{"colon"}, pos, tokens)
	if e == nil {
		// had a colon, so extract the second
		pos = potentialPos
		pos++

		secondNode, potentialPos, e := p.parseExpression(tokens, pos)
		if e != nil {
			err = e
			return
		}
		pos = potentialPos
		result = SelectorNode{RangeNode: &RangeNode{From: firstNode, To: secondNode}}
	} else {
		result = SelectorNode{IndexNode: &IndexNode{Index: firstNode}}
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

// <sized_value> -> [ LITERAL ] LCURL <sized_value> { COMMA <sized_value> } RCURL | <value>
func (p *Parser) parseSized(tokens []Token, pos int) (result SizedValueNode, newPos int, err error) {
	// seems to be a sized value
	potentialPos, e := p.checkToken("sized value", []string{"literal", "identifier"}, pos, tokens)
	if e == nil {
		// there was a size
		result.Size = &tokens[potentialPos]
		pos = potentialPos + 1
	}

	// now take the lcurl
	pos, err = p.checkToken("sized value", []string{"lcurl"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// now take the values
	sizedNode, potentialPos, e := p.parseSizedValueNode(tokens, pos)
	if e != nil {
		// there must be at least one value
		err = e
		return
	}

	// take the rest
	for e == nil {
		result.Values = append(result.Values, sizedNode.Values...)
		pos = potentialPos

		potentialPos, e = p.checkToken("sized value", []string{"comma"}, pos, tokens)
		if e != nil {
			break
		}
		sizedNode, potentialPos, e = p.parseSizedValueNode(tokens, potentialPos+1)
	}

	// take the rcurl
	pos, err = p.checkToken("sized value", []string{"rcurl"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

func (p *Parser) parseSizedValueNode(tokens []Token, pos int) (result SizedValueNode, newPos int, err error) {
	// try just taking a value
	sizedNode, potentialPos, e := p.parseSized(tokens, pos)
	if e == nil {
		// it was sized
		result = sizedNode
		pos = potentialPos
	} else {
		// it wasn't
		valueNode, potentialPos, e := p.parseValueNode(tokens, pos)
		if e != nil {
			err = e
			return
		}
		result.Values = []ValueNode{valueNode}
		pos = potentialPos
	}
	newPos = pos
	return
}

// <maybe_signed> -> <sized_value> | SIGNED LPAREN <sized_value> RPAREN
func (p *Parser) parseSigned(tokens []Token, pos int) (result SizedValueNode, newPos int, err error) {
	potentialPos, e := p.checkToken("signed", []string{"signed"}, pos, tokens)
	if e == nil {
		// it was signed
		pos = potentialPos + 1

		// take lparen
		pos, err = p.checkToken("signed", []string{"lparen"}, pos, tokens)
		if err != nil {
			return
		}
		pos++

		// take sized value
		result, pos, err = p.parseSizedValueNode(tokens, pos)
		if err != nil {
			return
		}

		// take rparen
		pos, err = p.checkToken("signed", []string{"rparen"}, pos, tokens)
		if err != nil {
			return
		}
		pos++

	} else {
		// just a regular sized value
		result, pos, err = p.parseSizedValueNode(tokens, pos)
	}
	newPos = pos
	return
}

// returned position is the position after the value node
func (p *Parser) parseValueNode(tokens []Token, pos int) (result ValueNode, newPos int, err error) {
	// <value> -> [TILDE| - ] (LITERAL|(<identifier> { DOT <identifier> })|FUNCLITERAL) { <selector> }

	// get optional tilde or minus
	potentialPos, e := p.checkToken("value node", []string{"tilde", "operator"}, pos, tokens)
	if e == nil {
		// there was some sort of unary operator
		if tokens[potentialPos].Type == "operator" && tokens[potentialPos].Value != "-" {
			err = fmt.Errorf("expected tilde or minus but got %s", tokens[potentialPos].Value)
			return
		}
		pos = potentialPos + 1
	}

	pos, err = p.checkToken("value node", []string{"identifier", "literal", "funcliteral"}, pos, tokens)
	if err != nil {
		return
	}
	// take the value
	result.Value = append(result.Value, tokens[pos])
	if tokens[pos].Type == "identifier" {
		pos++
		// potentially take the next identifiers
		potentialPos, e = p.checkToken("value node", []string{"dot"}, pos, tokens)
		for e == nil {
			// take the identifier
			pos, err = p.checkToken("value node", []string{"identifier"}, potentialPos+1, tokens)
			if err != nil {
				return
			}
			result.Value = append(result.Value, tokens[pos])
			pos++
			potentialPos, e = p.checkToken("value node", []string{"dot"}, pos, tokens)
		}
	} else {
		pos++
	}

	// and any selectors
	selectorNode, potentialPos, e := p.parseSelectorNode(tokens, pos)
	for e == nil {
		pos = potentialPos
		result.Selectors = append(result.Selectors, selectorNode)
		selectorNode, potentialPos, e = p.parseSelectorNode(tokens, pos)
	}

	newPos = pos
	return
}

func (p *Parser) parseExpression(tokens []Token, pos int) (result ExprNode, newPos int, err error) {
	// <expr> -> (<value> | LPAREN <expr> RPAREN) [(OPERATOR|COMPARATOR) <expr>]  [ QUESTION <expr> COLON <expr> ]

	potentialPos, e := p.checkToken("expression", []string{"lparen"}, pos, tokens)
	if e == nil {
		// nested expression
		pos = potentialPos + 1
		result, pos, err = p.parseExpression(tokens, pos)
		if err != nil {
			return
		}
		// check for rparen
		pos, err = p.checkToken("expression", []string{"rparen"}, pos, tokens)
		if err != nil {
			return
		}
		pos++
	} else {
		// just a value
		value, potentialPos, e := p.parseSigned(tokens, pos)
		if e != nil {
			err = e
			return
		}
		pos = potentialPos
		result.Value = value
	}

	// check for operator
	potentialPos, e = p.checkToken("expression", []string{"operator", "comparator"}, pos, tokens)
	if e == nil {
		// had an operator or comparator
		pos = potentialPos
		result.Combinator = &tokens[pos]
		pos++

		// get the right expression
		tmp, potentialPos, e := p.parseExpression(tokens, pos)
		if e != nil {
			err = e
			return
		}
		result.Right = &tmp
		pos = potentialPos
	}

	// check for ternary
	potentialPos, e = p.checkToken("expression", []string{"question"}, pos, tokens)
	if e == nil {
		// get the true expression
		pos = potentialPos + 1
		tmp, potentialPos, e := p.parseExpression(tokens, pos)
		if e != nil {
			err = e
			return
		}
		result.ExprTrue = &tmp
		pos = potentialPos

		// need a colon
		pos, err = p.checkToken("expression", []string{"colon"}, pos, tokens)
		if err != nil {
			return
		}
		pos++

		// get the false expression
		tmp, potentialPos, e = p.parseExpression(tokens, pos)
		if e != nil {
			err = e
			return
		}
		result.ExprFalse = &tmp
		pos = potentialPos
	}

	newPos = pos
	return
}

func (p *Parser) parseArgument(tokens []Token, pos int) (result ArgumentNode, newPos int, err error) {
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
		value, potentialPos, e := p.parseExpression(tokens, pos)
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
		value, potentialPos, e := p.parseExpression(tokens, pos)
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
func (p *Parser) parseArguments(tokens []Token, pos int) (result []ArgumentNode, newPos int, err error) {
	// arguments is just a bunch of argument nodes
	// so we take our first argument and then take any extras separated by commas

	// get the first argument
	argument, potentialPos, e := p.parseArgument(tokens, pos)
	if e == nil {
		// success!
		result = append(result, argument)
		pos = potentialPos
	}

	// now take the rest
	potentialPos, e = p.checkToken("arguments", []string{"comma"}, pos, tokens)
	for e == nil {
		argument, potentialPos, e = p.parseArgument(tokens, potentialPos+1) // potentialPos was position of comma
		if e == nil {
			// success!
			result = append(result, argument)
			pos = potentialPos
		} else {
			// something went wrong, return an error
			err = e
			return
		}
		potentialPos, e = p.checkToken("arguments", []string{"comma"}, pos, tokens)
	}
	newPos = pos
	return
}
func (p *Parser) parseModuleApplication(tokens []Token, pos int) (result ModuleApplicationNode, newPos int, err error) {
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

	// might have a gate range
	rangeNode, potentialPos, e := p.parseRangeNode(tokens, pos)
	if e == nil {
		result.Range = &rangeNode
		pos = potentialPos
	}

	// check for lparen
	pos, err = p.checkToken("module application", []string{"lparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get arguments
	arguments, potentialPos, e := p.parseArguments(tokens, pos)
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

func (p *Parser) parseVariableNode(tokens []Token, pos int) (result VariableNode, newPos int, err error) {
	// identifier optionally followed by a range
	pos, err = p.checkToken("variable", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Identifier = tokens[pos]
	pos++

	// now try taking the range; it's ok if it fails since it's optional
	rangeNode, potentialPos, e := p.parseRangeNode(tokens, pos)
	for e == nil {
		result.Ranges = append(result.Ranges, rangeNode)
		pos = potentialPos
		rangeNode, potentialPos, e = p.parseRangeNode(tokens, pos) // try taking the next range
	}
	newPos = pos
	return
}
func (p *Parser) parseAssignmentNodeWithoutSemicolon(tokens []Token, pos int) (result AssignmentNode, newPos int, err error) {
	// [ASSIGN] <identifier> [<index>] (EQUAL|<=) <expr>
	// TODO - get the expr, not just a value
	potentialPos, e := p.checkToken("assignment", []string{"assign"}, pos, tokens)
	// it's ok if it fails since it's optional
	if e == nil {
		pos = potentialPos + 1
		result.IsAssign = true
	}

	pos, err = p.checkToken("assignment", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Identifier = tokens[pos]
	pos++

	// now try taking the index; it's ok if it fails since it's optional
	indexNode, potentialPos, e := p.parseIndexNode(tokens, pos)
	if e == nil {
		// success!
		result.Index = &indexNode
		pos = potentialPos
	}

	// check for equal
	pos, err = p.checkToken("assignment", []string{"equal", "comparator"}, pos, tokens)
	if err != nil {
		return
	} else if tokens[pos].Type == "comparator" {
		if tokens[pos].Value != "<=" {
			err = fmt.Errorf("expected = or <=, got %s", tokens[pos].Value)
			return
		}
		result.IsDelayedAssign = true
	}
	pos++

	// get the value
	valueNode, potentialPos, e := p.parseExpression(tokens, pos)
	if e == nil {
		// success!
		result.Value = valueNode
		pos = potentialPos
	} else {
		// this is really an error since the value is required
		// TODO - error handling
	}

	newPos = pos
	return
}
func (p *Parser) parseAssignmentNode(tokens []Token, pos int) (result AssignmentNode, newPos int, err error) {
	// <assignment> SEMICOLON
	result, pos, err = p.parseAssignmentNodeWithoutSemicolon(tokens, pos)
	if err != nil {
		return
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

func (p *Parser) parseTypeNode(tokens []Token, pos int) (result TypeNode, newPos int, err error) {
	// (TYPE | DIRECTION [TYPE]) [<range>]
	pos, err = p.checkToken("type", []string{"type", "direction"}, pos, tokens)
	if err != nil {
		return
	}

	if tokens[pos].Type == "direction" {
		// potentially take a type
		potentialPos, e := p.checkToken("type", []string{"type"}, pos+1, tokens)
		if e == nil {
			// success!
			result.Type = tokens[potentialPos]
			pos = potentialPos + 1
		} else {
			result.Type = tokens[pos]
			pos++
		}
	} else {
		result.Type = tokens[pos]
		pos++
	}

	// now try taking the range; it's ok if it fails since it's optional
	rangeNode, potentialPos, e := p.parseRangeNode(tokens, pos)
	for e == nil {
		result.Ranges = append(result.Ranges, rangeNode)
		pos = potentialPos
		rangeNode, potentialPos, e = p.parseRangeNode(tokens, pos)
	}
	newPos = pos
	return
}

func (p *Parser) parseDeclarationNode(tokens []Token, pos int) (result DeclarationNode, newPos int, err error) {
	// <declaration> -> <type> <single_var> EQUAL <expr> { COMMA <single_var> EQUAL <expr> } SEMICOLON
	// | <type> <single_var> { COMMA <single_var> } SEMICOLON

	typeNode, potentialPos, e := p.parseTypeNode(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Type = typeNode
	pos = potentialPos

	// get the variable
	variableNode, potentialPos, e := p.parseVariableNode(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Variables = []VariableNode{variableNode}
	pos = potentialPos

	// see if it's an equal
	potentialPos, e = p.checkToken("declaration", []string{"equal"}, pos, tokens)
	if e == nil {
		// it's an equal, so there's a value
		valueNode, potentialPos, e := p.parseExpression(tokens, potentialPos+1)
		if e == nil {
			result.Values = append(result.Values, valueNode)
			pos = potentialPos
		} else {
			err = e
			return
		}

		// see if there's more variables
		potentialPos, e = p.checkToken("declaration", []string{"comma"}, pos, tokens)
		if e == nil {
			pos = potentialPos
			// it's a comma, so there's more variables
			for e == nil {
				// get var
				variableNode, pos, err = p.parseVariableNode(tokens, potentialPos+1)
				if err != nil {
					return
				}
				result.Variables = append(result.Variables, variableNode)

				// get equal
				pos, err = p.checkToken("declaration", []string{"equal"}, pos, tokens)
				if err != nil {
					return
				}
				pos++

				// get value
				valueNode, pos, err = p.parseExpression(tokens, pos)
				if err != nil {
					return
				}
				result.Values = append(result.Values, valueNode)

				// possibly continue
				potentialPos, e = p.checkToken("declaration", []string{"comma"}, pos, tokens)
			}
		}

	} else {
		// see if there's more variables
		potentialPos, e = p.checkToken("declaration", []string{"comma"}, pos, tokens)
		if e == nil {
			pos = potentialPos
			// it's a comma, so there's more variables
			for e == nil {
				variableNode, potentialPos, e = p.parseVariableNode(tokens, potentialPos+1)
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

func (p *Parser) parseBeginBlock(tokens []Token, pos int) (result BeginBlockNode, newPos int, err error) {
	// BEGIN [ COLON <identifier> ] { <alwaysable_statement> } END
	pos, err = p.checkToken("begin block", []string{"begin"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// optionally, get colon
	potentialPos, e := p.checkToken("begin block", []string{"colon"}, pos, tokens)
	if e == nil {
		pos = potentialPos + 1

		// get identifier
		pos, err = p.checkToken("begin block", []string{"identifier"}, pos, tokens)
		if err != nil {
			return
		}
		pos++
	}

	// get the alwaysable statements
	generateableStatements, potentialPos, e := p.parseAlwaysStatements(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Statements = generateableStatements
	pos = potentialPos

	// check for end
	pos, err = p.checkToken("begin block", []string{"end"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}
func (p *Parser) parseIfBlock(tokens []Token, pos int) (result IfBlockNode, newPos int, err error) {
	// IF LPAREN <expr> RPAREN <generateable_statement> [ELSE <generateable_statement>]
	// get if
	pos, err = p.checkToken("if block", []string{"if"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get lparen
	pos, err = p.checkToken("if block", []string{"lparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get expression
	expression, potentialPos, e := p.parseExpression(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Expr = expression
	pos = potentialPos

	// get rparen
	pos, err = p.checkToken("if block", []string{"rparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get generateable statement
	generateableStatement, potentialPos, e := p.parseAlwaysStatement(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Body = generateableStatement
	pos = potentialPos

	// get else
	potentialPos, e = p.checkToken("if block", []string{"else"}, pos, tokens)
	if e == nil {
		pos = potentialPos + 1
		// get always statement
		generateableStatement, potentialPos, e := p.parseAlwaysStatement(tokens, pos)
		if e != nil {
			err = e
			return
		}
		result.Else = &generateableStatement
		pos = potentialPos
	}

	newPos = pos
	return
}
func (p *Parser) parseForBlock(tokens []Token, pos int) (result ForBlockNode, newPos int, err error) {
	// <for> -> FOR LPAREN [<assignment_without_semicolon>] SEMICOLON [<expr>] SEMICOLON [<assignment_without_semicolon>] RPAREN <begin_block>
	// get for
	pos, err = p.checkToken("for block", []string{"for"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	// get lparen
	pos, err = p.checkToken("for block", []string{"lparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	// get assignment_without_semicolon, optionally
	assignmentWithoutSemicolon, potentialPos, e := p.parseAssignmentNodeWithoutSemicolon(tokens, pos)
	if e == nil {
		result.Initializer = &assignmentWithoutSemicolon
		pos = potentialPos
	}
	// get semicolon
	pos, err = p.checkToken("for block", []string{"semicolon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	// get expr, optionally
	expr, potentialPos, e := p.parseExpression(tokens, pos)
	if e == nil {
		result.Condition = &expr
		pos = potentialPos
	}
	// get semicolon
	pos, err = p.checkToken("for block", []string{"semicolon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	// get assignment_without_semicolon, optionally
	assignmentWithoutSemicolon, potentialPos, e = p.parseAssignmentNodeWithoutSemicolon(tokens, pos)
	if e == nil {
		result.Incrementor = &assignmentWithoutSemicolon
		pos = potentialPos
	}

	// get rparen
	pos, err = p.checkToken("for block", []string{"rparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	// get statement
	body, potentialPos, e := p.parseAlwaysStatement(tokens, pos)
	if e != nil {
		err = e
		return
	}
	pos = potentialPos
	result.Body = body
	newPos = pos
	return
}
func (p *Parser) parseCaseNode(tokens []Token, pos int) (result CaseNode, newPos int, err error) {
	// <case> -> <expr> { COMMA <expr> } COLON <alwaysable_statement>
	expr, pos, err := p.parseExpression(tokens, pos)
	if err != nil {
		return
	}
	result.Conditions = append(result.Conditions, expr)

	// get other expressions, optionally
	potentialPos, e := p.checkToken("case", []string{"comma"}, pos, tokens)
	for e == nil {
		pos = potentialPos + 1
		// get expr
		expr, pos, err = p.parseExpression(tokens, pos)
		if err != nil {
			return
		}
		result.Conditions = append(result.Conditions, expr)
		potentialPos, e = p.checkToken("case", []string{"comma"}, pos, tokens)
	}

	// get colon
	pos, err = p.checkToken("case", []string{"colon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get alwaysable statement
	body, potentialPos, e := p.parseAlwaysStatement(tokens, pos)
	if e != nil {
		err = e
		return
	}
	pos = potentialPos
	result.Statement = body
	newPos = pos
	return
}

// <case_block> -> CASE LPAREN <expr> RPAREN {<case>} [ DEFAULT COLON <alwaysable_statement> ] ENDCASE
func (p *Parser) parseCaseBlock(tokens []Token, pos int) (result CaseBlock, newPos int, err error) {
	// get case
	pos, err = p.checkToken("case block", []string{"case"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get lparen
	pos, err = p.checkToken("case block", []string{"lparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get expr
	expr, pos, err := p.parseExpression(tokens, pos)
	if err != nil {
		return
	}
	result.Expr = expr

	// get rparen
	pos, err = p.checkToken("case block", []string{"rparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get cases
	caseNode, potentialPos, e := p.parseCaseNode(tokens, pos)
	for e == nil {
		result.Cases = append(result.Cases, caseNode)
		pos = potentialPos
		caseNode, potentialPos, e = p.parseCaseNode(tokens, pos)
	}

	// get default, optionally
	potentialPos, e = p.checkToken("case block", []string{"default"}, pos, tokens)
	if e == nil {
		pos = potentialPos + 1
		// get colon
		pos, err = p.checkToken("case block", []string{"colon"}, pos, tokens)
		if err != nil {
			return
		}
		pos++
		// get alwaysable statement
		body, potentialPos, e := p.parseAlwaysStatement(tokens, pos)
		if e != nil {
			err = e
			return
		}
		pos = potentialPos
		result.Default = &body
	}

	// get endcase
	pos, err = p.checkToken("case block", []string{"endcase"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	newPos = pos
	return
}

func (p *Parser) parseAlwaysStatement(tokens []Token, pos int) (result AlwaysStatement, newPos int, err error) {
	// <always_statement> -> <begin_block> | <interior_statement> | <for> | <if> | <builtin_function_call> | <delay_statement>
	beginResult, potentialPos, e := p.parseBeginBlock(tokens, pos)
	if e == nil {
		result.BeginBlock = &beginResult
		pos = potentialPos
	} else {
		interiorResult, potentialPos, e := p.parseInteriorStatement(tokens, pos)
		if e == nil {
			result.InteriorNode = &interiorResult
			pos = potentialPos
		} else {
			forResult, potentialPos, e := p.parseForBlock(tokens, pos)
			if e == nil {
				result.ForBlock = &forResult
				pos = potentialPos
			} else {
				ifResult, potentialPos, e := p.parseIfBlock(tokens, pos)
				if e == nil {
					result.IfBlock = &ifResult
					pos = potentialPos
				} else {
					functionResult, potentialPos, e := p.parseBuiltinFunctionCall(tokens, pos)
					if e == nil {
						result.FunctionNode = &functionResult
						pos = potentialPos
					} else {
						delayNode, potentialPos, e := p.parseDelayStatement(tokens, pos)
						if e == nil {
							result.DelayNode = &delayNode
							pos = potentialPos
						} else {
							caseNode, potentialPos, e := p.parseCaseBlock(tokens, pos)
							if e == nil {
								result.CaseNode = &caseNode
								pos = potentialPos
							} else {
								err = e
							}
						}
					}
				}
			}
		}
	}
	newPos = pos
	return
}
func (p *Parser) parseAlwaysStatements(tokens []Token, pos int) (result []AlwaysStatement, newPos int, err error) {
	generateableStatement, potentialPos, e := p.parseAlwaysStatement(tokens, pos)
	for e == nil {
		result = append(result, generateableStatement)
		pos = potentialPos
		generateableStatement, potentialPos, e = p.parseAlwaysStatement(tokens, pos)
	}
	newPos = pos
	return
}
func (p *Parser) parseGenerate(tokens []Token, pos int) (result GenerateNode, newPos int, err error) {
	// <generate> -> GENERATE <generateable_statements> ENDGENERATE
	// get generate
	pos, err = p.checkToken("generate", []string{"generate"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get generateable_statements
	generateableStatements, potentialPos, e := p.parseAlwaysStatements(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Statements = generateableStatements
	pos = potentialPos

	// get endgenerate
	pos, err = p.checkToken("generate", []string{"endgenerate"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	newPos = pos
	return
}

// <time> -> [TIME] <identifier>
func (p *Parser) parseTime(tokens []Token, pos int) (result TimeNode, newPos int, err error) {
	// get time, optionally
	potentialPos, e := p.checkToken("time", []string{"time"}, pos, tokens)
	if e == nil {
		result.Time = &tokens[potentialPos]
		pos = potentialPos + 1
	}

	// get identifier
	pos, err = p.checkToken("time", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Identifier = tokens[pos]
	pos++
	newPos = pos
	return
}

// <event> -> <time> { OR <time> }
func (p *Parser) parseEvent(tokens []Token, pos int) (result []TimeNode, newPos int, err error) {
	// get time
	timeNode, pos, err := p.parseTime(tokens, pos)
	if err != nil {
		return
	}
	result = append(result, timeNode)

	// get other times
	potentialPos, e := p.checkToken("event", []string{"identifier"}, pos, tokens)
	for e == nil {
		// special case because or is technically a valid identifier
		if tokens[potentialPos].Value != "or" {
			err = fmt.Errorf("expected 'or' but got '%s'", tokens[potentialPos].Value)
			return
		}

		// take the time
		timeNode, potentialPos, e = p.parseTime(tokens, potentialPos+1)
		if e == nil {
			result = append(result, timeNode)
			pos = potentialPos

			potentialPos, e = p.checkToken("event", []string{"identifier"}, pos, tokens)
		} else {
			err = e
		}
	}
	newPos = pos
	return
}

// <delay_statement> -> POUND [ LITERAL | <identifier> ]
func (p *Parser) parseDelayStatement(tokens []Token, pos int) (result DelayNode, newPos int, err error) {
	pos, err = p.checkToken("delay", []string{"pound"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get literal or identifier
	pos, err = p.checkToken("delay", []string{"literal", "identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Amount = tokens[pos]
	pos++
	newPos = pos
	return
}

func (p *Parser) parseAlways(tokens []Token, pos int) (result AlwaysNode, newPos int, err error) {
	// <always> -> ALWAYS [ AT LPAREN <event> RPAREN ] <alwaysable_statement>

	// get always
	pos, err = p.checkToken("always", []string{"always"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get at, optionally
	potentialPos, e := p.checkToken("always", []string{"at"}, pos, tokens)
	if e == nil {
		pos = potentialPos + 1
		// get lparen
		pos, err = p.checkToken("always", []string{"lparen"}, pos, tokens)
		if err != nil {
			return
		}
		pos++
		// get event
		event, potentialPos, e := p.parseEvent(tokens, pos)
		if e != nil {
			err = e
			return
		}
		pos = potentialPos
		result.Times = event

		// get rparen
		pos, err = p.checkToken("always", []string{"rparen"}, pos, tokens)
		if err != nil {
			return
		}
		pos++
	}

	// get alwaysable statement
	alwaysStatement, potentialPos, e := p.parseAlwaysStatement(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Statement = alwaysStatement
	pos = potentialPos
	newPos = pos
	return
}

func (p *Parser) parseBuiltinFunctionCall(tokens []Token, pos int) (result FunctionNode, newPos int, err error) {
	// <builtin_function_call> -> DOLLAR <identifier> LPAREN <expr> { COMMA <expr> } RPAREN SEMICOLON

	// get dollar
	pos, err = p.checkToken("builtin function call", []string{"dollar"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get identifier
	pos, err = p.checkToken("builtin function call", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Function = tokens[pos]
	pos++

	// get lparen, optionally
	potentialPos, e := p.checkToken("builtin function call", []string{"lparen"}, pos, tokens)
	if e == nil {
		pos = potentialPos + 1

		// get expr
		expr, potentialPos, e := p.parseExpression(tokens, pos)
		if e != nil {
			err = e
			return
		}
		result.Expressions = append(result.Expressions, expr)
		pos = potentialPos

		// get any other expressions
		potentialPos, e = p.checkToken("builtin function call", []string{"comma"}, pos, tokens)
		for e == nil {
			// take the expr
			expr, potentialPos, e = p.parseExpression(tokens, potentialPos+1)
			if e == nil {
				result.Expressions = append(result.Expressions, expr)
				pos = potentialPos
			} else {
				err = e
				return
			}
			potentialPos, e = p.checkToken("builtin function call", []string{"comma"}, pos, tokens)
		}

		// get rparen
		pos, err = p.checkToken("builtin function call", []string{"rparen"}, pos, tokens)
		if err != nil {
			return
		}
		pos++
	}

	// get semicolon
	pos, err = p.checkToken("builtin function call", []string{"semicolon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

func (p *Parser) parseDefParamNode(tokens []Token, pos int) (result DefParamNode, newPos int, err error) {
	// <def_param> -> DEFPARAM <identifier> { DOT <identifier> } EQUAL <expr> SEMICOLON

	// get defparam
	pos, err = p.checkToken("def param", []string{"defparam"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get identifier
	pos, err = p.checkToken("def param", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Identifiers = append(result.Identifiers, tokens[pos])
	pos++

	// get any other identifiers
	potentialPos, e := p.checkToken("def param", []string{"dot"}, pos, tokens)
	for e == nil {
		// take the identifier
		pos, err = p.checkToken("def param", []string{"identifier"}, potentialPos+1, tokens)
		if err != nil {
			return
		}
		result.Identifiers = append(result.Identifiers, tokens[pos])
		pos++
		potentialPos, e = p.checkToken("def param", []string{"dot"}, pos, tokens)
	}

	// get equal
	pos, err = p.checkToken("def param", []string{"equal"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get expr
	expr, potentialPos, e := p.parseExpression(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Value = expr
	pos = potentialPos

	// get semicolon
	pos, err = p.checkToken("def param", []string{"semicolon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

func (p *Parser) parseInitial(tokens []Token, pos int) (result InitialNode, newPos int, err error) {
	// <initial> -> INITIAL <alwaysable_statement>
	pos, err = p.checkToken("initial", []string{"initial"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get alwaysable statement
	alwaysableStatement, potentialPos, e := p.parseAlwaysStatement(tokens, pos)
	if e != nil {
		err = e
		return
	}
	result.Statement = alwaysableStatement
	pos = potentialPos

	newPos = pos
	return
}

func (p *Parser) parseInteriorStatement(tokens []Token, pos int) (result InteriorNode, newPos int, err error) {
	// it could be either a declaration or module_application or assignment or generate

	// check if it's a declaration
	declarationNode, potentialPos, e := p.parseDeclarationNode(tokens, pos)
	if e == nil {
		// success!
		result.DeclarationNode = &declarationNode
		pos = potentialPos
	} else {
		// check if it's a module application
		moduleApplicationNode, potentialPos, e := p.parseModuleApplication(tokens, pos)
		if e == nil {
			// success!
			result.ModuleApplicationNode = &moduleApplicationNode
			pos = potentialPos
		} else {
			// check if it's an assignment
			assignmentNode, potentialPos, e := p.parseAssignmentNode(tokens, pos)
			if e == nil {
				// success!
				result.AssignmentNode = &assignmentNode
				pos = potentialPos
			} else {
				// check if it's a generate
				generateNode, potentialPos, e := p.parseGenerate(tokens, pos)
				if e == nil {
					result.GenerateNode = &generateNode
					pos = potentialPos
				} else {
					// check if it's an always
					alwaysNode, potentialPos, e := p.parseAlways(tokens, pos)
					if e == nil {
						result.AlwaysNode = &alwaysNode
						pos = potentialPos
					} else {
						// check if it's a defparam
						defParamNode, potentialPos, e := p.parseDefParamNode(tokens, pos)
						if e == nil {
							result.DefParamNode = &defParamNode
							pos = potentialPos
						} else {
							// check if it's an initial
							initialNode, potentialPos, e := p.parseInitial(tokens, pos)
							if e == nil {
								result.InitialNode = &initialNode
								pos = potentialPos
							} else {
								// check if it's a directive
								directiveNode, potentialPos, e := p.parseDirective(tokens, pos)
								if e == nil {
									result.DirectiveNode = directiveNode
									pos = potentialPos
								} else {
									// check if it's a task
									taskNode, potentialPos, e := p.parseTask(tokens, pos)
									if e == nil {
										result.TaskNode = &taskNode
										pos = potentialPos
									} else {
										err = e
									}
								}
							}
						}
					}
				}
			}
		}
	}

	newPos = pos
	return
}
func (p *Parser) parseModuleInterior(tokens []Token, pos int) (result []InteriorNode, newPos int, err error) {
	for {
		nestedStatement, potentialPos, e := p.parseInteriorStatement(tokens, pos)
		if e != nil {
			return
		}
		result = append(result, nestedStatement)
		pos = potentialPos
		newPos = pos
	}
}
func (p *Parser) parseTaskInterior(tokens []Token, pos int) (result []TaskStatement, newPos int, err error) {
	result = []TaskStatement{}
	for {
		nestedStatement, potentialPos, e := p.parseInteriorStatement(tokens, pos)
		if e != nil {
			// try to parse a begin block, which is legal in a task
			beginBlock, potentialPos, e := p.parseBeginBlock(tokens, pos)
			if e == nil {
				result = append(result, TaskStatement{BeginBlock: &beginBlock})
				pos = potentialPos
			} else {
				newPos = pos
				return
			}
		} else {
			result = append(result, TaskStatement{InteriorNode: &nestedStatement})
			pos = potentialPos
		}
	}
}
func (p *Parser) parseTask(tokens []Token, pos int) (result TaskNode, newPos int, err error) {
	pos, err = p.checkToken("task", []string{"task"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	pos, err = p.checkToken("task", []string{"identifier"}, pos, tokens)
	if err != nil {
		return
	}
	result.Identifier = tokens[pos]
	pos++

	pos, err = p.checkToken("task", []string{"semicolon"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	result.Statements, pos, err = p.parseTaskInterior(tokens, pos)
	if err != nil {
		return
	}

	// get endtask
	pos, err = p.checkToken("task", []string{"endtask"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}

// ==============================
// Module Definition Section
// ==============================

// Returns a list of ports, and newPos is the position after the list
func (p *Parser) parsePorts(tokens []Token, pos int) (result []Token, newPos int, err error) {
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
func (p *Parser) parsePortList(tokens []Token, pos int) (result PortListNode, newPos int, err error) {
	// take lparen
	pos, err = p.checkToken("port list", []string{"lparen"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	// get ports if any
	ports, potentialPos, e := p.parsePorts(tokens, pos)
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

func (p *Parser) parseModule(tokens []Token, pos int) (result ModuleNode, newPos int, err error) {
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
	portList, potentialPos, e := p.parsePortList(tokens, pos)
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
	interior, potentialPos, e := p.parseModuleInterior(tokens, pos)
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
func (p *Parser) parseDefine(tokens []Token, pos int) (result *DefineNode, newPos int, err error) {
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
func (p *Parser) skipTimescale(tokens []Token, pos int) (newPos int, err error) {
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

// <include> -> include <literal>
func (p *Parser) skipInclude(tokens []Token, pos int) (newPos int, err error) {
	pos, err = p.checkToken("include", []string{"include"}, pos, tokens)
	if err != nil {
		return
	}
	pos++

	pos, err = p.checkToken("include", []string{"literal"}, pos, tokens)
	if err != nil {
		return
	}
	pos++
	newPos = pos
	return
}
func (p *Parser) parseDirective(tokens []Token, pos int) (result *DefineNode, newPos int, err error) {
	// directive is just a define, timescale, or include
	result, newPos, err = p.parseDefine(tokens, pos)
	if err == nil {
		return // success with define
	}

	newPos, err = p.skipTimescale(tokens, pos)
	if err == nil {
		return // success with timescale
	}

	newPos, err = p.skipInclude(tokens, pos)
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
		directive, newPos, e := p.parseDirective(tokens, pos)
		if e != nil {
			// try module
			module, newPos, e := p.parseModule(tokens, pos)
			if e != nil {
				err = *p.FarthestError
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
func getInteriorStatementsFromAlwaysStatements(statements []AlwaysStatement) []InteriorNode {
	var result []InteriorNode
	for _, statement := range statements {
		result = append(result, getInteriorStatementsFromAlwaysStatement(statement)...)
	}
	return result
}
func getInteriorStatementsFromAlwaysStatement(statement AlwaysStatement) []InteriorNode {
	var result []InteriorNode
	if statement.BeginBlock != nil {
		result = append(result, getInteriorStatementsFromAlwaysStatements(statement.BeginBlock.Statements)...)
	} else if statement.CaseNode != nil {
		// add regular cases
		for _, caseNode := range statement.CaseNode.Cases {
			result = append(result, getInteriorStatementsFromAlwaysStatement(caseNode.Statement)...)
		}
		// add default case
		if statement.CaseNode.Default != nil {
			result = append(result, getInteriorStatementsFromAlwaysStatement(*statement.CaseNode.Default)...)
		}
	} else if statement.ForBlock != nil {
		result = append(result, getInteriorStatementsFromAlwaysStatement(statement.ForBlock.Body)...)
	} else if statement.IfBlock != nil {
		result = append(result, getInteriorStatementsFromAlwaysStatement(statement.IfBlock.Body)...)
		if statement.IfBlock.Else != nil {
			result = append(result, getInteriorStatementsFromAlwaysStatement(*statement.IfBlock.Else)...)
		}
	} else if statement.InteriorNode != nil {
		result = append(result, getInteriorStatementsFromInteriorNode(*statement.InteriorNode)...)
	}
	return result
}
func getIteriorStatementsFromTaskNode(taskNode TaskNode) []InteriorNode {
	var result []InteriorNode
	for _, statement := range taskNode.Statements {
		if statement.BeginBlock != nil {
			result = append(result, getInteriorStatementsFromAlwaysStatements(statement.BeginBlock.Statements)...)
		} else if statement.InteriorNode != nil {
			result = append(result, getInteriorStatementsFromInteriorNode(*statement.InteriorNode)...)
		}
	}
	return result
}
func getInteriorStatementsFromInteriorNode(interiorNode InteriorNode) []InteriorNode {
	var result []InteriorNode
	if interiorNode.AlwaysNode != nil {
		result = append(result, getInteriorStatementsFromAlwaysStatement(interiorNode.AlwaysNode.Statement)...)
	} else if interiorNode.GenerateNode != nil {
		result = append(result, getInteriorStatementsFromAlwaysStatements(interiorNode.GenerateNode.Statements)...)
	} else if interiorNode.InitialNode != nil {
		result = append(result, getInteriorStatementsFromAlwaysStatement(interiorNode.InitialNode.Statement)...)
	} else if interiorNode.TaskNode != nil {
		result = append(result, getIteriorStatementsFromTaskNode(*interiorNode.TaskNode)...)
	} else {
		// this belongs to the result
		result = append(result, interiorNode)
	}
	return result
}

func GetInteriorStatements(fileNode FileNode) []InteriorNode {
	var result []InteriorNode
	for _, statements := range fileNode.Statements {
		if statements.Module != nil {
			interior := statements.Module.Interior
			for _, statement := range interior {
				result = append(result, getInteriorStatementsFromInteriorNode(statement)...)
			}
		}
	}
	return result
}
func getFunctionStatementsFromAlwaysStatements(statements []AlwaysStatement) []FunctionNode {
	var result []FunctionNode
	for _, statement := range statements {
		result = append(result, getFunctionStatementsFromAlwaysStatement(statement)...)
	}
	return result
}
func getFunctionStatementsFromAlwaysStatement(statement AlwaysStatement) []FunctionNode {
	var result []FunctionNode
	if statement.BeginBlock != nil {
		result = append(result, getFunctionStatementsFromAlwaysStatements(statement.BeginBlock.Statements)...)
	} else if statement.CaseNode != nil {
		// add regular cases
		for _, caseNode := range statement.CaseNode.Cases {
			result = append(result, getFunctionStatementsFromAlwaysStatement(caseNode.Statement)...)
		}
		// add default case
		if statement.CaseNode.Default != nil {
			result = append(result, getFunctionStatementsFromAlwaysStatement(*statement.CaseNode.Default)...)
		}
	} else if statement.ForBlock != nil {
		result = append(result, getFunctionStatementsFromAlwaysStatement(statement.ForBlock.Body)...)
	} else if statement.FunctionNode != nil {
		result = append(result, *statement.FunctionNode)
	} else if statement.IfBlock != nil {
		result = append(result, getFunctionStatementsFromAlwaysStatement(statement.IfBlock.Body)...)
		if statement.IfBlock.Else != nil {
			result = append(result, getFunctionStatementsFromAlwaysStatement(*statement.IfBlock.Else)...)
		}
	} else if statement.InteriorNode != nil {
		result = append(result, getFunctionStatementsFromInteriorNode(*statement.InteriorNode)...)
	}
	return result
}
func getFunctionStatementsFromTaskNode(taskNode TaskNode) []FunctionNode {
	var result []FunctionNode
	for _, statement := range taskNode.Statements {
		if statement.BeginBlock != nil {
			result = append(result, getFunctionStatementsFromAlwaysStatements(statement.BeginBlock.Statements)...)
		} else if statement.InteriorNode != nil {
			result = append(result, getFunctionStatementsFromInteriorNode(*statement.InteriorNode)...)
		}
	}
	return result
}
func getFunctionStatementsFromInteriorNode(interiorNode InteriorNode) []FunctionNode {
	var result []FunctionNode
	if interiorNode.AlwaysNode != nil {
		result = append(result, getFunctionStatementsFromAlwaysStatement(interiorNode.AlwaysNode.Statement)...)
	} else if interiorNode.GenerateNode != nil {
		result = append(result, getFunctionStatementsFromAlwaysStatements(interiorNode.GenerateNode.Statements)...)
	} else if interiorNode.InitialNode != nil {
		result = append(result, getFunctionStatementsFromAlwaysStatement(interiorNode.InitialNode.Statement)...)
	} else if interiorNode.TaskNode != nil {
		result = append(result, getFunctionStatementsFromTaskNode(*interiorNode.TaskNode)...)
	}
	return result
}

func GetFunctionNodes(fileNode FileNode) []FunctionNode {
	var result []FunctionNode
	for _, statements := range fileNode.Statements {
		if statements.Module != nil {
			interior := statements.Module.Interior
			for _, statement := range interior {
				result = append(result, getFunctionStatementsFromInteriorNode(statement)...)
			}
		}
	}
	return result
}

/**
Grammar:

// ==============================
// Top Level Grammar
// ==============================
<file> -> <statement> { <statement> }
<statement> -> <module> | <directive>

// useful helper grammars
<identifier> -> IDENTIFIER


// ==============================
// Directive Grammar
// ==============================
<directive> -> <include> | <timescale> | <define>
<include> -> INCLUDE LITERAL
<timescale> -> TIMESCALE <non-newline> NEWLINE
<define> -> DEFINE <identifier> <non-newline> NEWLINE

// ==============================
// Module Grammar
// ==============================
<module> -> MODULE <identifier> [<portlist>] SEMICOLON <interior> ENDMODULE [SEMICOLON]
<portlist> -> LPAREN [<ports>] RPAREN
<ports> -> <identifier> { COMMA <identifier> }

<interior> -> { <interior_statement> }
<interior_statement>  -> <declaration> | <module_application> | <assignment> | <generate> | <always> | <defparam> | <initial> | <directive> | <task>
<task_statement> -> <interior_statement> | <begin_block>
<task> -> TASK <identifier> SEMICOLON <task_statement> ENDTASK [SEMICOLON]

<assignment_without_semicolon> -> [ASSIGN] <identifier> [<index>] (EQUAL | <=) <expr>
<assignment> -> <assignment_without_semicolon> SEMICOLON
<single_var> -> <identifier> {<range>}

<declaration> -> <type> <single_var> EQUAL <expr> { COMMA <single_var> EQUAL <expr> } SEMICOLON
| <type> <single_var> { COMMA <single_var> } SEMICOLON
<type> -> (TYPE | DIRECTION [TYPE]) {<range>}
<index> -> LBRACKET <identifier> RBRACKET | LBRACKET <integer> RBRACKET
<range> -> LBRACKET <integer> COLON <integer> RBRACKET
<integer> -> LITERAL | DEFINE

<module_application> -> <identifier> [<identifier>] [<range>] LPAREN <arguments> RPAREN SEMICOLON
<arguments> -> <argument> { COMMA <argument> }
<argument> -> DOT <identifier> LPAREN  <expr>  RPAREN | <expr>
<selector> -> LBRACKET <expr> [COLON <expr>] RBRACKET
<expr> -> <maybe_signed> [OPERATOR <expr>] [ COMPARATOR <expr> [ QUESTION <expr> COLON <expr> ] ]
			| LPAREN <expr> RPAREN
<maybed_signed> -> <sized_value> | SIGNED LPAREN <sized_value> RPAREN
<sized_value> -> [ LITERAL | <identifier> ] LCURL <sized_value> { COMMA <sized_value> } RCURL | <value>
<value> -> [TILDE| - ] (LITERAL|(<identifier> { DOT <identifier> })|FUNCLITERAL) { <selector> }

<defparam> -> DEFPARAM <identifier> { DOT <identifier> } EQUAL <expr> SEMICOLON

<generate> -> GENERATE { <alwaysable_statement> } ENDGENERATE
<begin_block> -> BEGIN [ COLON <identifier> ] { <alwaysable_statement> } END
<for> -> FOR LPAREN [<assignment_without_semicolon>] SEMICOLON [<expr>] SEMICOLON [<assignment_without_semicolon>] RPAREN <alwaysable_statement>
<if> -> IF LPAREN <expr> RPAREN <alwaysable_statement> [ELSE <alwaysable_statement>]
<builtin_function_call> -> DOLLAR <identifier> [LPAREN <expr> { COMMA <expr> } RPAREN] SEMICOLON

<always> -> ALWAYS [ AT LPAREN <event> RPAREN ] <alwaysable_statement>
<event> -> <time> { OR <time> }
<time> -> [ TIME ] <identifier>
<alwaysable_statement> -> <begin_block> | <interior_statement> | <for> | <if> | <builtin_function_call> | <delay_statement> | <case_block>
<delay_statement> -> POUND [ LITERAL | <identifier> ]
<case_block> -> CASE LPAREN <expr> RPAREN {<case>} [ DEFAULT COLON <alwaysable_statement> ] ENDCASE
<case> -> <expr> { COMMA <expr> } COLON <alwaysable_statement>

<initial> -> INITIAL <alwaysable_statement>
*/
