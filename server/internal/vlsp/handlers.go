package vlsp

import (
	"context"

	"github.com/chrehall68/vls/internal/mappers"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

var log *zap.Logger

type Handler struct {
	protocol.Server
}

func NewHandler(ctx context.Context, server protocol.Server, logger *zap.Logger) (Handler, context.Context, error) {
	log = logger
	// Do initialization logic here, including
	// stuff like setting state variables
	// by returning a new context with
	// context.WithValue(context, ...)
	// instead of just context
	return Handler{Server: server}, ctx, nil
}
func (h Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	log.Sugar().Infof("Initialize called")
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			CompletionProvider: &protocol.CompletionOptions{},
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "vlsp",
			Version: "0.1.0",
		},
	}, nil
}

func (h Handler) Completion(ctx context.Context, params *protocol.CompletionParams) (result *protocol.CompletionList, err error) {
	log.Sugar().Infof("Completion called")
	var completionItems []protocol.CompletionItem

	for word, emoji := range mappers.EmojiMapper {
		emojiCopy := emoji // Create a copy of emoji
		completionItems = append(completionItems, protocol.CompletionItem{
			Label:      word,
			Detail:     emojiCopy,
			InsertText: emojiCopy,
		})
	}
	log.Sugar().Infof("completionItems: %v", completionItems)

	return &protocol.CompletionList{Items: completionItems, IsIncomplete: true}, nil
}
