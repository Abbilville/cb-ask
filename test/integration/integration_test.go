package integration

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"cb-ask/internal/idxclient"
	"cb-ask/internal/workflows"
)

func TestIndexerAndQueryContractIntegration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Mock cb-indexer HTTP server that responds with exact JSON-RPC MCP format
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	serverPort := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			JSONRPC string `json:"jsonrpc"`
			ID      int    `json:"id"`
			Method  string `json:"method"`
			Params  struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		if req.Params.Name == "get_architecture_overview" {
			port4000 := 4000
			port3000 := 3000
			overview := idxclient.ArchitectureOverviewDTO{
				ProjectID:   "contract-eco",
				ProjectName: "Contract Ecosystem",
				Description: "Integration test ecosystem",
				TotalRepos:  2,
				Repos: []idxclient.RepoDTO{
					{
						Name:       "backend-api",
						LocalPath:  "/path/to/backend",
						TechStack:  []string{"Node.js", "Express"},
						EntryPoint: "server.js",
						Port:       &port4000,
					},
					{
						Name:       "web-client",
						LocalPath:  "/path/to/frontend",
						TechStack:  []string{"React", "TypeScript"},
						EntryPoint: "src/index.tsx",
						Port:       &port3000,
					},
				},
				Relationships: []idxclient.RelationDTO{
					{
						Source:      "web-client",
						Target:      "backend-api",
						Type:        "api_call",
						Description: "Frontend calls backend endpoints",
					},
				},
			}
			text, _ := json.Marshal(overview)
			resp := map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": string(text),
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	})

	httpServer := &http.Server{Handler: mux}
	go func() {
		_ = httpServer.Serve(listener)
	}()
	defer func() { _ = httpServer.Shutdown(ctx) }()

	time.Sleep(50 * time.Millisecond)

	// 2. Test idxclient from cb-ask calling cb-indexer
	client := idxclient.NewIndexerClient("http://127.0.0.1:"+string(rune('0'+serverPort%10)), "")
	client.Endpoint = "http://127.0.0.1:" + listener.Addr().String()[len("127.0.0.1:"):] + "/mcp"

	overview, err := client.GetArchitectureOverview(ctx, "contract-eco")
	if err != nil {
		t.Fatalf("client.GetArchitectureOverview failed: %v", err)
	}

	if overview.ProjectID != "contract-eco" || len(overview.Repos) != 2 {
		t.Fatalf("Unexpected overview DTO: %+v", overview)
	}

	// 3. Test trace_cross_service_flow workflow composite
	flow, err := workflows.TraceCrossServiceFlow(ctx, client, nil, "trace payment from client to api", "web-client", "backend-api", "contract-eco")
	if err != nil {
		t.Fatalf("TraceCrossServiceFlow failed: %v", err)
	}

	if len(flow.Hops) != 1 || flow.Hops[0].Source != "web-client" || flow.Hops[0].Target != "backend-api" {
		t.Fatalf("Unexpected flow report: %+v", flow)
	}

	t.Logf("Contract test successfully verified (Hops: %d)", len(flow.Hops))
}
