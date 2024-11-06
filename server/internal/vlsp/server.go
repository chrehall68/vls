package vlsp

import (
	"context"
	"strings"

	"github.com/chrehall68/vls/internal/lang"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

type ServerState struct {
	workspace string
	modules   []lang.Module     // list of all modules
	defines   []string          // list of all defines
	symbolMap map[string]string // map of symbol names to their full path
	log       *zap.Logger
}

type Handler struct {
	protocol.Server
	state *ServerState
}

func (h Handler) String() string {
	return "vlsp"
}

func NewHandler(ctx context.Context, server protocol.Server, logger *zap.Logger) (Handler, context.Context, error) {
	// Do initialization logic here, including
	// stuff like setting state variables
	// by returning a new context with
	// context.WithValue(context, ...)
	// instead of just context
	return Handler{Server: server, state: &ServerState{workspace: "", modules: []lang.Module{}, defines: []string{}, log: logger}}, ctx, nil
}

func (h Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	h.state.log.Sugar().Infof("Initialize called")

	workspace := params.WorkspaceFolders[0].URI
	h.state.log.Sugar().Infof("workspace: %v", workspace)
	if workspace != "" {
		h.state.log.Sugar().Info("setting up workspace since it wasn't empty")
		h.state.workspace = strings.TrimPrefix(workspace, "file://")
		go func() {
			h.GetSymbols()
			h.state.log.Sugar().Info("finished parsing workspace, have symbols:")
			h.state.log.Sugar().Info(h.state.modules, h.state.defines)
		}()
		h.state.log.Sugar().Info("finished parsing workspace")
	}

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			CompletionProvider:     &protocol.CompletionOptions{},
			DefinitionProvider:     true,
			DeclarationProvider:    true,
			ImplementationProvider: true,
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "vlsp",
			Version: "0.1.0",
		},
	}, nil
}
