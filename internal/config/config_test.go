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
