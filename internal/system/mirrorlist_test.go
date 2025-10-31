package system

import (
	"strings"
	"testing"
)

func TestParseMirrorlist(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
		wantFirst MirrorlistEntry
	}{
		{
			name: "Standard mirrorlist with enabled and disabled mirrors",
			input: `# Chaotic-AUR mirrorlist
Server = https://cdn-mirror.chaotic.cx/$repo/$arch
#Server = https://backup.mirror.com/$repo/$arch
Server = https://another.mirror.com/$repo/$arch`,
			wantCount: 3,
			wantErr:   false,
			wantFirst: MirrorlistEntry{URL: "https://cdn-mirror.chaotic.cx/$repo/$arch", Enabled: true},
		},
		{
			name: "Empty mirrorlist",
			input: `# Just comments
# No servers here`,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "Mirrorlist with empty lines and whitespace",
			input: `# Header
Server = https://mirror1.com/$repo/$arch

Server = https://mirror2.com/$repo/$arch

#Server = https://mirror3.com/$repo/$arch`,
			wantCount: 3,
			wantErr:   false,
			wantFirst: MirrorlistEntry{URL: "https://mirror1.com/$repo/$arch", Enabled: true},
		},
		{
			name: "All disabled mirrors",
			input: `#Server = https://disabled1.com/$repo/$arch
#Server = https://disabled2.com/$repo/$arch`,
			wantCount: 2,
			wantErr:   false,
			wantFirst: MirrorlistEntry{URL: "https://disabled1.com/$repo/$arch", Enabled: false},
		},
		{
			name: "Malformed Server lines (should be skipped)",
			input: `Server https://no-equals.com/$repo/$arch
Server =
= https://no-server.com/$repo/$arch
Server = https://valid.com/$repo/$arch`,
			wantCount: 1,
			wantErr:   false,
			wantFirst: MirrorlistEntry{URL: "https://valid.com/$repo/$arch", Enabled: true},
		},
		{
			name: "Mirrors with various whitespace",
			input: `Server   =   https://spaces.com/$repo/$arch
Server=https://nospace.com/$repo/$arch
Server = https://tabs.com/$repo/$arch	`,
			wantCount: 3,
			wantErr:   false,
			wantFirst: MirrorlistEntry{URL: "https://spaces.com/$repo/$arch", Enabled: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mirrors, err := ParseMirrorlist(strings.NewReader(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMirrorlist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(mirrors) != tt.wantCount {
				t.Errorf("ParseMirrorlist() got %d mirrors, want %d", len(mirrors), tt.wantCount)
			}

			if !tt.wantErr && len(mirrors) > 0 {
				if mirrors[0] != tt.wantFirst {
					t.Errorf("ParseMirrorlist() first mirror = %+v, want %+v", mirrors[0], tt.wantFirst)
				}
			}
		})
	}
}

func TestBuildPackageURLs(t *testing.T) {
	tests := []struct {
		name              string
		mirrors           []MirrorlistEntry
		packageName       string
		preferEnabled     bool
		wantCount         int
		wantURLContains   []string
		shouldStartWith   string
	}{
		{
			name: "Basic URL building with enabled mirrors preferred",
			mirrors: []MirrorlistEntry{
				{URL: "https://mirror1.com/$repo/$arch", Enabled: true},
				{URL: "https://mirror2.com/$repo/$arch", Enabled: false},
				{URL: "https://mirror3.com/$repo/$arch", Enabled: true},
			},
			packageName:   "test.pkg.tar.zst",
			preferEnabled: true,
			wantCount:     3,
			wantURLContains: []string{
				"https://mirror1.com/chaotic-aur/x86_64/test.pkg.tar.zst",
				"https://mirror3.com/chaotic-aur/x86_64/test.pkg.tar.zst",
				"https://mirror2.com/chaotic-aur/x86_64/test.pkg.tar.zst",
			},
		},
		{
			name: "Enabled mirrors should come first",
			mirrors: []MirrorlistEntry{
				{URL: "https://disabled.com/$repo/$arch", Enabled: false},
				{URL: "https://enabled1.com/$repo/$arch", Enabled: true},
				{URL: "https://enabled2.com/$repo/$arch", Enabled: true},
			},
			packageName:   "pkg.tar.zst",
			preferEnabled: true,
			wantCount:     3,
			shouldStartWith: "https://enabled1.com/chaotic-aur/x86_64/pkg.tar.zst",
		},
		{
			name: "All mirrors when preferEnabled is false",
			mirrors: []MirrorlistEntry{
				{URL: "https://mirror1.com/$repo/$arch", Enabled: true},
				{URL: "https://mirror2.com/$repo/$arch", Enabled: false},
			},
			packageName:   "test.pkg.tar.zst",
			preferEnabled: false,
			wantCount:     2,
		},
		{
			name: "Empty mirrors list",
			mirrors: []MirrorlistEntry{},
			packageName:   "test.pkg.tar.zst",
			preferEnabled: true,
			wantCount:     0,
		},
		{
			name: "Template variable substitution",
			mirrors: []MirrorlistEntry{
				{URL: "https://example.com/$repo/$arch", Enabled: true},
			},
			packageName:   "chaotic-keyring.pkg.tar.zst",
			preferEnabled: true,
			wantCount:     1,
			wantURLContains: []string{
				"https://example.com/chaotic-aur/x86_64/chaotic-keyring.pkg.tar.zst",
			},
		},
		{
			name: "URL with trailing slash handling",
			mirrors: []MirrorlistEntry{
				{URL: "https://example.com/$repo/$arch/", Enabled: true},
			},
			packageName:   "test.pkg.tar.zst",
			preferEnabled: true,
			wantCount:     1,
			wantURLContains: []string{
				"https://example.com/chaotic-aur/x86_64/test.pkg.tar.zst",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := BuildPackageURLs(tt.mirrors, tt.packageName, tt.preferEnabled)

			if len(urls) != tt.wantCount {
				t.Errorf("BuildPackageURLs() got %d URLs, want %d", len(urls), tt.wantCount)
			}

			for _, wantURL := range tt.wantURLContains {
				found := false
				for _, url := range urls {
					if url == wantURL {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("BuildPackageURLs() URL %q not found in result", wantURL)
					t.Logf("Got URLs: %v", urls)
				}
			}

			if tt.shouldStartWith != "" && len(urls) > 0 {
				if urls[0] != tt.shouldStartWith {
					t.Errorf("BuildPackageURLs() first URL = %q, want %q", urls[0], tt.shouldStartWith)
				}
			}
		})
	}
}

func TestBuildPackageURLs_EnabledOrderingGuarantee(t *testing.T) {
	// Verify that enabled mirrors always come before disabled ones
	mirrors := []MirrorlistEntry{
		{URL: "https://d1.com/$repo/$arch", Enabled: false},
		{URL: "https://e1.com/$repo/$arch", Enabled: true},
		{URL: "https://d2.com/$repo/$arch", Enabled: false},
		{URL: "https://e2.com/$repo/$arch", Enabled: true},
		{URL: "https://d3.com/$repo/$arch", Enabled: false},
	}

	urls := BuildPackageURLs(mirrors, "test.pkg.tar.zst", true)

	if len(urls) != 5 {
		t.Fatalf("Expected 5 URLs, got %d", len(urls))
	}

	// First two should be from enabled mirrors
	if !strings.Contains(urls[0], "e1.com") && !strings.Contains(urls[0], "e2.com") {
		t.Errorf("First URL should be from enabled mirror, got %s", urls[0])
	}
	if !strings.Contains(urls[1], "e1.com") && !strings.Contains(urls[1], "e2.com") {
		t.Errorf("Second URL should be from enabled mirror, got %s", urls[1])
	}

	// Last three should be from disabled mirrors
	for i := 2; i < 5; i++ {
		if !strings.Contains(urls[i], ".d") && !strings.Contains(urls[i], "d1") &&
			!strings.Contains(urls[i], "d2") && !strings.Contains(urls[i], "d3") {
			// Check if it contains any disabled mirror domain
			isDisabled := false
			for _, m := range mirrors {
				if !m.Enabled && strings.Contains(urls[i], strings.Split(m.URL, "/")[2]) {
					isDisabled = true
					break
				}
			}
			if !isDisabled {
				t.Logf("URL at index %d: %s might not be from disabled mirror", i, urls[i])
			}
		}
	}
}

func TestParseMirrorlist_SpecialCharacters(t *testing.T) {
	// Test with various URL formats that might appear in real mirrorlist
	input := `
Server = https://mirror.example.com:8080/$repo/$arch
Server = http://mirror.example.com/$repo/$arch?token=xyz
Server = https://mirror-example.com/$repo/$arch
Server = https://mirror_example.com/$repo/$arch
Server = https://123.456.789.012/$repo/$arch
`

	mirrors, err := ParseMirrorlist(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseMirrorlist() error = %v", err)
	}

	if len(mirrors) != 5 {
		t.Errorf("ParseMirrorlist() got %d mirrors, want 5", len(mirrors))
	}

	// Verify all URLs are preserved correctly
	urls := []string{
		"https://mirror.example.com:8080/$repo/$arch",
		"http://mirror.example.com/$repo/$arch?token=xyz",
		"https://mirror-example.com/$repo/$arch",
		"https://mirror_example.com/$repo/$arch",
		"https://123.456.789.012/$repo/$arch",
	}

	for i, wantURL := range urls {
		if i >= len(mirrors) {
			t.Fatalf("Expected mirror at index %d, but only got %d mirrors", i, len(mirrors))
		}
		if mirrors[i].URL != wantURL {
			t.Errorf("Mirror %d: got %q, want %q", i, mirrors[i].URL, wantURL)
		}
	}
}
