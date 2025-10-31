package system

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	ChaoticMirrorlistURL = "https://gitlab.com/chaotic-aur/pkgbuilds/-/raw/main/chaotic-mirrorlist/mirrorlist"
	HTTPTimeout          = 10 * time.Second
)

// MirrorlistEntry represents a parsed mirror entry
type MirrorlistEntry struct {
	URL     string
	Enabled bool
}

// FetchChaoticMirrorlist fetches the official Chaotic-AUR mirrorlist from GitLab
// Returns a list of mirrors with their enabled/disabled status
func FetchChaoticMirrorlist() ([]MirrorlistEntry, error) {
	client := &http.Client{Timeout: HTTPTimeout}

	// Fetch mirrorlist from GitLab
	resp, err := client.Get(ChaoticMirrorlistURL)
	if err != nil {
		return nil, fmt.Errorf("network request failed to %s (timeout: %v): %w", ChaoticMirrorlistURL, HTTPTimeout, err)
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch mirrorlist from %s: HTTP %d %s", ChaoticMirrorlistURL, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Parse the mirrorlist content
	mirrors, err := ParseMirrorlist(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mirrorlist from %s: %w", ChaoticMirrorlistURL, err)
	}

	return mirrors, nil
}

// ParseMirrorlist parses a pacman mirrorlist format
// Handles lines like:
//
//	Server = https://mirror.example.com/$repo/$arch
//	#Server = https://disabled.mirror.com/$repo/$arch
func ParseMirrorlist(r io.Reader) ([]MirrorlistEntry, error) {
	var mirrors []MirrorlistEntry
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and non-Server lines
		if line == "" || (!strings.HasPrefix(line, "Server") && !strings.HasPrefix(line, "#Server")) {
			continue
		}

		enabled := !strings.HasPrefix(line, "#")

		// Remove comment prefix if present
		if !enabled {
			line = strings.TrimPrefix(line, "#")
			line = strings.TrimSpace(line)
		}

		// Parse "Server = URL" format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		url := strings.TrimSpace(parts[1])
		if url == "" {
			continue
		}

		mirrors = append(mirrors, MirrorlistEntry{
			URL:     url,
			Enabled: enabled,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing mirrorlist: %w", err)
	}

	if len(mirrors) == 0 {
		return nil, fmt.Errorf("no mirrors found in mirrorlist")
	}

	return mirrors, nil
}

// BuildPackageURLs converts mirror templates to actual package URLs
// Replaces $repo and $arch with actual values, appends package filename
func BuildPackageURLs(mirrors []MirrorlistEntry, packageName string, preferEnabled bool) []string {
	const (
		repo = "chaotic-aur"
		arch = "x86_64"
	)

	var urls []string

	// First pass: enabled mirrors (if requested)
	if preferEnabled {
		for _, mirror := range mirrors {
			if !mirror.Enabled {
				continue
			}
			url := strings.ReplaceAll(mirror.URL, "$repo", repo)
			url = strings.ReplaceAll(url, "$arch", arch)
			url = strings.TrimSuffix(url, "/") + "/" + packageName
			urls = append(urls, url)
		}
	}

	// Second pass: all mirrors (fallback)
	for _, mirror := range mirrors {
		if preferEnabled && mirror.Enabled {
			continue // Already added in first pass
		}
		url := strings.ReplaceAll(mirror.URL, "$repo", repo)
		url = strings.ReplaceAll(url, "$arch", arch)
		url = strings.TrimSuffix(url, "/") + "/" + packageName
		urls = append(urls, url)
	}

	return urls
}
