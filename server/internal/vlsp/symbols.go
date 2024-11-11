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
		h.state.defines[fname] = []lang.DefineNode{}
		h.state.modules[fname] = []lang.ModuleNode{}

		// publish no diagnostics since stuff failed
		obj := protocol.PublishDiagnosticsParams{
			URI:         protocol.DocumentURI(PathToURI(fname)),
			Diagnostics: []protocol.Diagnostic{},
		}
		h.state.client.PublishDiagnostics(context.Background(), &obj)
	} else {
		for _, statement := range results.Statements {
			if statement.Module != nil {
				h.state.modules[fname] = append(h.state.modules[fname], *statement.Module)
			} else if statement.Directive != nil && statement.Directive.DefineNode != nil {
				h.state.defines[fname] = append(h.state.defines[fname], *statement.Directive.DefineNode)
			}
		}

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
