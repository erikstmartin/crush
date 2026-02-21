package tools

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/lsp"
	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
)

type SymbolsParams struct {
	Query    string `json:"query,omitempty" description:"Symbol name or pattern to search for workspace symbols. Use empty string or '*' to get all symbols. Leave empty and provide file_path for document symbols."`
	FilePath string `json:"file_path,omitempty" description:"File path for document symbols. Provide this to get symbols from a specific file."`
	Limit    int    `json:"limit,omitempty" description:"Maximum number of results to return (default: 50, max: 200)"`
}

const SymbolsToolName = "lsp_symbols"

//go:embed symbols.md
var symbolsDescription []byte

func NewSymbolsTool(lspManager *lsp.Manager) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		SymbolsToolName,
		string(symbolsDescription),
		func(ctx context.Context, params SymbolsParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			limit := params.Limit
			if limit == 0 {
				limit = 50
			}
			if limit > 200 {
				limit = 200
			}

			if params.FilePath != "" {
				return getDocumentSymbols(ctx, lspManager, params.FilePath, limit)
			}

			return getWorkspaceSymbols(ctx, lspManager, params.Query, limit)
		})
}

func getDocumentSymbols(ctx context.Context, lspManager *lsp.Manager, filePath string, limit int) (fantasy.ToolResponse, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to get absolute path: %s", err)), nil
	}

	lspManager.Start(ctx, absPath)

	var client *lsp.Client
	for c := range lspManager.Clients().Seq() {
		if c.HandlesFile(absPath) {
			client = c
			break
		}
	}

	if client == nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("no LSP client available for file: %s", filePath)), nil
	}

	if err := client.OpenFileOnDemand(ctx, absPath); err != nil {
		slog.Warn("Failed to open file in LSP", "error", err, "file", absPath)
	}

	symbols, err := client.DocumentSymbols(ctx, absPath)
	if err != nil {
		slog.Error("Failed to get document symbols", "error", err, "file", absPath)
		return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to get document symbols: %s", err)), nil
	}

	if len(symbols) == 0 {
		return fantasy.NewTextResponse(fmt.Sprintf("No symbols found in %s", filePath)), nil
	}

	output := formatDocumentSymbols(filePath, symbols, limit)
	return fantasy.NewTextResponse(output), nil
}

func getWorkspaceSymbols(ctx context.Context, lspManager *lsp.Manager, query string, limit int) (fantasy.ToolResponse, error) {
	lspManager.StartAll(ctx)

	var allSymbols []protocol.SymbolInformation

	for client := range lspManager.Clients().Seq() {
		symbols, err := client.WorkspaceSymbols(ctx, query)
		if err != nil {
			slog.Warn("Failed to get workspace symbols from client", "error", err, "client", client.GetName())
			continue
		}
		allSymbols = append(allSymbols, symbols...)
	}

	if len(allSymbols) == 0 {
		return fantasy.NewTextResponse(fmt.Sprintf("No symbols found matching '%s'", query)), nil
	}

	sort.Slice(allSymbols, func(i, j int) bool {
		if allSymbols[i].Location.URI != allSymbols[j].Location.URI {
			return allSymbols[i].Location.URI < allSymbols[j].Location.URI
		}
		return allSymbols[i].Location.Range.Start.Line < allSymbols[j].Location.Range.Start.Line
	})

	output := formatWorkspaceSymbols(query, allSymbols, limit)
	return fantasy.NewTextResponse(output), nil
}

func formatDocumentSymbols(filePath string, symbols []protocol.DocumentSymbol, limit int) string {
	var output strings.Builder

	fmt.Fprintf(&output, "File: %s\n\n", filePath)
	output.WriteString("<document_symbols>\n")

	total := countAllSymbols(symbols)
	truncated := total > limit

	count := 0
	formatSymbolTree(&output, symbols, "", &count, limit)

	output.WriteString("</document_symbols>\n\n")

	kindCounts := countSymbolKinds(symbols)
	output.WriteString("<summary>\n")
	fmt.Fprintf(&output, "Found %d symbols", total)
	if truncated {
		fmt.Fprintf(&output, " (showing first %d)", limit)
	}
	output.WriteString(": ")

	var kinds []string
	for kind, count := range kindCounts {
		kinds = append(kinds, fmt.Sprintf("%d %s", count, symbolKindString(kind)))
	}
	output.WriteString(strings.Join(kinds, ", "))
	output.WriteString("\n</summary>\n")

	return output.String()
}

