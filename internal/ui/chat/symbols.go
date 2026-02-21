package chat

import (
	"encoding/json"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/fsext"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/styles"
)

// SymbolsToolMessageItem is a message item that represents a symbols tool call.
type SymbolsToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*SymbolsToolMessageItem)(nil)

// NewSymbolsToolMessageItem creates a new [SymbolsToolMessageItem].
func NewSymbolsToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &SymbolsToolRenderContext{}, canceled)
}

// SymbolsToolRenderContext renders symbols tool messages.
type SymbolsToolRenderContext struct{}

// RenderTool implements the [ToolRenderer] interface.
func (s *SymbolsToolRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	if opts.IsPending() {
		return pendingTool(sty, "LSP Symbols", opts.Anim)
	}

	var params tools.SymbolsParams
	_ = json.Unmarshal([]byte(opts.ToolCall.Input), &params)

	var toolParams []string
	if params.FilePath != "" {
		toolParams = append(toolParams, fsext.PrettyPath(params.FilePath))
	} else if params.Query != "" {
		toolParams = append(toolParams, params.Query)
	}

	header := toolHeader(sty, opts.Status, "LSP Symbols", cappedWidth, opts.Compact, toolParams...)
	if opts.Compact {
		return header
	}

	if earlyState, ok := toolEarlyStateContent(sty, opts, cappedWidth); ok {
		return joinToolParts(header, earlyState)
	}

	if opts.HasEmptyResult() {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputPlainContent(sty, opts.Result.Content, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}
