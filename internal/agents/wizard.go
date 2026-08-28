package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func defaultIndexerURL(customURL string) string {
	url := customURL
	if url == "" {
		url = os.Getenv("OSS_INDEXER_URL")
	}
	if url == "" {
		url = "http://127.0.0.1:8080/mcp"
	}
	url = strings.TrimSuffix(url, "/")
	if !strings.HasSuffix(url, "/mcp") {
		url += "/mcp"
	}
	return url
}

func buildServerConfigs(indexerURL string) map[string]any {
	url := defaultIndexerURL(indexerURL)
	return map[string]any{
		"oss-indexer": map[string]any{
			"type": "http",
			"url":  url,
		},
		"oss-ask": map[string]any{
			"command": "oss-ask",
			"args":    []string{"run"},
		},
		"codebase-memory-mcp": map[string]any{
			"command": "codebase-memory-mcp",
		},
	}
}

// SetupAntigravity configures Google Antigravity.
func SetupAntigravity(targetWorkspace string, isGlobal bool, indexerURL string) []string {
	var logs []string
	home, _ := os.UserHomeDir()

	baseDir := filepath.Join(home, ".gemini", "config")
	if !isGlobal {
		ws := targetWorkspace
		if ws == "" {
			ws, _ = os.Getwd()
		}
		baseDir = filepath.Join(ws, ".agents")
	}

	mcpFile := filepath.Join(baseDir, "mcp_config.json")
	saved, err := MergeMcpConfig(mcpFile, buildServerConfigs(indexerURL))
	if err == nil {
		logs = append(logs, fmt.Sprintf("✓ Configured MCP Servers (oss-indexer, oss-ask, codebase-memory-mcp) in: %s", saved))
	} else {
		logs = append(logs, fmt.Sprintf("! Failed to configure MCP in %s: %v", mcpFile, err))
	}

	return logs
}

// SetupClaude configures Claude Desktop and Claude Code CLI.
func SetupClaude(targetWorkspace string, isGlobal bool, indexerURL string) []string {
	var logs []string
	home, _ := os.UserHomeDir()

	var desktopPath string
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		desktopPath = filepath.Join(appData, "Claude", "claude_desktop_config.json")
	} else if runtime.GOOS == "darwin" {
		desktopPath = filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	} else {
		desktopPath = filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	}

	// For Claude Desktop (stdio only), register oss-ask and codebase-memory-mcp
	desktopConfigs := map[string]any{
		"oss-ask": map[string]any{
			"command": "oss-ask",
			"args":    []string{"run"},
		},
		"codebase-memory-mcp": map[string]any{
			"command": "codebase-memory-mcp",
		},
	}
	saved, err := MergeMcpConfig(desktopPath, desktopConfigs)
	if err == nil {
		logs = append(logs, fmt.Sprintf("✓ Configured Claude Desktop MCP in: %s", saved))
	} else {
		logs = append(logs, fmt.Sprintf("! Could not update Claude Desktop config: %v", err))
	}

	return logs
}

// SetupCursor configures Cursor IDE.
func SetupCursor(targetWorkspace string, isGlobal bool, indexerURL string) []string {
	var logs []string
	ws := targetWorkspace
	if ws == "" {
		ws, _ = os.Getwd()
	}

	cursorMcp := filepath.Join(ws, ".cursor", "mcp.json")
	saved, err := MergeMcpConfig(cursorMcp, buildServerConfigs(indexerURL))
	if err == nil {
		logs = append(logs, fmt.Sprintf("✓ Configured Cursor MCP in: %s", saved))
	} else {
		logs = append(logs, fmt.Sprintf("! Could not update Cursor MCP: %v", err))
	}

	return logs
}

// SetupCodex configures OpenAI Codex CLI.
func SetupCodex(targetWorkspace string, isGlobal bool, indexerURL string) []string {
	var logs []string
	home, _ := os.UserHomeDir()

	codexConfig := filepath.Join(home, ".codex", "config.json")
	if !isGlobal {
		ws := targetWorkspace
		if ws == "" {
			ws, _ = os.Getwd()
		}
		codexConfig = filepath.Join(ws, ".codex", "config.json")
	}

	saved, err := MergeMcpConfig(codexConfig, buildServerConfigs(indexerURL))
	if err == nil {
		logs = append(logs, fmt.Sprintf("✓ Configured Codex MCP in: %s", saved))
	} else {
		logs = append(logs, fmt.Sprintf("! Could not update Codex MCP: %v", err))
	}

	return logs
}

