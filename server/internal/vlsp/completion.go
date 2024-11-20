package vlsp

import (
	"context"
	"fmt"
	"strings"

	"github.com/chrehall68/vls/internal/lang"
	"github.com/chrehall68/vls/internal/mappers"
	"go.lsp.dev/protocol"
)

func (h Handler) formatModuleApplication(module lang.ModuleNode) string {
	params := make([]string, len(module.PortList.Ports))
	i := 2
	for _, param := range module.PortList.Ports {
		params[i-2] = fmt.Sprintf(".%s($%d)", param.Value, i)
		i++
	}
	return fmt.Sprintf("%s ${1:name}(%s);", module.Identifier.Value, strings.Join(params, ", "))
}

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
				Label:            module.Identifier.Value,
				Detail:           "module",
				InsertText:       h.formatModuleApplication(module),
				InsertTextFormat: protocol.InsertTextFormatSnippet,
			})
		}
	}
	//h.state.log.Sugar().Infof("completionItems: %v", completionItems)

	return &protocol.CompletionList{Items: completionItems, IsIncomplete: true}, nil
}
