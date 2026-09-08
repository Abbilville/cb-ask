package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cb-ask/internal/cbmclient"
	"cb-ask/internal/idxclient"
)

// NewServer creates and initializes a cb-ask MCP server instance.
func NewServer(idx *idxclient.IndexerClient, cbm *cbmclient.CbmClient) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "cb-ask",
		Version: "0.1.0",
	}, nil)

	RegisterTools(s, idx, cbm)
	return s
}

// ServeStdio starts the MCP server over standard input/output for AI agent interaction.
func ServeStdio(ctx context.Context, s *mcp.Server) error {
	return s.Run(ctx, &mcp.StdioTransport{})
}
