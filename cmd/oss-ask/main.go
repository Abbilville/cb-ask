package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"oss-ask/internal/agents"
	"oss-ask/internal/cbmclient"
	"oss-ask/internal/idxclient"
	"oss-ask/internal/mcpserver"
)

func printHelp() {
	fmt.Println(`oss-ask — Repository Architecture Query Engine & Composition Layer

Usage:
  oss-ask [command] [options]

Commands:
  run              Launch MCP server on stdio transport (default)
  list             Query architecture topology from oss-indexer
  projects         List registered projects and indexed CBM knowledge graphs
  setup-agent      Auto-configure MCP server in AI agents (Antigravity, Claude, Cursor, Codex)
  update-agent     Re-sync MCP server configurations in AI agents
  uninstall-agent  Remove MCP server configurations from AI agents

Run 'oss-ask [command] -h' for more details on each command.`)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if len(os.Args) < 2 {
		// Default: Run MCP server over stdio
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
	case "run":
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		indexerURL := fs.String("indexer-url", "", "oss-indexer HTTP URL")
		authToken := fs.String("auth-token", "", "Secret token for oss-indexer auth")
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
		indexerURL := fs.String("indexer-url", "", "oss-indexer HTTP URL")
		authToken := fs.String("auth-token", "", "Secret token for oss-indexer auth")
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
		indexerURL := fs.String("indexer-url", "", "oss-indexer HTTP URL")
		authToken := fs.String("auth-token", "", "Secret token for oss-indexer auth")
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

	case "setup-agent", "setup:agent":
		fs := flag.NewFlagSet("setup-agent", flag.ExitOnError)
		agent := fs.String("agent", "all", "Target agent: antigravity, claude, cursor, codex, all")
		workspace := fs.String("workspace", ".", "Target workspace path")
		isGlobal := fs.Bool("global", false, "Configure globally where applicable")
		indexerURL := fs.String("indexer-url", "http://127.0.0.1:8080", "URL to deployed oss-indexer")
		fs.Parse(os.Args[2:])

		wsPath, _ := filepath.Abs(*workspace)
		fmt.Printf("[INFO] Setting up MCP servers for agent: %s (Workspace: %s)\n", *agent, wsPath)
		results := agents.RunAgentSetup(*agent, wsPath, *isGlobal, *indexerURL)
		for name, logs := range results {
			fmt.Printf("\n--- %s ---\n", name)
			for _, log := range logs {
				fmt.Printf("  %s\n", log)
			}
		}
		fmt.Println("\n[INFO] Agent configuration completed.")

	case "update-agent", "update:agent":
		fs := flag.NewFlagSet("update-agent", flag.ExitOnError)
		agent := fs.String("agent", "all", "Target agent: antigravity, claude, cursor, codex, all")
		workspace := fs.String("workspace", ".", "Target workspace path")
		isGlobal := fs.Bool("global", false, "Configure globally where applicable")
		indexerURL := fs.String("indexer-url", "http://127.0.0.1:8080", "URL to deployed oss-indexer")
		fs.Parse(os.Args[2:])

		wsPath, _ := filepath.Abs(*workspace)
		fmt.Printf("[INFO] Updating MCP servers for agent: %s (Workspace: %s)\n", *agent, wsPath)
		results := agents.RunAgentUpdate(*agent, wsPath, *isGlobal, *indexerURL)
		for name, logs := range results {
			fmt.Printf("\n--- %s ---\n", name)
			for _, log := range logs {
				fmt.Printf("  %s\n", log)
			}
		}
		fmt.Println("\n[INFO] Agent update completed.")

	case "uninstall-agent", "uninstall:agent":
		fs := flag.NewFlagSet("uninstall-agent", flag.ExitOnError)
		agent := fs.String("agent", "all", "Target agent: antigravity, claude, cursor, codex, all")
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

	case "help", "-h", "--help":
		printHelp()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s'\n\n", command)
		printHelp()
		os.Exit(1)
	}
}
