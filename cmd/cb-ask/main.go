package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"cb-ask/internal/agents"
	"cb-ask/internal/cbmclient"
	"cb-ask/internal/idxclient"
	"cb-ask/internal/mcpserver"
)

func printHelp() {
	fmt.Println(`cb-ask — Codebase RAG & Knowledge Graph Agent Setup

Usage:
  cb-ask [command] [options]

Commands:
  setup            Launch interactive configuration wizard (default in terminal)
  run              Launch MCP server on stdio transport (used by AI assistants)
  list             Query architecture topology from cb-indexer
  projects         List registered projects and indexed knowledge graphs
  skill            Install Codebase RAG Skills (.cursorrules, CLAUDE.md, .agents/)
  uninstall        Remove cb-ask configurations and skills from all agents
Run 'cb-ask [command] -h' for more details on each command.`)
}

func isTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func tryRunScript(subcmd string) bool {
	candidates := []string{
		"bin/cb-ask.js",
		"cb-ask.js",
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "cb-ask.js"),
			filepath.Join(exeDir, "bin", "cb-ask.js"),
			filepath.Join(exeDir, "..", "cb-ask", "bin", "cb-ask.js"),
			filepath.Join(exeDir, "..", "src", "cb-ask", "bin", "cb-ask.js"),
		)
	}
	candidates = append(candidates, "C:\\Telkom\\mcp\\cb-ask\\bin\\cb-ask.js")

	for _, c := range candidates {
		if stat, err := os.Stat(c); err == nil && !stat.IsDir() {
			var cmd *exec.Cmd
			if subcmd != "" {
				cmd = exec.Command("node", c, subcmd)
			} else {
				cmd = exec.Command("node", c)
			}
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			return true
		}
	}
	return false
}

