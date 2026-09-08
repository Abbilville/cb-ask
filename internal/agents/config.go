package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// MergeMcpConfig safely injects or updates MCP server configurations in a target JSON file.
func MergeMcpConfig(filePath string, servers map[string]any) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	parent := filepath.Dir(absPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return "", err
	}

	data := map[string]any{
		"mcpServers": make(map[string]any),
	}

	if content, err := os.ReadFile(absPath); err == nil && len(content) > 0 {
		var existing map[string]any
		if err := json.Unmarshal(content, &existing); err == nil {
			data = existing
			if _, ok := data["mcpServers"]; !ok {
				data["mcpServers"] = make(map[string]any)
			}
		}
	}

	mcpServers, ok := data["mcpServers"].(map[string]any)
	if !ok {
		mcpServers = make(map[string]any)
		data["mcpServers"] = mcpServers
	}

	// Remove legacy server keys if present
	delete(mcpServers, "oss-ask")
	delete(mcpServers, "oss-indexer")
	delete(mcpServers, "oss-mcp")
	delete(mcpServers, "oss-query")
	for name, cfg := range servers {
		mcpServers[name] = cfg
	}

	outBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(absPath, outBytes, 0644); err != nil {
		return "", err
	}

	return absPath, nil
}

func cleanEmptyParentsGo(dir string) {
	current, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	cwd, _ := os.Getwd()
	cwd, _ = filepath.Abs(cwd)
	home, _ := os.UserHomeDir()
	home, _ = filepath.Abs(home)
	for range 3 {
		if current == "" || current == "." || current == cwd || current == home || filepath.Dir(current) == current {
			break
		}
		entries, err := os.ReadDir(current)
		if err != nil || len(entries) > 0 {
			break
		}
		_ = os.Remove(current)
		current = filepath.Dir(current)
	}
}

// UnmergeMcpConfig safely removes named servers from an MCP JSON configuration file.
func UnmergeMcpConfig(filePath string, serverNames ...string) (string, bool, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return absPath, false, nil
	}

	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		return absPath, false, err
	}

	mcpServers, ok := data["mcpServers"].(map[string]any)
	if !ok {
		return absPath, false, nil
	}

	removedAny := false
	for _, name := range serverNames {
		if _, exists := mcpServers[name]; exists {
			delete(mcpServers, name)
			removedAny = true
		}
	}

	// If mcpServers is empty and was the only key, delete file and clean parent directory
	if len(data) <= 1 && len(mcpServers) == 0 {
		_ = os.Remove(absPath)
		cleanEmptyParentsGo(filepath.Dir(absPath))
		return absPath, true, nil
	}

	if removedAny {
		outBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return absPath, false, err
		}
		if err := os.WriteFile(absPath, outBytes, 0644); err != nil {
			return absPath, false, err
		}
	}
	return absPath, removedAny, nil
}

// WriteGuidelineFile writes or appends markdown instructions without duplicate header sections.
func WriteGuidelineFile(filePath, content, headerIdentifier string) (string, string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	parent := filepath.Dir(absPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return absPath, "", err
	}

	if existing, err := os.ReadFile(absPath); err == nil {
		text := string(existing)
		if headerIdentifier != "" && strings.Contains(text, headerIdentifier) {
			return absPath, "Already exists", nil
		}
		newText := strings.TrimSpace(text) + "\n\n" + strings.TrimSpace(content) + "\n"
		if err := os.WriteFile(absPath, []byte(newText), 0644); err != nil {
			return absPath, "", err
		}
		return absPath, "Appended", nil
	}

	if err := os.WriteFile(absPath, []byte(strings.TrimSpace(content)+"\n"), 0644); err != nil {
		return absPath, "", err
	}
	return absPath, "Created", nil
}

// RemoveGuidelineSection removes a markdown section matching headerIdentifier.
func RemoveGuidelineSection(filePath, headerIdentifier string) (string, bool, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return absPath, false, nil
	}

	text := string(content)
	if !strings.Contains(text, headerIdentifier) {
		return absPath, false, nil
	}

	lines := strings.Split(text, "\n")
	var filtered []string
	inSection := false

	for _, line := range lines {
		if strings.Contains(line, headerIdentifier) {
			inSection = true
			continue
		}
		if inSection && (strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ")) {
			inSection = false
		}
		if !inSection {
			filtered = append(filtered, line)
		}
	}

	newText := strings.TrimSpace(strings.Join(filtered, "\n"))
	if newText == "" {
		_ = os.Remove(absPath)
		return absPath, true, nil
	}

	if err := os.WriteFile(absPath, []byte(newText+"\n"), 0644); err != nil {
		return absPath, false, err
	}
	return absPath, true, nil
}
