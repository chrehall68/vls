package vlsp

import (
	"context"
	"os"
	"strings"

	"github.com/chrehall68/vls/internal/lang"
	"go.lsp.dev/protocol"
)

func (h *Handler) getFileFullPaths(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		h.state.log.Sugar().Panic("while reading dir: ", err)
	}
	paths := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			paths = append(paths, h.getFileFullPaths(dir+"/"+entry.Name())...)
		} else {
			paths = append(paths, dir+"/"+entry.Name())
		}
	}
	return paths
}

func (h Handler) GetSymbolsForFile(fname string, firstTime bool) {
	vlexer := lang.NewVLexer(h.state.log)
	parser := lang.NewParser()

	// lex
	tokens, err := vlexer.Lex(h.state.files[fname].GetContents())
	if err != nil {
		h.state.log.Sugar().Errorf("error lexing file %s: %s", fname, err)
		return
	}

	// parse
	results, err := parser.ParseFile(tokens)
	if err != nil {
		h.state.log.Sugar().Errorf("error parsing file %s: %s", fname, err)
		// let's not clear anything for now...
		//h.state.defines[fname] = []lang.DefineNode{}
		//h.state.modules[fname] = []lang.ModuleNode{}

		// publish the error as a diagnostic
		if parser.FarthestErrorPosition >= 0 && parser.FarthestErrorPosition < len(tokens) {
			tok := tokens[parser.FarthestErrorPosition]
			diag := protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(tok.Line()), Character: uint32(tok.StartCharacter())},
					End:   protocol.Position{Line: uint32(tok.Line()), Character: uint32(tok.EndCharacter())},
				},
				Severity: protocol.DiagnosticSeverityError,
				Message:  err.Error(),
			}
			obj := protocol.PublishDiagnosticsParams{
				URI:         protocol.DocumentURI(PathToURI(fname)),
				Diagnostics: []protocol.Diagnostic{diag},
			}
			h.state.client.PublishDiagnostics(context.Background(), &obj)
		} else {
			// somehow we got an error out of bounds
			// might happen if someone puts an emoji in the file
			obj := protocol.PublishDiagnosticsParams{
				URI:         protocol.DocumentURI(PathToURI(fname)),
				Diagnostics: []protocol.Diagnostic{},
			}
			h.state.client.PublishDiagnostics(context.Background(), &obj)
		}
	} else {
		// reset maps for this file
		h.state.defines[fname] = []lang.DefineNode{}
		h.state.modules[fname] = []lang.ModuleNode{}

		// store all modules that way we can easily go to definition
		for _, statement := range results.Statements {
			if statement.Module != nil {
				h.state.modules[fname] = append(h.state.modules[fname], *statement.Module)

				// clear the existing variable definitions
				moduleName := statement.Module.Identifier.Value
				h.state.variableDefinitions[moduleName] = map[string]protocol.Location{}
				// and also store all variable definitions inside the module
				for _, statement := range lang.GetInteriorStatementsFromModule(*statement.Module) {
					if statement.DeclarationNode != nil {
						for _, v := range statement.DeclarationNode.Variables {
							h.state.variableDefinitions[moduleName][v.Identifier.Value] = protocol.Location{
								URI: protocol.DocumentURI(PathToURI(fname)),
								Range: protocol.Range{
									Start: protocol.Position{Line: uint32(v.Identifier.Line()), Character: uint32(v.Identifier.StartCharacter())},
									End:   protocol.Position{Line: uint32(v.Identifier.Line()), Character: uint32(v.Identifier.EndCharacter())}},
							}
						}
					}
				}
			} else if statement.Directive != nil && statement.Directive.DefineNode != nil {
				h.state.defines[fname] = append(h.state.defines[fname], *statement.Directive.DefineNode)
			}
		}
		// store all known global symbols (modules and defines)
		for _, module := range h.state.modules[fname] {
			h.state.symbolMap[module.Identifier.Value] = protocol.Location{
				URI: protocol.DocumentURI(PathToURI(fname)),
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(module.Identifier.Line()), Character: uint32(module.Identifier.StartCharacter())},
					End:   protocol.Position{Line: uint32(module.Identifier.Line()), Character: uint32(module.Identifier.EndCharacter())}},
			}
		}
		for _, define := range h.state.defines[fname] {
			h.state.symbolMap[define.Identifier.Value] = protocol.Location{
				URI: protocol.DocumentURI(PathToURI(fname)),
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(define.Identifier.Line()), Character: uint32(define.Identifier.StartCharacter())},
					End:   protocol.Position{Line: uint32(define.Identifier.Line()), Character: uint32(define.Identifier.EndCharacter())}},
			}
		}

		// get diagnostics
		if !firstTime {
			interpreter := lang.NewInterpreter(h.state.log, h.state.modules, h.state.defines)
			diagnostics := interpreter.Interpret(results)
			obj := protocol.PublishDiagnosticsParams{
				URI:         protocol.DocumentURI(PathToURI(fname)),
				Diagnostics: diagnostics,
			}
			h.state.client.PublishDiagnostics(context.Background(), &obj)
		}
	}
}

func (h Handler) GetSymbols() {
	// first, reset state
	h.state.defines = map[string][]lang.DefineNode{}
	h.state.modules = map[string][]lang.ModuleNode{}
	h.state.symbolMap = map[string]protocol.Location{}

	// then, get the files to parse
	files := h.getFileFullPaths(h.state.workspace)

	for _, file := range files {
		if strings.HasSuffix(file, ".v") {
			// create the file object
			h.state.files[file] = NewFile(file)

			// parse the file
			h.GetSymbolsForFile(file, true)

			// clear any existing diagnostics
			obj := protocol.PublishDiagnosticsParams{
				URI:         protocol.DocumentURI(PathToURI(file)),
				Diagnostics: []protocol.Diagnostic{},
			}
			h.state.client.PublishDiagnostics(context.Background(), &obj)
		}
	}

	// then publish actual diagnostics
	vlexer := lang.NewVLexer(h.state.log)
	parser := lang.NewParser()
	for _, file := range files {
		if strings.HasSuffix(file, ".v") {
			tokens, err := vlexer.Lex(h.state.files[file].GetContents())
			if err != nil {
				continue
			}
			results, err := parser.ParseFile(tokens)
			if err != nil {
				continue
			}
			interpreter := lang.NewInterpreter(h.state.log, h.state.modules, h.state.defines)
			diagnostics := interpreter.Interpret(results)
			obj := protocol.PublishDiagnosticsParams{
				URI:         protocol.DocumentURI(PathToURI(file)),
				Diagnostics: diagnostics,
			}
			h.state.client.PublishDiagnostics(context.Background(), &obj)
		}
	}
}
