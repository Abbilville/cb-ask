package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cb-ask/internal/cbmclient"
	"cb-ask/internal/guides"
	"cb-ask/internal/idxclient"
	"cb-ask/internal/workflows"
)

type TraceFlowInput struct {
	Query      string `json:"query" jsonschema:"Description of the cross-service request or feature to trace"`
	SourceRepo string `json:"source_repo,omitempty" jsonschema:"Optional starting repository (caller/client)"`
	TargetRepo string `json:"target_repo,omitempty" jsonschema:"Optional destination repository (callee/backend)"`
	Project    string `json:"project,omitempty" jsonschema:"Optional project ID or registry path"`
}

type GuideInput struct {
	GuideName string `json:"guide_name" jsonschema:"Name of the workflow guide: 'oss' or 'oss-navigator'"`
}

type OverviewInput struct {
	Project string `json:"project,omitempty" jsonschema:"Optional project ID or registry path"`
}

type QuerySymbolsInput struct {
	Query    string `json:"query" jsonschema:"Symbol name or pattern to search (e.g. 'CheckoutHandler', 'getUser')"`
	RepoName string `json:"repo_name,omitempty" jsonschema:"Optional repository name to filter"`
	Label    string `json:"label,omitempty" jsonschema:"Optional AST node label ('Function', 'Method', 'Struct', 'Class')"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Max results (default: 20)"`
}

type SymbolContextInput struct {
	RepoName  string `json:"repo_name" jsonschema:"Repository name containing the file"`
	FilePath  string `json:"file_path" jsonschema:"Relative file path inside the repository"`
	StartLine int    `json:"start_line,omitempty" jsonschema:"Start line number"`
	EndLine   int    `json:"end_line,omitempty" jsonschema:"End line number"`
	Project   string `json:"project,omitempty" jsonschema:"Optional project ID"`
}

func RegisterTools(s *mcp.Server, idx *idxclient.IndexerClient, cbm *cbmclient.CbmClient) {
	// Tool 1: trace_cross_service_flow
	mcp.AddTool(s, &mcp.Tool{
		Name:        "trace_cross_service_flow",
		Description: "Trace end-to-end cross-service communication by merging topology from cb-indexer with AST graph symbols from codebase-memory-mcp.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TraceFlowInput) (*mcp.CallToolResult, any, error) {
		report, err := workflows.TraceCrossServiceFlow(ctx, idx, cbm, input.Query, input.SourceRepo, input.TargetRepo, input.Project)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(report), nil, nil
	})

	// Tool 2: get_workflow_guide
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_workflow_guide",
		Description: "Retrieve architectural navigation and multi-repository workflow instructions ('oss' or 'oss-navigator').",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GuideInput) (*mcp.CallToolResult, any, error) {
		guideText, err := guides.GetWorkflowGuide(input.GuideName)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: guideText,
				},
			},
		}, nil, nil
	})

	// Tool 3: get_architecture_overview
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_architecture_overview",
		Description: "Fetch the complete microservice architecture topology, listening ports, tech stacks, and dependency edges.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input OverviewInput) (*mcp.CallToolResult, any, error) {
		overview, err := idx.GetArchitectureOverview(ctx, input.Project)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(overview), nil, nil
	})

	// Tool 4: query_codebase_symbols
	mcp.AddTool(s, &mcp.Tool{
		Name:        "query_codebase_symbols",
		Description: "Search indexed AST symbols across microservices (functions, methods, classes, structs) via the knowledge graph.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input QuerySymbolsInput) (*mcp.CallToolResult, any, error) {
		limit := input.Limit
		if limit <= 0 {
			limit = 20
		}
		res, err := idx.QueryCodebaseSymbols(ctx, input.Query, input.RepoName, input.Label, limit)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(res), nil, nil
	})

	// Tool 5: get_symbol_context
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_symbol_context",
		Description: "Retrieve source code context snippet around a file and line range from the repository.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SymbolContextInput) (*mcp.CallToolResult, any, error) {
		res, err := idx.GetSymbolContext(ctx, input.RepoName, input.FilePath, input.StartLine, input.EndLine, input.Project)
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(res), nil, nil
	})
}

func jsonResult(data any) *mcp.CallToolResult {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errorResult(err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(bytes),
			},
		},
	}
}

func errorResult(err error) *mcp.CallToolResult {
	errMap := map[string]string{"error": fmt.Sprintf("%v", err)}
	bytes, _ := json.MarshalIndent(errMap, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(bytes),
			},
		},
		IsError: true,
	}
}
