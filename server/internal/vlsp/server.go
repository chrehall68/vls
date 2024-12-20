package vlsp

import (
	"context"

	"github.com/chrehall68/vls/internal/lang"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

type ServerState struct {
	workspace           string
	modules             map[string][]lang.ModuleNode              // list of all modules, grouped by file (w/o the file://)
	defines             map[string][]lang.DefineNode              // list of all defines, grouped by file (w/o the file://)
	symbolMap           map[string]protocol.Location              // map of symbol names to their location (path w/ the file://)
	files               map[string]*File                          // map of file names (w/o the file://) to corresponding File objects
	variableDefinitions map[string](map[string]protocol.Location) // map of module name : (variable name: declaration)
	log                 *zap.Logger
	stream              *jsonrpc2.Stream
	client              protocol.Client
}

type Handler struct {
	protocol.Server
	state *ServerState
}

func (h Handler) String() string {
	return "vlsp"
}

func NewHandler(ctx context.Context, server protocol.Server, client protocol.Client, logger *zap.Logger, stream *jsonrpc2.Stream) (Handler, context.Context, error) {
	// Do initialization logic here, including
	// stuff like setting state variables
	// by returning a new context with
	// context.WithValue(context, ...)
	// instead of just context
	return Handler{
		Server: server,
		state: &ServerState{
			workspace:           "",
			modules:             map[string][]lang.ModuleNode{},
			defines:             map[string][]lang.DefineNode{},
			variableDefinitions: map[string](map[string]protocol.Location){},
			log:                 logger,
			stream:              stream,
			client:              client,
			files:               map[string]*File{}}}, ctx, nil
}

func (h Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	h.state.log.Sugar().Infof("Initialize called")

	workspace := params.WorkspaceFolders[0].URI
	h.state.log.Sugar().Infof("workspace: %v", workspace)
	if workspace != "" {
		h.state.log.Sugar().Info("setting up workspace since it wasn't empty")
		h.state.workspace = URIToPath(workspace)
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
			TextDocumentSync:       protocol.TextDocumentSyncKindFull,
			SemanticTokensProvider: GetSemanticTokensOptions(),
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "vlsp",
			Version: "0.1.0",
		},
	}, nil
}
