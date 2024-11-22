package lang

import (
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

type Interpreter struct {
	builtins    map[string]bool
	defines     []DefineNode
	Diagnostics []protocol.Diagnostic
	moduleMap   map[string]ModuleNode
	log         *zap.Logger
}

func NewInterpreter(logger *zap.Logger, modules map[string][]ModuleNode, defines map[string][]DefineNode) *Interpreter {
	moduleMap := map[string]ModuleNode{}
	flattenedDefines := []DefineNode{}
	for _, mods := range modules {
		for _, module := range mods {
			moduleMap[module.Identifier.Value] = module
		}
	}
	for _, defs := range defines {
		flattenedDefines = append(flattenedDefines, defs...)
	}
	builtins := map[string]bool{
		"and":    true,
		"or":     true,
		"xor":    true,
		"nand":   true,
		"nor":    true,
		"xnor":   true,
		"buf":    true,
		"not":    true,
		"bufif1": true,
		"notif1": true,
		"bufif0": true,
		"notif0": true,
	}

	return &Interpreter{
		defines:     flattenedDefines,
		Diagnostics: []protocol.Diagnostic{},
		moduleMap:   moduleMap,
		log:         logger,
		builtins:    builtins,
	}
}
func (i *Interpreter) addUnknownDiagnostic(identifier Token, description string) {
	i.Diagnostics = append(i.Diagnostics, protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(identifier.line),
				Character: uint32(identifier.startCharacter),
			},
			End: protocol.Position{
				Line:      uint32(identifier.line),
				Character: uint32(identifier.endCharacter),
			},
		},
		Severity: protocol.DiagnosticSeverityWarning,
		Message:  "Unknown " + description + ": " + identifier.Value,
	})
}
func (i *Interpreter) diagnoseSelector(node SelectorNode, curSymbols map[string]bool) {
	if node.IndexNode != nil {
		i.diagnoseExpression(node.IndexNode.Index, curSymbols)
	} else if node.RangeNode != nil {
		i.diagnoseExpression(node.RangeNode.From, curSymbols)
		i.diagnoseExpression(node.RangeNode.To, curSymbols)
	}
}
func (i *Interpreter) diagnoseExpression(node ExprNode, curSymbols map[string]bool) {
	// expressions don't add new variables, so just look at existing
	for _, val := range node.Value.Values {
		for _, tok := range val.Value {
			if tok.Type == "identifier" {
				_, ok := curSymbols[tok.Value]
				if !ok {
					i.addUnknownDiagnostic(tok, "variable")
				}
			}
		}
		for _, selector := range val.Selectors {
			i.diagnoseSelector(selector, curSymbols)
		}
	}
	if node.Right != nil {
		i.diagnoseExpression(*node.Right, curSymbols)
	}
	if node.ExprTrue != nil {
		i.diagnoseExpression(*node.ExprTrue, curSymbols)
	}
	if node.ExprFalse != nil {
		i.diagnoseExpression(*node.ExprFalse, curSymbols)
	}
}
func (i *Interpreter) diagnoseAlwaysNode(node AlwaysStatement, curSymbols map[string]bool) map[string]bool {
	knownSymbols := curSymbols
	if node.BeginBlock != nil {
		for _, statement := range node.BeginBlock.Statements {
			knownSymbols = i.diagnoseAlwaysNode(statement, knownSymbols)
		}
	} else if node.CaseNode != nil {
		i.diagnoseExpression(node.CaseNode.Expr, knownSymbols)
		for _, case_ := range node.CaseNode.Cases {
			for _, cond := range case_.Conditions {
				i.diagnoseExpression(cond, knownSymbols)
			}
			knownSymbols = i.diagnoseAlwaysNode(case_.Statement, knownSymbols)
		}
	} else if node.ForBlock != nil {
		if node.ForBlock.Initializer != nil {
			i.diagnoseAssignmentNode(*node.ForBlock.Initializer, knownSymbols)
		}
		if node.ForBlock.Condition != nil {
			i.diagnoseExpression(*node.ForBlock.Condition, knownSymbols)
		}
		if node.ForBlock.Incrementor != nil {
			i.diagnoseAssignmentNode(*node.ForBlock.Incrementor, knownSymbols)
		}
		i.diagnoseAlwaysNode(node.ForBlock.Body, knownSymbols)
	} else if node.FunctionNode != nil {
		for _, arg := range node.FunctionNode.Expressions {
			i.diagnoseExpression(arg, knownSymbols)
		}
	} else if node.IfBlock != nil {
		i.diagnoseExpression(node.IfBlock.Expr, knownSymbols)
		knownSymbols = i.diagnoseAlwaysNode(node.IfBlock.Body, knownSymbols)
		if node.IfBlock.Else != nil {
			knownSymbols = i.diagnoseAlwaysNode(*node.IfBlock.Else, knownSymbols)
		}
	} else if node.InteriorNode != nil {
		knownSymbols = i.diagnoseInteriorNode(*node.InteriorNode, knownSymbols)
	}
	return knownSymbols
}
func (i *Interpreter) diagnoseAlwaysStatements(statements []AlwaysStatement, curSymbols map[string]bool) map[string]bool {
	knownSymbols := curSymbols
	for _, statement := range statements {
		knownSymbols = i.diagnoseAlwaysNode(statement, knownSymbols)
	}
	return knownSymbols
}
func (i *Interpreter) diagnoseAssignmentNode(node AssignmentNode, curSymbols map[string]bool) map[string]bool {
	knownSymbols := curSymbols
	for _, variable := range node.Variables {
		_, ok := knownSymbols[variable.Identifier.Value]
		if !ok {
			i.addUnknownDiagnostic(variable.Identifier, "variable")
		}
		for _, selector := range variable.Selectors {
			if selector.IndexNode != nil {
				i.diagnoseExpression(selector.IndexNode.Index, curSymbols)
			} else if selector.RangeNode != nil {
				i.diagnoseExpression(selector.RangeNode.From, curSymbols)
				i.diagnoseExpression(selector.RangeNode.To, curSymbols)
			}
		}
	}
	// also check the right hand side
	i.diagnoseExpression(node.Value, knownSymbols)
	return knownSymbols
}
func (i *Interpreter) diagnoseInteriorNode(node InteriorNode, curSymbols map[string]bool) map[string]bool {
	knownSymbols := curSymbols

	if node.AssignmentNode != nil {
		i.diagnoseAssignmentNode(*node.AssignmentNode, knownSymbols)
	} else if node.DeclarationNode != nil {
		for _, variable := range node.DeclarationNode.Variables {
			knownSymbols[variable.Identifier.Value] = true
		}
	} else if node.ModuleApplicationNode != nil {
		name := node.ModuleApplicationNode.ModuleName.Value
		mod, ok := i.moduleMap[name]
		_, lessOk := i.builtins[name]
		if !ok && !lessOk {
			i.addUnknownDiagnostic(node.ModuleApplicationNode.ModuleName, "module")
		}
		for _, argument := range node.ModuleApplicationNode.Arguments {
			i.diagnoseExpression(argument.Value, knownSymbols)

			if argument.Label != nil && ok {
				exists := false
				for _, port := range mod.PortList.Ports {
					if port.Value == argument.Label.Value {
						exists = true
					}
				}
				if !exists {
					i.addUnknownDiagnostic(*argument.Label, "module port")
				}
			}
		}
	} else if node.AlwaysNode != nil {
		knownSymbols = i.diagnoseAlwaysNode(node.AlwaysNode.Statement, knownSymbols)
	} else if node.DefParamNode != nil {
		i.diagnoseExpression(node.DefParamNode.Value, knownSymbols)
		for _, variable := range node.DefParamNode.Identifiers {
			knownSymbols[variable.Value] = true
		}
	} else if node.DirectiveNode != nil {
		knownSymbols[node.DirectiveNode.Identifier.Value] = true
	} else if node.GenerateNode != nil {
		knownSymbols = i.diagnoseAlwaysStatements(node.GenerateNode.Statements, knownSymbols)
	} else if node.InitialNode != nil {
		i.diagnoseAlwaysNode(node.InitialNode.Statement, knownSymbols)
	} else if node.TaskNode != nil {
		knownSymbols = i.diagnoseAlwaysStatements(node.TaskNode.Statements, knownSymbols)
	}

	return knownSymbols
}

func (i *Interpreter) diagnoseModule(module ModuleNode) {
	knownSymbols := map[string]bool{}
	for _, define := range i.defines {
		knownSymbols["`"+define.Identifier.Value] = true
	}
	for _, statement := range module.Interior {
		knownSymbols = i.diagnoseInteriorNode(statement, knownSymbols)
	}
}

func (i *Interpreter) Interpret(FileNode FileNode) []protocol.Diagnostic {
	for _, topLevelStatement := range FileNode.Statements {
		if topLevelStatement.Module != nil {
			i.diagnoseModule(*topLevelStatement.Module)
		}
	}

	return i.Diagnostics
}