func runInteractiveSetupGo() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\ncb-ask — Codebase RAG & Knowledge Graph Agent Setup\n───────────────────────────────────────────────────\n\n")

	fmt.Print("?  cb-indexer location [1: Localhost (http://127.0.0.1:43770), 2: Remote server] (default: 1): ")
	locChoice, _ := reader.ReadString('\n')
	locChoice = strings.TrimSpace(locChoice)

	indexerURL := "http://127.0.0.1:43770/mcp"
	if locChoice == "2" {
		fmt.Print("?  Server URL (default: https://cb-indexer.internal.corp:43770): ")
		urlInput, _ := reader.ReadString('\n')
		urlInput = strings.TrimSpace(urlInput)
		if urlInput == "" {
			urlInput = "https://cb-indexer.internal.corp:43770"
		}
		urlInput = strings.TrimSuffix(urlInput, "/")
		if !strings.HasSuffix(urlInput, "/mcp") {
			urlInput += "/mcp"
		}
		indexerURL = urlInput
	}

	fmt.Print("?  Authentication token/password (leave empty if none): ")
	authToken, _ := reader.ReadString('\n')
	authToken = strings.TrimSpace(authToken)

	fmt.Print("?  Select AI coding agents to configure [all, vscode, cursor, claude, omp, antigravity] (default: all): ")
	agentChoice, _ := reader.ReadString('\n')
	agentChoice = strings.TrimSpace(agentChoice)
	if agentChoice == "" {
		agentChoice = "all"
	}

	fmt.Print("?  Install Codebase RAG skill for agents? [Y/n] (default: Y): ")
	skillChoice, _ := reader.ReadString('\n')
	skillChoice = strings.TrimSpace(strings.ToLower(skillChoice))
	installSkill := skillChoice == "" || skillChoice == "y" || skillChoice == "yes"

	fmt.Print("?  Configuration scope [1: Workspace local, 2: Global] (default: 1): ")
	scopeChoice, _ := reader.ReadString('\n')
	scopeChoice = strings.TrimSpace(scopeChoice)
	isGlobal := scopeChoice == "2"

	ws, _ := os.Getwd()
	fmt.Println("\nWriting agent configurations:")
	results := agents.RunAgentSetup(agentChoice, ws, isGlobal, indexerURL)
	writtenCount := 0
	for name, logs := range results {
		if name == "Codebase RAG Skills" {
			continue
		}
		for _, log := range logs {
			fmt.Printf("  %s\n", log)
			writtenCount++
		}
	}

	if installSkill {
		fmt.Println("\nInstalling Codebase RAG skills:")
		skillLogs := agents.SetupSkillFiles(ws, agentChoice)
		for _, log := range skillLogs {
			fmt.Printf("  %s\n", log)
		}
	}

	fmt.Printf("\n✓ Setup complete. Configured %d agent target(s).\n\n", writtenCount)
	fmt.Println("Ask your AI assistant:")
	fmt.Println("  \"Explain our microservice architecture and listening ports.\"")
	fmt.Println("  \"Trace how payment-service communicates with account-service.\"")
	fmt.Println("  \"Find all handlers for user checkout across the repositories.\"")
	fmt.Println()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if len(os.Args) < 2 {
		if isTerminal() {
			if tryRunScript("") {
				return
			}
			runInteractiveSetupGo()
			return
		}
		// Non-interactive / Piped by AI agent: Run MCP server over stdio
		idx := idxclient.NewIndexerClient("", "")
		cbm := cbmclient.NewCbmClient()
		srv := mcpserver.NewServer(idx, cbm)

		if err := mcpserver.ServeStdio(ctx, srv); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	command := os.Args[1]

	switch command {
	case "setup", "init", "wizard", "config":
		if tryRunScript("") {
			return
		}
		runInteractiveSetupGo()
		return

	case "run":
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		indexerURL := fs.String("indexer-url", "", "cb-indexer HTTP URL")
		authToken := fs.String("auth-token", "", "Secret token for cb-indexer auth")
		fs.Parse(os.Args[2:])

		idx := idxclient.NewIndexerClient(*indexerURL, *authToken)
		cbm := cbmclient.NewCbmClient()
		srv := mcpserver.NewServer(idx, cbm)

		if err := mcpserver.ServeStdio(ctx, srv); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}

	case "list":
		fs := flag.NewFlagSet("list", flag.ExitOnError)
		project := fs.String("p", "", "Project ID or registry path")
		regPath := fs.String("r", "", "Path to custom registry.yaml")
		outputJSON := fs.Bool("json", false, "Output raw JSON representation")
		indexerURL := fs.String("indexer-url", "", "cb-indexer HTTP URL")
		authToken := fs.String("auth-token", "", "Secret token for cb-indexer auth")
		fs.Parse(os.Args[2:])

		target := *regPath
		if target == "" {
			target = *project
		}

		idx := idxclient.NewIndexerClient(*indexerURL, *authToken)
		overview, err := idx.GetArchitectureOverview(ctx, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			os.Exit(1)
		}

		if *outputJSON {
			bytes, _ := json.MarshalIndent(overview, "", "  ")
			fmt.Println(string(bytes))
			return
		}

		fmt.Printf("\n=== Project Ecosystem: %s [%s] ===\n", overview.ProjectName, overview.ProjectID)
		if overview.Description != "" {
			fmt.Printf("Description: %s\n", overview.Description)
		}
		if overview.SourcePath != "" {
			fmt.Printf("Manifest:    %s\n", overview.SourcePath)
		}
		fmt.Printf("Repositories (%d):\n", len(overview.Repos))
		for _, r := range overview.Repos {
			stackStr := "Generic"
			if len(r.TechStack) > 0 {
				stackStr = fmt.Sprintf("%v", r.TechStack)
			}
			portStr := ""
			if r.Port != nil {
				portStr = fmt.Sprintf(" [port %d]", *r.Port)
			}
			fmt.Printf("  * %-22s %s%s\n", r.Name, stackStr, portStr)
			if r.Description != "" {
				fmt.Printf("    - %s\n", r.Description)
			}
			if r.LocalPath != "" {
				fmt.Printf("    - Path: %s\n", r.LocalPath)
			}
		}

		if len(overview.Relationships) > 0 {
			fmt.Printf("\nRelationships (%d):\n", len(overview.Relationships))
			for _, rel := range overview.Relationships {
				fmt.Printf("  * %s --[%s]--> %s\n", rel.Source, rel.Type, rel.Target)
				if rel.Description != "" {
					fmt.Printf("    - %s\n", rel.Description)
				}
			}
		}
		fmt.Println()

	case "projects":
		fs := flag.NewFlagSet("projects", flag.ExitOnError)
		outputJSON := fs.Bool("json", false, "Output raw JSON representation")
		indexerURL := fs.String("indexer-url", "", "cb-indexer HTTP URL")
		authToken := fs.String("auth-token", "", "Secret token for cb-indexer auth")
		fs.Parse(os.Args[2:])

		target := ""
		if fs.NArg() > 0 {
			target = fs.Arg(0)
		}

		idx := idxclient.NewIndexerClient(*indexerURL, *authToken)
		projList, err := idx.ListProjects(ctx, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			os.Exit(1)
		}

		if *outputJSON {
			bytes, _ := json.MarshalIndent(projList, "", "  ")
			fmt.Println(string(bytes))
			return
		}

		fmt.Printf("\n=== Registered Projects (%d) ===\n", projList.TotalRegisteredProjects)
		for _, p := range projList.RegisteredProjects {
			fmt.Printf("  * %-16s - %s\n", p.ProjectID, p.Name)
			if p.Description != "" {
				fmt.Printf("    %s\n", p.Description)
			}
			if p.RegistryPath != "" {
				fmt.Printf("    Registry: %s\n", p.RegistryPath)
			}
		}

		fmt.Printf("\n=== Indexed codebase-memory-mcp Knowledge Graphs (%d) ===\n", projList.TotalIndexedGraphs)
		for _, g := range projList.IndexedCodebaseMemoryGraphs {
			name, _ := g["name"].(string)
			fmt.Printf("  * %-25s\n", name)
		}
		fmt.Println()

	case "uninstall", "uninstall-agent", "uninstall:agent", "remove":
		if tryRunScript("uninstall") {
			return
		}
		fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
		agent := fs.String("agent", "all", "Target agent: all, cursor, vscode, claude, etc.")
		workspace := fs.String("workspace", ".", "Target workspace path")
		isGlobal := fs.Bool("global", false, "Uninstall from global configuration")
		cleanCache := fs.Bool("clean-cache", false, "Purge global projects catalog")
		fs.Parse(os.Args[2:])

		wsPath, _ := filepath.Abs(*workspace)
		fmt.Printf("[INFO] Uninstalling MCP servers for agent: %s\n", *agent)
		results := agents.RunAgentUninstall(*agent, wsPath, *isGlobal, *cleanCache)
		for name, logs := range results {
			fmt.Printf("\n--- %s ---\n", name)
			for _, log := range logs {
				fmt.Printf("  %s\n", log)
			}
		}
		fmt.Println("\n[INFO] Agent uninstallation completed.")


	case "skill", "skills":
		fs := flag.NewFlagSet("skill", flag.ExitOnError)
		workspace := fs.String("workspace", ".", "Target workspace path")
		fs.Parse(os.Args[2:])

		wsPath, _ := filepath.Abs(*workspace)
		logs := agents.SetupSkillFiles(wsPath)
		fmt.Printf("\n=== Installed Codebase RAG Skills ===\n")
		for _, log := range logs {
			fmt.Printf("  %s\n", log)
		}
		fmt.Println()
	case "help", "-h", "--help":
		printHelp()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s'\n\n", command)
		printHelp()
		os.Exit(1)
	}
}
