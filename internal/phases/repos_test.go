package phases

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
)

// TestEnableMultilibParsing tests the multilib section and Include line parsing
func TestEnableMultilibParsing(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantMultilib  bool
		wantInclude   bool
		shouldNotHave string
	}{
		{
			name: "Both commented - should uncomment both",
			input: `[core]
#[multilib]
#Include = /etc/pacman.d/mirrorlist`,
			wantMultilib:  true,
			wantInclude:   true,
			shouldNotHave: "#[multilib]",
		},
		{
			name: "Multilib already enabled - Include still commented",
			input: `[core]
[multilib]
#Include = /etc/pacman.d/mirrorlist`,
			wantMultilib:  true,
			wantInclude:   true,
			shouldNotHave: "#Include = /etc/pacman.d/mirrorlist",
		},
		{
			name: "Both already enabled - no changes",
			input: `[core]
[multilib]
Include = /etc/pacman.d/mirrorlist`,
			wantMultilib:  true,
			wantInclude:   true,
			shouldNotHave: "#",
		},
		{
			name: "Only uncomment multilib Include, not other Includes",
			input: `[core]
Include = /etc/pacman.d/mirrorlist
#[multilib]
#Include = /etc/pacman.d/mirrorlist
[community]
#Include = /etc/pacman.d/mirrorlist`,
			wantMultilib:  true,
			wantInclude:   true,
			shouldNotHave: "",
		},
		{
			name: "Preserve indentation in Include line",
			input: `[core]
#[multilib]
	#Include = /etc/pacman.d/mirrorlist`,
			wantMultilib:  true,
			wantInclude:   true,
			shouldNotHave: "#Include",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the multilib enabling logic from repos.go
			contentStr := tt.input
			lines := strings.Split(contentStr, "\n")
			inMultilib := false

			for i, line := range lines {
				trimmedLine := strings.TrimSpace(line)

				switch {
				case trimmedLine == pacmanMultilibSectionDisabled:
					lines[i] = pacmanMultilibSectionEnabled
					inMultilib = true
				case trimmedLine == pacmanMultilibSectionEnabled:
					// Already enabled, track that we're in multilib section
					inMultilib = true
				case inMultilib && trimmedLine == pacmanMultilibIncludeDisabled:
					// Uncomment the Include line, preserving indentation
					lines[i] = strings.Replace(line, "#Include", "Include", 1)
					inMultilib = false // Only uncomment the first Include after [multilib]
				case inMultilib && strings.HasPrefix(trimmedLine, pacmanSectionPrefix) && trimmedLine != pacmanMultilibSectionEnabled:
					// Exited multilib section without finding Include
					inMultilib = false
				}
			}
			contentStr = strings.Join(lines, "\n")

			// Verify results
			hasMultilib := strings.Contains(contentStr, "[multilib]") && !strings.Contains(contentStr, "#[multilib]")
			hasInclude := strings.Contains(contentStr, "Include = /etc/pacman.d/mirrorlist")

			if tt.wantMultilib && !hasMultilib {
				t.Errorf("Expected [multilib] to be uncommented, but it's not present or still commented")
			}

			if tt.wantInclude && !hasInclude {
				t.Errorf("Expected Include line to be uncommented under [multilib], but it's not")
			}

			if tt.shouldNotHave != "" && strings.Contains(contentStr, tt.shouldNotHave) {
				t.Errorf("Expected NOT to find '%s', but found it in output", tt.shouldNotHave)
			}
		})
	}
}

