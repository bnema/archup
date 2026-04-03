package config

import "testing"

func TestNewConfig_UsesReleaseVersionAsBootstrapRef(t *testing.T) {
	cfg := NewConfig("v1.2.3")

	if cfg.RawURL != "https://raw.githubusercontent.com/bnema/archup/v1.2.3" {
		t.Fatalf("expected release raw URL, got %q", cfg.RawURL)
	}

	if cfg.Branch() != "v1.2.3" {
		t.Fatalf("expected release ref, got %q", cfg.Branch())
	}
}

func TestNewConfig_UsesDevRefForDevelopmentBuilds(t *testing.T) {
	cfg := NewConfig("v1.2.3-dev")

	if cfg.RawURL != "https://raw.githubusercontent.com/bnema/archup/dev" {
		t.Fatalf("expected dev raw URL, got %q", cfg.RawURL)
	}

	if cfg.Branch() != "dev" {
		t.Fatalf("expected dev ref, got %q", cfg.Branch())
	}
}

func TestNewConfig_GoReleaserVersionWithoutVPrefix(t *testing.T) {
	// GoReleaser strips the 'v' prefix: version={{.Version}} yields "0.5.1" not "v0.5.1".
	cfg := NewConfig("0.5.1")

	if cfg.RawURL != "https://raw.githubusercontent.com/bnema/archup/v0.5.1" {
		t.Fatalf("expected v-prefixed raw URL, got %q", cfg.RawURL)
	}

	if cfg.Branch() != "v0.5.1" {
		t.Fatalf("expected v-prefixed branch, got %q", cfg.Branch())
	}
}

func TestNewConfig_ENVDevOverridesReleaseVersion(t *testing.T) {
	t.Setenv("ENV", "dev")

	cfg := NewConfig("v1.2.3")

	if cfg.RawURL != "https://raw.githubusercontent.com/bnema/archup/dev" {
		t.Fatalf("expected dev raw URL override, got %q", cfg.RawURL)
	}

	if cfg.Branch() != "dev" {
		t.Fatalf("expected dev ref override, got %q", cfg.Branch())
	}
}
