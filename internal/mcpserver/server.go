package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"oss-ask/internal/cbmclient"
	"oss-ask/internal/idxclient"
)

// NewServer creates and initializes an oss-ask MCP server instance.
func NewServer(idx *idxclient.IndexerClient, cbm *cbmclient.CbmClient) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "oss-ask",
		Version: "0.2.0",
	}, nil)

	RegisterTools(s, idx, cbm)
	return s
}

// ServeStdio starts the MCP server over standard input/output for AI agent interaction.
func ServeStdio(ctx context.Context, s *mcp.Server) error {
	return s.Run(ctx, &mcp.StdioTransport{})
}