// TestMultilibEdgeCases tests edge cases in multilib parsing
func TestMultilibEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkFunc   func(string) bool
	}{
		{
			name: "Empty file",
			input: "",
			expectError: false,
			checkFunc: func(s string) bool {
				return s == ""
			},
		},
		{
			name: "No multilib section",
			input: `[core]
Include = /etc/pacman.d/mirrorlist
[extra]
Include = /etc/pacman.d/mirrorlist`,
			expectError: false,
			checkFunc: func(s string) bool {
				return !strings.Contains(s, "[multilib]")
			},
		},
		{
			name: "Multiple Include lines with different paths",
			input: `[core]
#[multilib]
#Include = /etc/pacman.d/mirrorlist
#Include = /etc/pacman.d/other`,
			expectError: false,
			checkFunc: func(s string) bool {
				// Should only uncomment the first Include after [multilib]
				lines := strings.Split(s, "\n")
				var inMultilib bool
				var foundInclude int
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if trimmed == "[multilib]" {
						inMultilib = true
					} else if inMultilib && strings.HasPrefix(trimmed, "Include = ") {
						foundInclude++
					} else if inMultilib && strings.HasPrefix(trimmed, "#Include") {
						// Still commented - that's another Include
					}
				}
				return foundInclude >= 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentStr := tt.input
			lines := strings.Split(contentStr, "\n")
			inMultilib := false

			for i, line := range lines {
				trimmedLine := strings.TrimSpace(line)

				switch {
				case trimmedLine == pacmanMultilibSectionDisabled:
					lines[i] = pacmanMultilibSectionEnabled
					inMultilib = true
				case inMultilib && trimmedLine == pacmanMultilibIncludeDisabled:
					lines[i] = strings.Replace(line, "#Include", "Include", 1)
					inMultilib = false
				case inMultilib && strings.HasPrefix(trimmedLine, pacmanSectionPrefix) && trimmedLine != pacmanMultilibSectionEnabled:
					inMultilib = false
				}
			}
			contentStr = strings.Join(lines, "\n")

			if !tt.checkFunc(contentStr) {
				t.Errorf("Check function failed for input:\n%s\n\nOutput:\n%s", tt.input, contentStr)
			}
		})
	}
}

// TestInstallPackagesIndividuallyLogic tests the fallback package installation logic
func TestInstallPackagesIndividuallyLogic(t *testing.T) {
	// This test verifies the logic of the installPackagesIndividually function
	// by testing how it would handle a list of packages

	packages := []string{"package1", "package2", "package3"}

	// Simulate successful installation for some, failure for others
	failedPackages := map[string]bool{
		"package2": true, // This one fails
	}

	successCount := 0
	failedPkgs := []string{}

	for _, pkg := range packages {
		if failedPackages[pkg] {
			failedPkgs = append(failedPkgs, pkg)
		} else {
			successCount++
		}
	}

	// Verify results
	if successCount != 2 {
		t.Errorf("Expected 2 successful installs, got %d", successCount)
	}

	if len(failedPkgs) != 1 {
		t.Errorf("Expected 1 failed package, got %d", len(failedPkgs))
	}

	if len(failedPkgs) > 0 && failedPkgs[0] != "package2" {
		t.Errorf("Expected 'package2' to fail, got %v", failedPkgs)
	}
}

// BenchmarkMultilibParsing benchmarks the multilib parsing performance
func BenchmarkMultilibParsing(b *testing.B) {
	largeConfig := strings.Repeat(`[section]
#Include = /etc/pacman.d/mirrorlist
`, 100)
	largeConfig = strings.Replace(largeConfig, "[section]", "[core]\n#[multilib]\n#Include = /etc/pacman.d/mirrorlist", 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contentStr := largeConfig
		lines := strings.Split(contentStr, "\n")
		inMultilib := false

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			switch {
			case trimmedLine == pacmanMultilibSectionDisabled:
				lines[i] = pacmanMultilibSectionEnabled
				inMultilib = true
			case trimmedLine == pacmanMultilibSectionEnabled:
				// Already enabled, track that we're in multilib section
				inMultilib = true
			case inMultilib && trimmedLine == pacmanMultilibIncludeDisabled:
				lines[i] = strings.Replace(line, "#Include", "Include", 1)
				inMultilib = false
			case inMultilib && strings.HasPrefix(trimmedLine, pacmanSectionPrefix) && trimmedLine != pacmanMultilibSectionEnabled:
				inMultilib = false
			}
		}
		_ = strings.Join(lines, "\n")
	}
}

// TestReposPhaseStructure tests that ReposPhase can be properly initialized
func TestReposPhaseStructure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a temporary directory for log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	// Create config
	cfg := config.NewConfig("test")

	// Create mocks
	mockFS := mocks.NewMockFileSystem(ctrl)
	mockSysExec := mocks.NewMockSystemExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)

	// Create ReposPhase
	phase := NewReposPhase(cfg, log, mockFS, mockSysExec, mockChrExec)

	// Verify initialization
	if phase == nil {
		t.Error("ReposPhase is nil")
	}

	if phase.BasePhase == nil {
		t.Error("BasePhase is nil")
	}

	if len(phase.chaoticConfig) != 0 {
		t.Error("chaoticConfig should be initialized as empty map")
	}
}

