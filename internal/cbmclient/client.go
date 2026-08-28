package cbmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CbmSymbol represents a code symbol returned by codebase-memory-mcp graph search.
type CbmSymbol struct {
	Name          string `json:"name"`
	QualifiedName string `json:"qualified_name"`
	Type          string `json:"type"`
	FilePath      string `json:"file_path"`
	LineNumber    int    `json:"line_number,omitempty"`
	Signature     string `json:"signature,omitempty"`
}

// CbmClient executes queries against codebase-memory-mcp AST graphs.
type CbmClient struct {
	Executable string
}

// NewCbmClient initializes a new codebase-memory-mcp client.
func NewCbmClient() *CbmClient {
	return &CbmClient{
		Executable: findCbmExecutable(),
	}
}

func findCbmExecutable() string {
	name := "codebase-memory-mcp"
	pathDirs := filepath.SplitList(os.Getenv("PATH"))

	if runtime.GOOS == "windows" {
		home, err := os.UserHomeDir()
		if err == nil {
			localApp := os.Getenv("LOCALAPPDATA")
			if localApp == "" {
				localApp = filepath.Join(home, "AppData", "Local")
			}
			appData := os.Getenv("APPDATA")
			if appData == "" {
				appData = filepath.Join(home, "AppData", "Roaming")
			}
			pathDirs = append(pathDirs,
				filepath.Join(localApp, "Programs", "codebase-memory-mcp"),
				filepath.Join(appData, "npm"),
			)
		}
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			pathDirs = append(pathDirs, filepath.Join(home, ".local", "bin"))
		}
		pathDirs = append(pathDirs, "/usr/local/bin", "/opt/homebrew/bin")
	}

	extensions := []string{""}
	if runtime.GOOS == "windows" {
		extensions = []string{".cmd", ".exe", ".bat", ".ps1", ""}
	}

	for _, dir := range pathDirs {
		for _, ext := range extensions {
			fullPath := filepath.Join(dir, name+ext)
			if stat, err := os.Stat(fullPath); err == nil && !stat.IsDir() {
				return fullPath
			}
		}
	}

	if runtime.GOOS == "windows" {
		return "codebase-memory-mcp.cmd"
	}
	return "codebase-memory-mcp"
}

// SearchGraph queries codebase-memory-mcp for matching symbols in a project repo.
func (c *CbmClient) SearchGraph(ctx context.Context, pattern, project string) ([]CbmSymbol, error) {
	args := []string{"cli", "--json", "search_graph", "--name-pattern", pattern}
	if project != "" {
		args = append(args, "--project", project)
	}

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, c.Executable, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cbm search_graph error: %s (%w)", errBuf.String(), err)
	}

	var symbols []CbmSymbol
	lines := strings.Split(outBuf.String(), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "{") {
			var raw struct {
				Symbols []CbmSymbol `json:"symbols"`
				Results []CbmSymbol `json:"results"`
			}
			if err := json.Unmarshal([]byte(trimmed), &raw); err == nil {
				if len(raw.Symbols) > 0 {
					symbols = append(symbols, raw.Symbols...)
				}
				if len(raw.Results) > 0 {
					symbols = append(symbols, raw.Results...)
				}
			}
		}
	}

	return symbols, nil
}

// TracePath traces inbound or outbound call paths for a function in codebase-memory-mcp.
func (c *CbmClient) TracePath(ctx context.Context, functionName, direction, project string) (string, error) {
	if direction == "" {
		direction = "outbound"
	}
	args := []string{"cli", "--json", "trace_path", "--function-name", functionName, "--direction", direction}
	if project != "" {
		args = append(args, "--project", project)
	}

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, c.Executable, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cbm trace_path error: %s (%w)", errBuf.String(), err)
	}
	return strings.TrimSpace(outBuf.String()), nil
}

// GetCodeSnippet retrieves the exact code snippet for a qualified symbol.
func (c *CbmClient) GetCodeSnippet(ctx context.Context, qualifiedName, project string) (string, error) {
	args := []string{"cli", "--json", "get_code_snippet", "--qualified-name", qualifiedName}
	if project != "" {
		args = append(args, "--project", project)
	}

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, c.Executable, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cbm get_code_snippet error: %s (%w)", errBuf.String(), err)
	}
	return strings.TrimSpace(outBuf.String()), nil
}
