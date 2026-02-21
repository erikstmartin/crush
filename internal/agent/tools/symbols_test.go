package tools

import (
	"testing"

	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
	"github.com/stretchr/testify/require"
)

func TestSymbolKindString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind protocol.SymbolKind
		want string
	}{
		{protocol.File, "File"},
		{protocol.Module, "Module"},
		{protocol.Namespace, "Namespace"},
		{protocol.Package, "Package"},
		{protocol.Class, "Class"},
		{protocol.Method, "Method"},
		{protocol.Property, "Property"},
		{protocol.Field, "Field"},
		{protocol.Constructor, "Constructor"},
		{protocol.Enum, "Enum"},
		{protocol.Interface, "Interface"},
		{protocol.Function, "Function"},
		{protocol.Variable, "Variable"},
		{protocol.Constant, "Constant"},
		{protocol.String, "String"},
		{protocol.Number, "Number"},
		{protocol.Boolean, "Boolean"},
		{protocol.Array, "Array"},
		{protocol.Object, "Object"},
		{protocol.Key, "Key"},
		{protocol.Null, "Null"},
		{protocol.EnumMember, "EnumMember"},
		{protocol.Struct, "Struct"},
		{protocol.Event, "Event"},
		{protocol.Operator, "Operator"},
		{protocol.TypeParameter, "TypeParameter"},
		{protocol.SymbolKind(999), "Kind999"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := symbolKindString(tt.kind)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCountAllSymbols(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		symbols []protocol.DocumentSymbol
		want    int
	}{
		{
			name:    "empty",
			symbols: []protocol.DocumentSymbol{},
			want:    0,
		},
		{
			name: "flat list",
			symbols: []protocol.DocumentSymbol{
				{Name: "func1", Kind: protocol.Function},
				{Name: "func2", Kind: protocol.Function},
				{Name: "var1", Kind: protocol.Variable},
			},
			want: 3,
		},
		{
			name: "nested structure",
			symbols: []protocol.DocumentSymbol{
				{
					Name: "Class1",
					Kind: protocol.Class,
					Children: []protocol.DocumentSymbol{
						{Name: "method1", Kind: protocol.Method},
						{Name: "field1", Kind: protocol.Field},
					},
				},
				{
					Name: "func1",
					Kind: protocol.Function,
				},
			},
			want: 4,
		},
		{
			name: "deeply nested",
			symbols: []protocol.DocumentSymbol{
				{
					Name: "Module1",
					Kind: protocol.Module,
					Children: []protocol.DocumentSymbol{
						{
							Name: "Class1",
							Kind: protocol.Class,
							Children: []protocol.DocumentSymbol{
								{Name: "method1", Kind: protocol.Method},
								{
									Name: "method2",
									Kind: protocol.Method,
									Children: []protocol.DocumentSymbol{
										{Name: "var1", Kind: protocol.Variable},
									},
								},
							},
						},
					},
				},
			},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := countAllSymbols(tt.symbols)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCountSymbolKinds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		symbols []protocol.DocumentSymbol
		want    map[protocol.SymbolKind]int
	}{
		{
			name:    "empty",
			symbols: []protocol.DocumentSymbol{},
			want:    map[protocol.SymbolKind]int{},
		},
		{
			name: "flat list",
			symbols: []protocol.DocumentSymbol{
				{Name: "func1", Kind: protocol.Function},
				{Name: "func2", Kind: protocol.Function},
				{Name: "var1", Kind: protocol.Variable},
			},
			want: map[protocol.SymbolKind]int{
				protocol.Function: 2,
				protocol.Variable: 1,
			},
		},
		{
			name: "nested structure",
			symbols: []protocol.DocumentSymbol{
				{
					Name: "Class1",
					Kind: protocol.Class,
					Children: []protocol.DocumentSymbol{
						{Name: "method1", Kind: protocol.Method},
						{Name: "method2", Kind: protocol.Method},
						{Name: "field1", Kind: protocol.Field},
					},
				},
				{
					Name: "func1",
					Kind: protocol.Function,
				},
			},
			want: map[protocol.SymbolKind]int{
				protocol.Class:    1,
				protocol.Method:   2,
				protocol.Field:    1,
				protocol.Function: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := countSymbolKinds(tt.symbols)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFormatDocumentSymbols(t *testing.T) {
	t.Parallel()

	symbols := []protocol.DocumentSymbol{
		{
			Name: "main",
			Kind: protocol.Function,
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
			},
		},
		{
			Name: "helper",
			Kind: protocol.Function,
			Range: protocol.Range{
				Start: protocol.Position{Line: 5, Character: 0},
			},
		},
	}

	output := formatDocumentSymbols("test.go", symbols, 100)

	require.Contains(t, output, "File: test.go")
	require.Contains(t, output, "<document_symbols>")
	require.Contains(t, output, "</document_symbols>")
	require.Contains(t, output, "<summary>")
	require.Contains(t, output, "</summary>")
	require.Contains(t, output, "main")
	require.Contains(t, output, "helper")
	require.Contains(t, output, "Function")
	require.Contains(t, output, "Found 2 symbols")
}

func TestFormatDocumentSymbolsWithLimit(t *testing.T) {
	t.Parallel()

	symbols := []protocol.DocumentSymbol{
		{Name: "func1", Kind: protocol.Function, Range: protocol.Range{Start: protocol.Position{Line: 0}}},
		{Name: "func2", Kind: protocol.Function, Range: protocol.Range{Start: protocol.Position{Line: 1}}},
		{Name: "func3", Kind: protocol.Function, Range: protocol.Range{Start: protocol.Position{Line: 2}}},
		{Name: "func4", Kind: protocol.Function, Range: protocol.Range{Start: protocol.Position{Line: 3}}},
		{Name: "func5", Kind: protocol.Function, Range: protocol.Range{Start: protocol.Position{Line: 4}}},
	}

	output := formatDocumentSymbols("test.go", symbols, 3)

	require.Contains(t, output, "Found 5 symbols (showing first 3)")
	require.Contains(t, output, "func1")
	require.Contains(t, output, "func2")
	require.Contains(t, output, "func3")
}

func TestFormatWorkspaceSymbols(t *testing.T) {
	t.Parallel()

	symbols := []protocol.SymbolInformation{
		{
			Name: "TestFunc",
			Kind: protocol.Function,
			Location: protocol.Location{
				URI: protocol.DocumentURI("file:///test.go"),
				Range: protocol.Range{
					Start: protocol.Position{Line: 9, Character: 4},
				},
			},
		},
		{
			Name: "HelperFunc",
			Kind: protocol.Function,
			Location: protocol.Location{
				URI: protocol.DocumentURI("file:///helper.go"),
				Range: protocol.Range{
					Start: protocol.Position{Line: 19, Character: 9},
				},
			},
			ContainerName: "package main",
		},
	}

	output := formatWorkspaceSymbols("Func", symbols, 100)

	require.Contains(t, output, "Workspace Symbols matching 'Func'")
	require.Contains(t, output, "<workspace_symbols>")
	require.Contains(t, output, "</workspace_symbols>")
	require.Contains(t, output, "TestFunc")
	require.Contains(t, output, "HelperFunc")
	require.Contains(t, output, "line 10:5")
	require.Contains(t, output, "line 20:10")
	require.Contains(t, output, "in package main")
	require.Contains(t, output, "Found 2 symbols")
}

func TestFormatWorkspaceSymbolsWithLimit(t *testing.T) {
	t.Parallel()

	symbols := []protocol.SymbolInformation{
		{Name: "Func1", Kind: protocol.Function, Location: protocol.Location{URI: protocol.DocumentURI("file:///a.go")}},
		{Name: "Func2", Kind: protocol.Function, Location: protocol.Location{URI: protocol.DocumentURI("file:///b.go")}},
		{Name: "Func3", Kind: protocol.Function, Location: protocol.Location{URI: protocol.DocumentURI("file:///c.go")}},
	}

	output := formatWorkspaceSymbols("Func", symbols, 2)

	require.Contains(t, output, "Found 3 symbols (showing first 2)")
	require.Contains(t, output, "Func1")
	require.Contains(t, output, "Func2")
}