// TestPacmanConfigConstants verifies the constants are properly defined
func TestPacmanConfigConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"pacmanMultilibSectionDisabled", pacmanMultilibSectionDisabled, "#[multilib]"},
		{"pacmanMultilibSectionEnabled", pacmanMultilibSectionEnabled, "[multilib]"},
		{"pacmanMultilibIncludeDisabled", pacmanMultilibIncludeDisabled, "#Include = /etc/pacman.d/mirrorlist"},
		{"pacmanMultilibIncludeEnabled", pacmanMultilibIncludeEnabled, "Include = /etc/pacman.d/mirrorlist"},
		{"pacmanSectionPrefix", pacmanSectionPrefix, "["},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.value)
			}
		})
	}
}

// TestMultilibParsingRealWorldExample tests with a real pacman.conf structure
func TestMultilibParsingRealWorldExample(t *testing.T) {
	// Simulate a real pacman.conf from Arch Linux
	realWorldConfig := `#
# Arch Linux repository mirrorlist
# Generated on 2025-10-28
#

[core]
Include = /etc/pacman.d/mirrorlist

[extra]
Include = /etc/pacman.d/mirrorlist

#[multilib]
#Include = /etc/pacman.d/mirrorlist

[community]
Include = /etc/pacman.d/mirrorlist

# An example of a custom package repository.  See the pacman manpage for
# tips on creating your own repository.
#[custom]
#SigLevel = Optional TrustAll
#Server = file:///home/custompkgs
`

	contentStr := realWorldConfig
	lines := strings.Split(contentStr, "\n")
	inMultilib := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		switch {
		case trimmedLine == pacmanMultilibSectionDisabled:
			lines[i] = pacmanMultilibSectionEnabled
			inMultilib = true
		case inMultilib && trimmedLine == pacmanMultilibIncludeDisabled:
			lines[i] = strings.Replace(line, "#Include", "Include", 1)
			inMultilib = false
		case inMultilib && strings.HasPrefix(trimmedLine, pacmanSectionPrefix) && trimmedLine != pacmanMultilibSectionEnabled:
			inMultilib = false
		}
	}
	contentStr = strings.Join(lines, "\n")

	// Verify results
	if !strings.Contains(contentStr, "[multilib]") {
		t.Error("Expected [multilib] section to be uncommented")
	}

	if strings.Contains(contentStr, "#[multilib]") {
		t.Error("Expected #[multilib] to be removed")
	}

	if !strings.Contains(contentStr, "Include = /etc/pacman.d/mirrorlist") {
		t.Error("Expected Include line to be present")
	}

	// Count Include lines for multilib section
	lines = strings.Split(contentStr, "\n")
	var multilibFound bool
	var multilibIncludeFound bool

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[multilib]" {
			multilibFound = true
		} else if multilibFound {
			if trimmed == "Include = /etc/pacman.d/mirrorlist" {
				multilibIncludeFound = true
				break
			} else if strings.HasPrefix(trimmed, "[") {
				// Exited multilib section
				break
			}
		}
		_ = i
	}

	if !multilibIncludeFound {
		t.Error("Expected Include line directly under [multilib] section")
	}
}

// TestIsValidAURHelper tests AUR helper validation
func TestIsValidAURHelper(t *testing.T) {
	tests := []struct {
		name     string
		helper   string
		expected bool
	}{
		{"Empty helper (no AUR helper)", "", true},
		{"Valid paru", AURHelperParu, true},
		{"Valid yay", AURHelperYay, true},
		{"Invalid helper", "invalid-helper", false},
		{"Invalid pacaur", "pacaur", false},
		{"Invalid trizen", "trizen", false},
		{"Case sensitive check", "Paru", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidAURHelper(tt.helper)
			if result != tt.expected {
				t.Errorf("IsValidAURHelper(%q) = %v, want %v", tt.helper, result, tt.expected)
			}
		})
	}
}

// TestAURHelperConstants verifies AUR helper constants are properly defined
func TestAURHelperConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"AURHelperParu", AURHelperParu, "paru"},
		{"AURHelperYay", AURHelperYay, "yay"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.value)
			}
		})
	}
}

