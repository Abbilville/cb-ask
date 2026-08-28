package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"oss-ask/internal/cbmclient"
	"oss-ask/internal/guides"
	"oss-ask/internal/idxclient"
	"oss-ask/internal/workflows"
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

func RegisterTools(s *mcp.Server, idx *idxclient.IndexerClient, cbm *cbmclient.CbmClient) {
	// Tool 1: trace_cross_service_flow
	mcp.AddTool(s, &mcp.Tool{
		Name:        "trace_cross_service_flow",
		Description: "Trace end-to-end cross-service communication by merging topology from oss-indexer with AST graph symbols from codebase-memory-mcp.",
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
