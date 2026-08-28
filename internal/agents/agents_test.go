package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMergeAndUnmergeMcpConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfgPath := filepath.Join(tmpDir, "mcp.json")
	servers := buildServerConfigs("http://127.0.0.1:8080")

	saved, err := MergeMcpConfig(cfgPath, servers)
	if err != nil {
		t.Fatalf("MergeMcpConfig failed: %v", err)
	}

	data, err := os.ReadFile(saved)
	if err != nil {
		t.Fatal(err)
	}

	var parsed struct {
		McpServers map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	if _, ok := parsed.McpServers["oss-indexer"]; !ok {
		t.Fatal("Expected oss-indexer in mcpServers")
	}
	if _, ok := parsed.McpServers["oss-ask"]; !ok {
		t.Fatal("Expected oss-ask in mcpServers")
	}
	if _, ok := parsed.McpServers["codebase-memory-mcp"]; !ok {
		t.Fatal("Expected codebase-memory-mcp in mcpServers")
	}

	// Test Unmerge
	_, removed, err := UnmergeMcpConfig(saved, "oss-indexer", "oss-ask")
	if err != nil || !removed {
		t.Fatalf("Unmerge failed (removed: %v, err: %v)", removed, err)
	}

	dataAfter, _ := os.ReadFile(saved)
	var parsedAfter struct {
		McpServers map[string]any `json:"mcpServers"`
	}
	_ = json.Unmarshal(dataAfter, &parsedAfter)

	if _, ok := parsedAfter.McpServers["oss-indexer"]; ok {
		t.Fatal("oss-indexer should have been removed")
	}
	if _, ok := parsedAfter.McpServers["codebase-memory-mcp"]; !ok {
		t.Fatal("codebase-memory-mcp should still be present")
	}
}
