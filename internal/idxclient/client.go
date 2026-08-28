package idxclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// IndexerClient communicates with a deployed or local oss-indexer HTTP MCP server.
type IndexerClient struct {
	Endpoint   string
	AuthToken  string
	HTTPClient *http.Client
}

// NewIndexerClient creates a new client pointing to the indexer endpoint.
func NewIndexerClient(baseURL, authToken string) *IndexerClient {
	endpoint := baseURL
	if endpoint == "" {
		endpoint = os.Getenv("OSS_INDEXER_URL")
	}
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080"
	}
	endpoint = strings.TrimSuffix(endpoint, "/")
	if !strings.HasSuffix(endpoint, "/mcp") {
		endpoint += "/mcp"
	}

	token := authToken
	if token == "" {
		token = os.Getenv("OSS_INDEXER_AUTH_TOKEN")
	}

	return &IndexerClient{
		Endpoint:  endpoint,
		AuthToken: token,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type jsonrpcRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
}

type jsonrpcResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *IndexerClient) callTool(ctx context.Context, toolName string, args map[string]any, target any) error {
	reqBody := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("oss-indexer is unreachable at %s (%w). Ensure oss-indexer daemon is running", c.Endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized request to oss-indexer (invalid auth token)")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("oss-indexer returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(respData, &rpcResp); err != nil {
		return fmt.Errorf("invalid JSON-RPC response from oss-indexer: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("oss-indexer RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if rpcResp.Result.IsError {
		errMsg := "tool execution failed"
		if len(rpcResp.Result.Content) > 0 {
			errMsg = rpcResp.Result.Content[0].Text
		}
		return fmt.Errorf("oss-indexer tool error: %s", errMsg)
	}

	if len(rpcResp.Result.Content) == 0 {
		return fmt.Errorf("empty tool result content from oss-indexer")
	}

	text := rpcResp.Result.Content[0].Text
	if target != nil {
		if err := json.Unmarshal([]byte(text), target); err != nil {
			return fmt.Errorf("failed to unmarshal tool output into DTO: %w (raw: %s)", err, text)
		}
	}

	return nil
}

// GetArchitectureOverview calls get_architecture_overview on oss-indexer.
func (c *IndexerClient) GetArchitectureOverview(ctx context.Context, project string) (*ArchitectureOverviewDTO, error) {
	args := make(map[string]any)
	if project != "" {
		args["project"] = project
	}
	var dto ArchitectureOverviewDTO
	if err := c.callTool(ctx, "get_architecture_overview", args, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}

// GetRepoDetails calls get_repo_details on oss-indexer.
func (c *IndexerClient) GetRepoDetails(ctx context.Context, repoName, project string) (*RepoDetailsDTO, error) {
	args := map[string]any{"repo_name": repoName}
	if project != "" {
		args["project"] = project
	}
	var dto RepoDetailsDTO
	if err := c.callTool(ctx, "get_repo_details", args, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}

// GetRelatedRepos calls get_related_repos on oss-indexer.
func (c *IndexerClient) GetRelatedRepos(ctx context.Context, repoName, direction, project string) (*RelatedReposDTO, error) {
	args := map[string]any{"repo_name": repoName}
	if direction != "" {
		args["direction"] = direction
	}
	if project != "" {
		args["project"] = project
	}
	var dto RelatedReposDTO
	if err := c.callTool(ctx, "get_related_repos", args, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}

// ListProjects calls list_projects on oss-indexer.
func (c *IndexerClient) ListProjects(ctx context.Context, project string) (*ProjectListDTO, error) {
	args := make(map[string]any)
	if project != "" {
		args["project"] = project
	}
	var dto ProjectListDTO
	if err := c.callTool(ctx, "list_projects", args, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}
