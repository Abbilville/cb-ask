package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)


func getIndexerURLEnv() string {
	if u := os.Getenv("CB_INDEXER_URL"); u != "" {
		return u
	}
	return os.Getenv("OSS_INDEXER_URL")
}

func defaultIndexerURL(customURL string) string {
	url := customURL
	if url == "" {
		url = getIndexerURLEnv()
	}
	if url == "" {
		url = "http://127.0.0.1:43770/mcp"
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
		"cb-indexer": map[string]any{
			"type": "http",
			"url":  url,
		},
		"cb-ask": map[string]any{
			"command": "cb-ask",
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
		logs = append(logs, fmt.Sprintf("✓ Configured MCP Servers (cb-indexer, cb-ask, codebase-memory-mcp) in: %s", saved))
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

	// For Claude Desktop (stdio only), register cb-ask and codebase-memory-mcp
	desktopConfigs := map[string]any{
		"cb-ask": map[string]any{
			"command": "cb-ask",
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

// SetupVSCode configures VS Code workspace and Cline/Roo Code.
func SetupVSCode(targetWorkspace string, isGlobal bool, indexerURL string) []string {
	var logs []string
	ws := targetWorkspace
	if ws == "" {
		ws, _ = os.Getwd()
	}
	vscodeMcp := filepath.Join(ws, ".vscode", "mcp.json")
	saved, err := MergeMcpConfig(vscodeMcp, buildServerConfigs(indexerURL))
	if err == nil {
		logs = append(logs, fmt.Sprintf("✓ Configured VS Code MCP in: %s", saved))
	} else {
		logs = append(logs, fmt.Sprintf("! Could not update VS Code MCP: %v", err))
	}
	return logs
}

// SetupOmp configures Oh My Pi (omp) harness.
func SetupOmp(targetWorkspace string, isGlobal bool, indexerURL string) []string {
	var logs []string
	ws := targetWorkspace
	if ws == "" {
		ws, _ = os.Getwd()
	}
	ompMcp := filepath.Join(ws, "omp.json")
	saved, err := MergeMcpConfig(ompMcp, buildServerConfigs(indexerURL))
	if err == nil {
		logs = append(logs, fmt.Sprintf("✓ Configured Oh My Pi MCP in: %s", saved))
	}
	agentMcp := filepath.Join(ws, ".agents", "mcp_config.json")
	saved2, err2 := MergeMcpConfig(agentMcp, buildServerConfigs(indexerURL))
	if err2 == nil {
		logs = append(logs, fmt.Sprintf("✓ Configured OMP .agents MCP in: %s", saved2))
	}
	return logs
}

// SetupCopilotCLI configures GitHub Copilot CLI.
func SetupCopilotCLI(targetWorkspace string, isGlobal bool, indexerURL string) []string {
	var logs []string
	home, _ := os.UserHomeDir()
	copilotMcp := filepath.Join(home, ".config", "github-copilot", "mcp.json")
	saved, err := MergeMcpConfig(copilotMcp, buildServerConfigs(indexerURL))
	if err == nil {
		logs = append(logs, fmt.Sprintf("✓ Configured GitHub Copilot CLI MCP in: %s", saved))
	} else {
		logs = append(logs, fmt.Sprintf("! Could not update Copilot CLI MCP: %v", err))
	}
	return logs
}

// SetupWindsurf configures Windsurf (Codeium).
func SetupWindsurf(targetWorkspace string, isGlobal bool, indexerURL string) []string {
	var logs []string
	home, _ := os.UserHomeDir()
	windsurfMcp := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
	saved, err := MergeMcpConfig(windsurfMcp, buildServerConfigs(indexerURL))
	if err == nil {
		logs = append(logs, fmt.Sprintf("✓ Configured Windsurf MCP in: %s", saved))
	} else {
		logs = append(logs, fmt.Sprintf("! Could not update Windsurf MCP: %v", err))
	}
	return logs
}

// SetupSkillFiles generates Codebase RAG skills into target workspace based on selected agents.
func SetupSkillFiles(targetWorkspace string, agentChoices ...string) []string {
	var logs []string
	ws := targetWorkspace
	if ws == "" {
		ws, _ = os.Getwd()
	}

	choiceSet := make(map[string]bool)
	for _, c := range agentChoices {
		choiceSet[strings.ToLower(strings.TrimSpace(c))] = true
	}
	all := len(choiceSet) == 0 || choiceSet["all"]

	const skillContent = `# Skill: Codebase Knowledge Graph & RAG Navigation

## Intent
Use this skill whenever the user asks questions about:
- Microservice architecture, dependencies, communication routes, or service boundaries.
- Cross-service request flows, API endpoints, or event triggers.
- Code definitions, method signatures, or implementations across repositories.

## Priority & Tool Selection
- **ALWAYS use cb-ask tools FIRST** for cross-service, microservice, or architectural inquiries:
  - cb-ask:trace_cross_service_flow(query="...")
  - cb-ask:get_architecture_overview()
  - cb-ask:query_codebase_symbols(query="...")
  - cb-ask:get_symbol_context(repo_name="...", file_path="...")
- **DO NOT call codebase-memory-mcp:list_projects or trace_path** for cross-service communication or topology queries. codebase-memory-mcp is single-repo and lacks network ports and inter-service routing. cb-ask already merges AST knowledge graphs with macro topology.
- **NEVER start with broad whole-workspace grep or blind file searches.**

## Step-by-Step Execution
1. STEP 1 — Identify Boundaries & Topology:
   - Tool: cb-ask:get_architecture_overview()
   - Discovers which services exist, tech stacks, and listening ports.
2. STEP 2 — Trace Cross-Service Flow:
   - Tool: cb-ask:trace_cross_service_flow(query="<feature or endpoint>")
   - Returns ordered hops, source/target ports, and caller/callee AST code symbols.
3. STEP 3 — Query Remote Symbols & Context:
   - Tool: cb-ask:query_codebase_symbols(query="<symbol_pattern>", repo_name="<optional>")
   - Tool: cb-ask:get_symbol_context(repo_name="...", file_path="...", start_line=N, end_line=M)
4. STEP 4 — Read Only Identified Code:
   - Use the knowledge graph's exact file paths and line ranges to read only the critical slices of code.
5. Structure the final response with:
   - Multi-hop request flow summary (Caller -> Gateway -> Callee -> DB).
   - Exact file locations and function signatures (path/to/file:line).
`

	// Universal .agents/skills/codebase-rag/skill.md
	if all || choiceSet["antigravity"] || choiceSet["antigravity-ide"] || choiceSet["omp"] || choiceSet["vscode"] || choiceSet["cursor"] || choiceSet["gemini"] || choiceSet["codex"] || choiceSet["zed"] {
		skillFile := filepath.Join(ws, ".agents", "skills", "codebase-rag", "skill.md")
		if saved, _, err := WriteGuidelineFile(skillFile, skillContent, "Codebase Knowledge Graph"); err == nil {
			logs = append(logs, fmt.Sprintf("✓ Created Codebase RAG skill in: %s", saved))
		}
	}

	// .cursorrules (Only for Cursor)
	if all || choiceSet["cursor"] {
		cursorRules := filepath.Join(ws, ".cursorrules")
		if saved, _, err := WriteGuidelineFile(cursorRules, skillContent, "Codebase Knowledge Graph"); err == nil {
			logs = append(logs, fmt.Sprintf("✓ Added skill to .cursorrules in: %s", saved))
		}
	}

	// CLAUDE.md (Only for Claude)
	if all || choiceSet["claude"] || choiceSet["claude-code"] || choiceSet["claude-desktop"] {
		claudeMd := filepath.Join(ws, "CLAUDE.md")
		if saved, _, err := WriteGuidelineFile(claudeMd, skillContent, "Codebase Knowledge Graph"); err == nil {
			logs = append(logs, fmt.Sprintf("✓ Added skill to CLAUDE.md in: %s", saved))
		}
	}

	// .windsurfrules (Only for Windsurf)
	if all || choiceSet["windsurf"] {
		windsurfRules := filepath.Join(ws, ".windsurfrules")
		if saved, _, err := WriteGuidelineFile(windsurfRules, skillContent, "Codebase Knowledge Graph"); err == nil {
			logs = append(logs, fmt.Sprintf("✓ Added skill to .windsurfrules in: %s", saved))
		}
	}

	return logs
}
// RunAgentSetup dispatches setup to target agents.
func RunAgentSetup(agentChoice, targetWorkspace string, isGlobal bool, indexerURL string) map[string][]string {
	results := make(map[string][]string)
	choice := strings.ToLower(strings.TrimSpace(agentChoice))

	if choice == "vscode" || choice == "all" {
		results["VS Code (Cline / Roo / Copilot)"] = SetupVSCode(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "cursor" || choice == "all" || choice == "3" {
		results["Cursor IDE"] = SetupCursor(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "claude" || choice == "all" || choice == "2" {
		results["Claude (Desktop & Code)"] = SetupClaude(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "omp" || choice == "all" {
		results["Oh My Pi (omp)"] = SetupOmp(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "antigravity" || choice == "all" || choice == "1" {
		results["Google Antigravity"] = SetupAntigravity(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "copilot" || choice == "copilot-cli" || choice == "all" {
		results["GitHub Copilot CLI"] = SetupCopilotCLI(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "windsurf" || choice == "all" {
		results["Windsurf (Codeium)"] = SetupWindsurf(targetWorkspace, isGlobal, indexerURL)
	}
	if choice == "codex" || choice == "all" || choice == "4" {
		results["OpenAI Codex"] = SetupCodex(targetWorkspace, isGlobal, indexerURL)
	}
	// Always ensure skill files are placed
	results["Codebase RAG Skills"] = SetupSkillFiles(targetWorkspace, choice)
	return results
}

// RunAgentUninstall removes cb-ask and legacy servers from all 21 agent configs and cleans skills.
func RunAgentUninstall(agentChoice, targetWorkspace string, isGlobal bool, cleanCache bool) map[string][]string {
	results := make(map[string][]string)
	home, _ := os.UserHomeDir()
	ws := targetWorkspace
	if ws == "" {
		ws, _ = os.Getwd()
	}

	serversToRemove := []string{"cb-ask", "oss-ask", "cb-indexer", "oss-indexer", "oss-mcp", "oss-query"}

	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(home, "AppData", "Roaming")
	}

	var claudeDesktop string
	if runtime.GOOS == "windows" {
		claudeDesktop = filepath.Join(appData, "Claude", "claude_desktop_config.json")
	} else if runtime.GOOS == "darwin" {
		claudeDesktop = filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	} else {
		claudeDesktop = filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	}

	allTargets := map[string][]string{
		"VS Code (Cline / Roo / Copilot)": {
			filepath.Join(ws, ".vscode", "mcp.json"),
			filepath.Join(appData, "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json"),
			filepath.Join(appData, "Code", "User", "globalStorage", "rooveterinaryinc.roo-cline", "settings", "cline_mcp_settings.json"),
		},
		"Cursor IDE": {
			filepath.Join(ws, ".cursor", "mcp.json"),
		},
		"Claude Code (CLI)": {
			filepath.Join(home, ".claude.json"),
		},
		"Claude Desktop": {
			claudeDesktop,
		},
		"Oh My Pi (omp)": {
			filepath.Join(ws, "omp.json"),
			filepath.Join(ws, ".agents", "mcp_config.json"),
		},
		"GitHub Copilot CLI": {
			filepath.Join(ws, ".github", "mcp.json"),
			filepath.Join(home, ".config", "github-copilot", "mcp.json"),
		},
		"Google Antigravity": {
			filepath.Join(ws, ".agents", "mcp_config.json"),
			filepath.Join(home, ".gemini", "config", "mcp_config.json"),
		},
		"Antigravity IDE": {
			filepath.Join(ws, ".antigravity", "mcp.json"),
		},
		"Windsurf (Codeium)": {
			filepath.Join(home, ".codeium", "windsurf", "mcp_config.json"),
		},
		"Gemini CLI": {
			filepath.Join(home, ".gemini", "mcp.json"),
		},
		"OpenAI Codex CLI": {
			filepath.Join(ws, ".codex", "config.json"),
			filepath.Join(home, ".codex", "config.json"),
		},
		"Visual Studio (IDE)": {
			filepath.Join(ws, ".vs", "mcp.json"),
		},
		"Zed Editor": {
			filepath.Join(home, ".config", "zed", "settings.json"),
		},
		"GitLab Duo CLI": {
			filepath.Join(home, ".gitlab", "duo_mcp.json"),
		},
		"Qwen Code": {
			filepath.Join(ws, ".qwen", "mcp.json"),
		},
		"Kimi Code CLI": {
			filepath.Join(ws, ".kimi", "mcp.json"),
		},
		"Grok Build": {
			filepath.Join(ws, ".grok", "mcp.json"),
		},
		"OpenCode": {
			filepath.Join(ws, ".opencode", "mcp.json"),
		},
		"OpenClaw": {
			filepath.Join(ws, ".openclaw", "mcp.json"),
		},
		"KiloCode": {
			filepath.Join(ws, ".kilocode", "mcp.json"),
		},
		"Devin CLI": {
			filepath.Join(ws, ".devin", "mcp.json"),
		},
	}

	for agentName, paths := range allTargets {
		for _, p := range paths {
			if _, exists := os.Stat(p); exists == nil {
				_, removed, _ := UnmergeMcpConfig(p, serversToRemove...)
				if removed {
					results[agentName] = append(results[agentName], fmt.Sprintf("✓ Removed server from: %s", p))
				}
			}
		}
	}

	// Clean skill files
	var skillLogs []string
	skillFile := filepath.Join(ws, ".agents", "skills", "codebase-rag", "skill.md")
	if _, err := os.Stat(skillFile); err == nil {
		_ = os.Remove(skillFile)
		skillLogs = append(skillLogs, fmt.Sprintf("✓ Deleted: %s", skillFile))
		cleanEmptyParentsGo(filepath.Dir(skillFile))
	}

	copilotInstr := filepath.Join(ws, ".github", "copilot-instructions.md")
	if _, err := os.Stat(copilotInstr); err == nil {
		_ = os.Remove(copilotInstr)
		skillLogs = append(skillLogs, fmt.Sprintf("✓ Deleted: %s", copilotInstr))
		cleanEmptyParentsGo(filepath.Dir(copilotInstr))
	}
	cursorRules := filepath.Join(ws, ".cursorrules")
	if data, err := os.ReadFile(cursorRules); err == nil {
		content := string(data)
		if strings.Contains(content, "Codebase Knowledge Graph") {
			_ = os.Remove(cursorRules)
			skillLogs = append(skillLogs, fmt.Sprintf("✓ Cleaned: %s", cursorRules))
			cleanEmptyParentsGo(filepath.Dir(cursorRules))
		}
	}

	claudeMd := filepath.Join(ws, "CLAUDE.md")
	if data, err := os.ReadFile(claudeMd); err == nil {
		content := string(data)
		if strings.Contains(content, "Codebase Knowledge Graph") {
			_ = os.Remove(claudeMd)
			skillLogs = append(skillLogs, fmt.Sprintf("✓ Cleaned: %s", claudeMd))
			cleanEmptyParentsGo(filepath.Dir(claudeMd))
		}
	}

	windsurfRules := filepath.Join(ws, ".windsurfrules")
	if data, err := os.ReadFile(windsurfRules); err == nil {
		content := string(data)
		if strings.Contains(content, "Codebase Knowledge Graph") {
			_ = os.Remove(windsurfRules)
			skillLogs = append(skillLogs, fmt.Sprintf("✓ Deleted: %s", windsurfRules))
			cleanEmptyParentsGo(filepath.Dir(windsurfRules))
		}
	}
	if len(skillLogs) > 0 {
		results["Skills Cleanup"] = skillLogs
	}

	// Remove credentials file
	cfgFile := filepath.Join(appData, "cb-ask", "config.json")
	if _, err := os.Stat(cfgFile); err == nil {
		_ = os.Remove(cfgFile)
		results["Credentials"] = []string{fmt.Sprintf("✓ Deleted: %s", cfgFile)}
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
