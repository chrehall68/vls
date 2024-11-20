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

func (i *Interpreter) diagnoseInteriorNode(node InteriorNode, curSymbols map[string]bool) (newSymbols map[string]bool) {
	knownSymbols := curSymbols

	if node.AssignmentNode != nil {
		for _, variable := range node.AssignmentNode.Variables {
			_, ok := knownSymbols[variable.Identifier.Value]
			if !ok {
				i.addUnknownDiagnostic(variable.Identifier, "variable")
			}
		}
		// also check the right hand side
		i.diagnoseExpression(node.AssignmentNode.Value, knownSymbols)
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
	} // TODO - do the rest?

	newSymbols = knownSymbols
	return
}

func (i *Interpreter) diagnoseModule(module ModuleNode) {
	knownSymbols := map[string]bool{}
	for _, define := range i.defines {
		knownSymbols[define.Identifier.Value] = true
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
