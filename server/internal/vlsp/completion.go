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

	// global-level completions
	for word, emoji := range mappers.EmojiMapper {
		completionItems = append(completionItems, protocol.CompletionItem{
			Label:      word,
			Detail:     emoji,
			InsertText: emoji,
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
	for _, keyword := range lang.Keywords {
		completionItems = append(completionItems, protocol.CompletionItem{
			Label:      keyword,
			Detail:     "keyword",
			InsertText: keyword,
		})
	}
	for snippetName, snippet := range lang.Snippets {
		completionItems = append(completionItems, protocol.CompletionItem{
			Label:            snippetName,
			Detail:           "snippet",
			InsertText:       snippet,
			InsertTextFormat: protocol.InsertTextFormatSnippet,
		})
	}

	// local-level completions
	details, err := h.getLocationDetails(URIToPath(string(params.TextDocument.URI)), int(params.Position.Line), int(params.Position.Character))
	if err == nil {
		for name := range h.state.variableDefinitions[details.currentModule] {
			completionItems = append(completionItems, protocol.CompletionItem{
				Label:      name,
				Detail:     "variable",
				InsertText: name,
			})
		}
	}
	//h.state.log.Sugar().Infof("completionItems: %v", completionItems)

	return &protocol.CompletionList{Items: completionItems, IsIncomplete: true}, nil
}
