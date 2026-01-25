package defaults

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed packages/*.packages
var packageLists embed.FS

// ReadPackageList loads a named package list.
func ReadPackageList(name string) ([]string, error) {
	path := fmt.Sprintf("packages/%s.packages", name)
	data, err := packageLists.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read package list %s: %w", name, err)
	}
	return splitPackages(string(data)), nil
}

func splitPackages(contents string) []string {
	packages := []string{}
	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		packages = append(packages, strings.Fields(trimmed)...)
	}
	return packages
}
