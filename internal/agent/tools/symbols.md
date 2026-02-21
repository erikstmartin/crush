Find symbol definitions (functions, classes, types, methods, variables, constants) using the Language Server Protocol (LSP).

<usage>
- Provide file_path to get document symbols (outline/structure) for a specific file
- Provide query to search for symbols across the workspace (e.g., "BuildTools", "Manager", "handleRequest")
- Use empty string or "*" for query to get all workspace symbols
- Omit both parameters to get all workspace symbols with no filtering
- Use limit to control number of results (default: 50, max: 200)
</usage>

<features>
- Semantic-aware symbol search (more accurate than grep/glob/view)
- Shows functions, classes, methods, variables, constants, interfaces, structs, and more
- Displays symbol hierarchy (nested symbols like methods within classes)
- Includes precise location information (file, line, column)
- Supports multiple programming languages via LSP (Go, TypeScript, Python, Rust, etc.)
- 90%+ faster than reading entire files to find structure
</features>

<when_to_use>
**Use this tool first for questions about:**
- "What's the structure of X file?" → lsp_symbols(file_path="X")
- "Is there a X function/class/type?" → lsp_symbols(query="X")
- "What functions are in X file?" → lsp_symbols(file_path="X")
- "What methods does X type have?" → lsp_symbols(query="X")
- "What types/interfaces are in X package?" → lsp_symbols(query="package_name")
- "Show me the API surface of X" → lsp_symbols(file_path="X" or query="X")
- "What's defined in this codebase?" → lsp_symbols(query="")

**Do NOT use this tool for:**
- "List all interfaces" → use ast_grep_search(pattern="type $NAME interface { $$$ }", lang="go") instead
- "List all types" → use ast_grep_search or grep with pattern matching instead
- "Find all structs" → use ast_grep_search(pattern="type $NAME struct { $$$ }", lang="go") instead
- Any query requiring filtering by symbol kind (interface, struct, class, etc.) → LSP workspace/symbol does NOT support kind filtering
</when_to_use>

<limitations>
- Requires LSP client for the file's language
- Results depend on LSP server capabilities
- Some LSP servers may not support workspace symbols
- Hierarchical structure only available for document symbols
- **Cannot filter by symbol kind** (e.g., "list all interfaces" or "list all types") - use ast_grep or grep for kind-based filtering
</limitations>

<tips>
- Use this first when searching for symbol definitions (functions, types, classes, methods).
- Do not use view/grep/glob for finding symbols - use this tool instead.
- For file structure: lsp_symbols(file_path="...") is 90%+ faster than view
- For finding symbols: lsp_symbols(query="...") is more accurate than grep
- Combine with lsp_references to find where symbols are used (references finds usage, symbols finds definitions)
- Narrow workspace symbol searches with specific queries for better results
</tips>