// RunAgentSetup dispatches setup to target agents.
func RunAgentSetup(agentChoice, targetWorkspace string, isGlobal bool, indexerURL string) map[string][]string {
	results := make(map[string][]string)
	choice := strings.ToLower(strings.TrimSpace(agentChoice))

	if choice == "antigravity" || choice == "all" || choice == "1" || choice == "5" {
		results["Google Antigravity"] = SetupAntigravity(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "claude" || choice == "all" || choice == "2" || choice == "5" {
		results["Claude (Desktop & Code)"] = SetupClaude(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "cursor" || choice == "all" || choice == "3" || choice == "5" {
		results["Cursor IDE"] = SetupCursor(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "codex" || choice == "all" || choice == "4" || choice == "5" {
		results["OpenAI Codex"] = SetupCodex(targetWorkspace, isGlobal, indexerURL)
	}

	return results
}

// RunAgentUninstall removes oss-ask, oss-query, oss-indexer, and legacy oss-mcp from configs.
func RunAgentUninstall(agentChoice, targetWorkspace string, isGlobal bool, cleanCache bool) map[string][]string {
	results := make(map[string][]string)
	choice := strings.ToLower(strings.TrimSpace(agentChoice))
	home, _ := os.UserHomeDir()
	ws := targetWorkspace
	if ws == "" {
		ws, _ = os.Getwd()
	}

	serversToRemove := []string{"oss-mcp", "oss-query", "oss-indexer", "oss-ask"}

	if choice == "antigravity" || choice == "all" || choice == "1" || choice == "5" {
		baseDir := filepath.Join(home, ".gemini", "config")
		if !isGlobal {
			baseDir = filepath.Join(ws, ".agents")
		}
		mcpFile := filepath.Join(baseDir, "mcp_config.json")
		_, removed, _ := UnmergeMcpConfig(mcpFile, serversToRemove...)
		results["Google Antigravity"] = []string{fmt.Sprintf("- MCP Server: unmerged (removed: %v) in %s", removed, mcpFile)}
	}

	if choice == "claude" || choice == "all" || choice == "2" || choice == "5" {
		var desktopPath string
		if runtime.GOOS == "windows" {
			appData := os.Getenv("APPDATA")
			if appData == "" {
				appData = filepath.Join(home, "AppData", "Roaming")
			}
			desktopPath = filepath.Join(appData, "Claude", "claude_desktop_config.json")
		} else {
			desktopPath = filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
		}
		_, removed, _ := UnmergeMcpConfig(desktopPath, serversToRemove...)
		results["Claude (Desktop & Code)"] = []string{fmt.Sprintf("- Claude Desktop MCP: unmerged (removed: %v) in %s", removed, desktopPath)}
	}

	if choice == "cursor" || choice == "all" || choice == "3" || choice == "5" {
		cursorMcp := filepath.Join(ws, ".cursor", "mcp.json")
		_, removed, _ := UnmergeMcpConfig(cursorMcp, serversToRemove...)
		results["Cursor IDE"] = []string{fmt.Sprintf("- Cursor MCP: unmerged (removed: %v) in %s", removed, cursorMcp)}
	}

	if choice == "codex" || choice == "all" || choice == "4" || choice == "5" {
		codexConfig := filepath.Join(home, ".codex", "config.json")
		if !isGlobal {
			codexConfig = filepath.Join(ws, ".codex", "config.json")
		}
		_, removed, _ := UnmergeMcpConfig(codexConfig, serversToRemove...)
		results["OpenAI Codex"] = []string{fmt.Sprintf("- Codex MCP: unmerged (removed: %v) in %s", removed, codexConfig)}
	}

	if cleanCache {
		catalogPath := filepath.Join(home, ".config", "oss-mcp", "projects.yaml")
		if err := os.Remove(catalogPath); err == nil {
			results["Global Catalog Purge"] = []string{fmt.Sprintf("✓ Purged %s", catalogPath)}
		}
	}

	return results
}

// RunAgentUpdate re-applies the latest server configurations.
func RunAgentUpdate(agentChoice, targetWorkspace string, isGlobal bool, indexerURL string) map[string][]string {
	setupResults := RunAgentSetup(agentChoice, targetWorkspace, isGlobal, indexerURL)
	for agent, logs := range setupResults {
		setupResults[agent] = append([]string{"✓ Re-synced latest MCP server configurations"}, logs...)
	}
	return setupResults
}
