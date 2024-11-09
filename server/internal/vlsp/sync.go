package vlsp

import (
	"context"
	"strings"

	"go.lsp.dev/protocol"
)

func (h Handler) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) (err error) {
	file := params.TextDocument.URI.Filename()
	h.state.log.Sugar().Info("File that did change: ", file)

	if strings.HasSuffix(file, ".v") {
		// update file
		fnode, ok := h.state.files[file]
		if !ok {
			h.state.files[file] = NewFile(file)
			fnode = h.state.files[file]
		}
		fnode.SetContents(params.ContentChanges[len(params.ContentChanges)-1].Text)

		// update symbols
		h.GetSymbolsForFile(file, false)
	}
	return
}
func (h Handler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) (err error) {
	file := params.TextDocument.URI.Filename()
	h.state.log.Sugar().Info("File that did change: ", file)

	if strings.HasSuffix(file, ".v") {
		// update file
		fnode, ok := h.state.files[file]
		if !ok {
			h.state.files[file] = NewFile(file)
			fnode = h.state.files[file]
		}
		fnode.SetContents(params.TextDocument.Text)

		// update symbols
		h.GetSymbolsForFile(file, false)
	}
	return
}
func (h Handler) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) (err error) {
	file := params.TextDocument.URI.Filename()
	h.state.log.Sugar().Info("File that did close: ", file)

	return
}
func (h Handler) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) (err error) {
	file := params.TextDocument.URI.Filename()
	h.state.log.Sugar().Info("File that did change: ", file)

	if strings.HasSuffix(file, ".v") {
		// update file
		h.state.files[file].Save()

		// update symbols
		h.GetSymbolsForFile(file, false)
	}
	return
}
