package vlsp

import (
	"context"

	"github.com/chrehall68/vls/internal/mappers"
	"go.lsp.dev/protocol"
)

func (h Handler) Completion(ctx context.Context, params *protocol.CompletionParams) (result *protocol.CompletionList, err error) {
	h.state.log.Sugar().Infof("Completion called")
	var completionItems []protocol.CompletionItem

	for word, emoji := range mappers.EmojiMapper {
		emojiCopy := emoji // Create a copy of emoji
		completionItems = append(completionItems, protocol.CompletionItem{
			Label:      word,
			Detail:     emojiCopy,
			InsertText: emojiCopy,
		})
	}
	for _, defines := range h.state.defines {
		for _, define := range defines {
			completionItems = append(completionItems, protocol.CompletionItem{
				Label:      "`" + define.Identifier.Value,
				Detail:     "define",
				InsertText: "`" + define.Identifier.Value,
			})
		}
	}
	for _, modules := range h.state.modules {
		for _, module := range modules {
			completionItems = append(completionItems, protocol.CompletionItem{
				Label:      module.Identifier.Value,
				Detail:     "module",
				InsertText: module.Identifier.Value,
			})
		}
	}
	//h.state.log.Sugar().Infof("completionItems: %v", completionItems)

	return &protocol.CompletionList{Items: completionItems, IsIncomplete: true}, nil
}
