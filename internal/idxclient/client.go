package idxclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// IndexerClient communicates with a deployed or local cb-indexer HTTP MCP server.
type IndexerClient struct {
	Endpoint   string
	AuthToken  string
	HTTPClient *http.Client
}

// NewIndexerClient creates a new client pointing to the indexer endpoint.
func NewIndexerClient(baseURL, authToken string) *IndexerClient {
	endpoint := baseURL
	if endpoint == "" {
		endpoint = os.Getenv("CB_INDEXER_URL")
	}
	if endpoint == "" {
		endpoint = os.Getenv("CB_INDEXER_URL")
	}
	if endpoint == "" {
		endpoint = "http://127.0.0.1:43770"
	}
	endpoint = strings.TrimSuffix(endpoint, "/")
	if !strings.HasSuffix(endpoint, "/mcp") {
		endpoint += "/mcp"
	}

	token := authToken
	if token == "" {
		token = os.Getenv("CB_INDEXER_AUTH_TOKEN")
	}
	if token == "" {
		token = os.Getenv("CB_INDEXER_AUTH_TOKEN")
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
	req.Header.Set("Accept", "application/json, text/event-stream")
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("cb-indexer is unreachable at %s (%w). Ensure cb-indexer daemon is running", c.Endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized request to cb-indexer (invalid auth token)")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cb-indexer returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(respData, &rpcResp); err != nil {
		return fmt.Errorf("invalid JSON-RPC response from cb-indexer: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("cb-indexer RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if rpcResp.Result.IsError {
		errMsg := "tool execution failed"
		if len(rpcResp.Result.Content) > 0 {
			errMsg = rpcResp.Result.Content[0].Text
		}
		return fmt.Errorf("cb-indexer tool error: %s", errMsg)
	}

	if len(rpcResp.Result.Content) == 0 {
		return fmt.Errorf("empty tool result content from cb-indexer")
	}

	text := rpcResp.Result.Content[0].Text
	if target != nil {
		if err := json.Unmarshal([]byte(text), target); err != nil {
			return fmt.Errorf("failed to unmarshal tool output into DTO: %w (raw: %s)", err, text)
		}
	}

	return nil
}

func (c *IndexerClient) getJSON(ctx context.Context, apiPath string, target any) error {
	base := strings.TrimSuffix(c.Endpoint, "/mcp")
	base = strings.TrimSuffix(base, "/")
	reqURL := base + apiPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
		req.Header.Set("X-API-Key", c.AuthToken)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

// GetArchitectureOverview calls get_architecture_overview on cb-indexer.
func (c *IndexerClient) GetArchitectureOverview(ctx context.Context, project string) (*ArchitectureOverviewDTO, error) {
	var dto ArchitectureOverviewDTO
	apiPath := "/api/overview"
	if project != "" {
		apiPath += "?project=" + url.QueryEscape(project)
	}
	if err := c.getJSON(ctx, apiPath, &dto); err == nil && dto.ProjectID != "" {
		return &dto, nil
	}

	args := make(map[string]any)
	if project != "" {
		args["project"] = project
	}
	if err := c.callTool(ctx, "get_architecture_overview", args, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}

// GetRepoDetails calls get_repo_details on cb-indexer.
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

// GetRelatedRepos calls get_related_repos on cb-indexer.
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

// ListProjects calls list_projects on cb-indexer.
func (c *IndexerClient) ListProjects(ctx context.Context, project string) (*ProjectListDTO, error) {
	var dto ProjectListDTO
	if err := c.getJSON(ctx, "/api/projects", &dto); err == nil && dto.TotalRegisteredProjects >= 0 {
		return &dto, nil
	}

	args := make(map[string]any)
	if project != "" {
		args["project"] = project
	}
	if err := c.callTool(ctx, "list_projects", args, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}

// QueryCodebaseSymbols calls query_codebase_symbols on remote cb-indexer.
func (c *IndexerClient) QueryCodebaseSymbols(ctx context.Context, query, repoName, label string, limit int) (*SymbolSearchDTO, error) {
	var dto SymbolSearchDTO
	params := url.Values{}
	params.Set("q", query)
	if repoName != "" {
		params.Set("repo", repoName)
	}
	if label != "" {
		params.Set("label", label)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if err := c.getJSON(ctx, "/api/rag/search?"+params.Encode(), &dto); err == nil && dto.Total >= 0 {
		return &dto, nil
	}

	args := map[string]any{"query": query}
	if repoName != "" {
		args["repo_name"] = repoName
	}
	if label != "" {
		args["label"] = label
	}
	if limit > 0 {
		args["limit"] = limit
	}
	if err := c.callTool(ctx, "query_codebase_symbols", args, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}

// GetSymbolContext calls get_symbol_context on remote cb-indexer.
func (c *IndexerClient) GetSymbolContext(ctx context.Context, repoName, filePath string, startLine, endLine int, project string) (*CodeSnippetDTO, error) {
	var dto CodeSnippetDTO
	params := url.Values{}
	params.Set("repo", repoName)
	params.Set("file", filePath)
	if startLine > 0 {
		params.Set("start", strconv.Itoa(startLine))
	}
	if endLine > 0 {
		params.Set("end", strconv.Itoa(endLine))
	}
	if project != "" {
		params.Set("project", project)
	}
	var res struct {
		Context CodeSnippetDTO `json:"context"`
	}
	if err := c.getJSON(ctx, "/api/rag/context?"+params.Encode(), &res); err == nil && res.Context.FilePath != "" {
		return &res.Context, nil
	}

	args := map[string]any{
		"repo_name": repoName,
		"file_path": filePath,
	}
	if startLine > 0 {
		args["start_line"] = startLine
	}
	if endLine > 0 {
		args["end_line"] = endLine
	}
	if project != "" {
		args["project"] = project
	}
	if err := c.callTool(ctx, "get_symbol_context", args, &dto); err != nil {
		return nil, err
	}
	return &dto, nil
}