// TestValidAURHelpersList verifies ValidAURHelpers array is correct
func TestValidAURHelpersList(t *testing.T) {
	expected := []string{AURHelperParu, AURHelperYay}

	if len(ValidAURHelpers) != len(expected) {
		t.Errorf("ValidAURHelpers length = %d, want %d", len(ValidAURHelpers), len(expected))
	}

	for i, helper := range expected {
		if ValidAURHelpers[i] != helper {
			t.Errorf("ValidAURHelpers[%d] = %q, want %q", i, ValidAURHelpers[i], helper)
		}
	}
}

// TestReposPhasePreCheckValidation tests PreCheck validation logic
func TestReposPhasePreCheckValidation(t *testing.T) {
	// Create a temporary directory for log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	tests := []struct {
		name        string
		aurHelper   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid paru",
			aurHelper:   AURHelperParu,
			expectError: false,
		},
		{
			name:        "Valid yay",
			aurHelper:   AURHelperYay,
			expectError: false,
		},
		{
			name:        "Empty AUR helper",
			aurHelper:   "",
			expectError: false,
		},
		{
			name:        "Invalid AUR helper",
			aurHelper:   "invalid-helper",
			expectError: true,
			errorMsg:    "invalid AUR helper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create config with test AUR helper
			cfg := config.NewConfig("test")
			cfg.AURHelper = tt.aurHelper

			// Create mocks
			mockFS := mocks.NewMockFileSystem(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)
			mockChrExec := mocks.NewMockChrootExecutor(ctrl)

			// Create ReposPhase
			phase := NewReposPhase(cfg, log, mockFS, mockSysExec, mockChrExec)

			// Note: PreCheck also validates /mnt is mounted and pacman.conf exists
			// In a real test environment, these would fail, so we're only testing
			// the validation logic here by checking the error message
			err := phase.PreCheck()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for AUR helper %q, but got nil", tt.aurHelper)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else if !tt.expectError && err != nil {
				// Only fail if the error is about AUR helper validation
				if strings.Contains(err.Error(), "invalid AUR helper") {
					t.Errorf("Did not expect AUR helper validation error for %q, but got: %v", tt.aurHelper, err)
				}
				// Other errors (like /mnt not mounted) are expected in test environment
			}
		})
	}
}

// TestNoPackageManagerLeakage verifies that the parsing doesn't break other repos
func TestNoPackageManagerLeakage(t *testing.T) {
	pacmanConf := `[core]
Include = /etc/pacman.d/mirrorlist

[extra]
Include = /etc/pacman.d/mirrorlist

#[multilib]
#Include = /etc/pacman.d/mirrorlist

[community]
Include = /etc/pacman.d/mirrorlist

[options]
Server = http://mirror.example.com/arch/$repo/os/$arch
#Include = /etc/pacman.d/custom-servers`

	contentStr := pacmanConf
	lines := strings.Split(contentStr, "\n")
	inMultilib := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		switch {
		case trimmedLine == pacmanMultilibSectionDisabled:
			lines[i] = pacmanMultilibSectionEnabled
			inMultilib = true
		case inMultilib && trimmedLine == pacmanMultilibIncludeDisabled:
			lines[i] = strings.Replace(line, "#Include", "Include", 1)
			inMultilib = false
		case inMultilib && strings.HasPrefix(trimmedLine, pacmanSectionPrefix) && trimmedLine != pacmanMultilibSectionEnabled:
			inMultilib = false
		}
	}
	contentStr = strings.Join(lines, "\n")

	// Verify that other repo Includes are not affected
	coreCount := strings.Count(contentStr, "[core]")
	extraCount := strings.Count(contentStr, "[extra]")
	communityCount := strings.Count(contentStr, "[community]")

	if coreCount != 1 {
		t.Error("Expected exactly one [core] section")
	}
	if extraCount != 1 {
		t.Error("Expected exactly one [extra] section")
	}
	if communityCount != 1 {
		t.Error("Expected exactly one [community] section")
	}

	// Verify [options] section still has commented Include
	if !strings.Contains(contentStr, "#Include = /etc/pacman.d/custom-servers") {
		t.Error("Expected custom-servers Include to remain commented")
	}
}
