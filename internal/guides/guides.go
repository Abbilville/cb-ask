package guides

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed *.md
var guideFS embed.FS

// GetWorkflowGuide retrieves an embedded workflow markdown guide by name.
func GetWorkflowGuide(guideName string) (string, error) {
	name := strings.TrimSpace(strings.ToLower(guideName))
	name = strings.TrimSuffix(name, ".md")

	// Normalize legacy names
	if name == "oss" {
		name = "cb"
	} else if name == "oss-navigator" {
		name = "cb-navigator"
	}

	fileName := name + ".md"
	data, err := guideFS.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("guide '%s' not found. Available guides: 'cb', 'cb-navigator'", guideName)
	}

	return string(data), nil
}