func formatWorkspaceSymbols(query string, symbols []protocol.SymbolInformation, limit int) string {
	var output strings.Builder

	total := len(symbols)
	if total > limit {
		symbols = symbols[:limit]
	}

	if query == "" || query == "*" {
		fmt.Fprintf(&output, "Workspace Symbols (all)\n\n")
	} else {
		fmt.Fprintf(&output, "Workspace Symbols matching '%s'\n\n", query)
	}
	output.WriteString("<workspace_symbols>\n")

	currentFile := ""
	for _, sym := range symbols {
		path, err := sym.Location.URI.Path()
		if err != nil {
			slog.Error("Failed to convert URI to path", "uri", sym.Location.URI, "error", err)
			continue
		}

		if path != currentFile {
			if currentFile != "" {
				output.WriteString("\n")
			}
			fmt.Fprintf(&output, "%s:\n", path)
			currentFile = path
		}

		line := sym.Location.Range.Start.Line + 1
		char := sym.Location.Range.Start.Character + 1
		kindStr := symbolKindString(sym.Kind)

		containerInfo := ""
		if sym.ContainerName != "" {
			containerInfo = fmt.Sprintf(" in %s", sym.ContainerName)
		}

		fmt.Fprintf(&output, "  - %s %s%s [line %d:%d]\n",
			kindStr, sym.Name, containerInfo, line, char)
	}

	output.WriteString("</workspace_symbols>\n\n")
	output.WriteString("<summary>\n")
	fmt.Fprintf(&output, "Found %d symbols", total)
	if total > limit {
		fmt.Fprintf(&output, " (showing first %d)", limit)
	}
	output.WriteString("\n</summary>\n")

	return output.String()
}

func formatSymbolTree(output *strings.Builder, symbols []protocol.DocumentSymbol, prefix string, count *int, limit int) {
	for i, sym := range symbols {
		if *count >= limit {
			return
		}

		isLast := i == len(symbols)-1
		connector := "├─"
		childPrefix := "│  "
		if isLast {
			connector = "└─"
			childPrefix = "   "
		}

		line := sym.Range.Start.Line + 1
		char := sym.Range.Start.Character + 1
		kindStr := symbolKindString(sym.Kind)

		detail := ""
		if sym.Detail != "" {
			detail = fmt.Sprintf(" %s", sym.Detail)
		}

		fmt.Fprintf(output, "%s%s %s %s (%s)%s [line %d:%d]\n",
			prefix, connector, kindStr, sym.Name, kindStr, detail, line, char)

		*count++

		if len(sym.Children) > 0 && *count < limit {
			formatSymbolTree(output, sym.Children, prefix+childPrefix, count, limit)
		}
	}
}

func countAllSymbols(symbols []protocol.DocumentSymbol) int {
	count := len(symbols)
	for _, sym := range symbols {
		count += countAllSymbols(sym.Children)
	}
	return count
}

func countSymbolKinds(symbols []protocol.DocumentSymbol) map[protocol.SymbolKind]int {
	counts := make(map[protocol.SymbolKind]int)
	for _, sym := range symbols {
		counts[sym.Kind]++
		for kind, c := range countSymbolKinds(sym.Children) {
			counts[kind] += c
		}
	}
	return counts
}

func symbolKindString(kind protocol.SymbolKind) string {
	switch kind {
	case protocol.File:
		return "File"
	case protocol.Module:
		return "Module"
	case protocol.Namespace:
		return "Namespace"
	case protocol.Package:
		return "Package"
	case protocol.Class:
		return "Class"
	case protocol.Method:
		return "Method"
	case protocol.Property:
		return "Property"
	case protocol.Field:
		return "Field"
	case protocol.Constructor:
		return "Constructor"
	case protocol.Enum:
		return "Enum"
	case protocol.Interface:
		return "Interface"
	case protocol.Function:
		return "Function"
	case protocol.Variable:
		return "Variable"
	case protocol.Constant:
		return "Constant"
	case protocol.String:
		return "String"
	case protocol.Number:
		return "Number"
	case protocol.Boolean:
		return "Boolean"
	case protocol.Array:
		return "Array"
	case protocol.Object:
		return "Object"
	case protocol.Key:
		return "Key"
	case protocol.Null:
		return "Null"
	case protocol.EnumMember:
		return "EnumMember"
	case protocol.Struct:
		return "Struct"
	case protocol.Event:
		return "Event"
	case protocol.Operator:
		return "Operator"
	case protocol.TypeParameter:
		return "TypeParameter"
	default:
		return fmt.Sprintf("Kind%d", kind)
	}
}
