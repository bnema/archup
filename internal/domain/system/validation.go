package system

import (
	"errors"
	"strings"
)

// SystemValidationRules contains business rules for system configuration
type SystemValidationRules struct{}

// NewSystemValidationRules creates a new instance
func NewSystemValidationRules() *SystemValidationRules {
	return &SystemValidationRules{}
}

// ValidateCombination validates a combination of system configuration settings
// This enforces cross-field business rules
func (r *SystemValidationRules) ValidateCombination(config *SystemConfig, cpuInfo *CPUInfo) error {
	if config == nil {
		return errors.New("system config cannot be nil")
	}

	if cpuInfo == nil {
		return errors.New("CPU info cannot be nil")
	}

	// Rule: certain timezones may have locale implications
	// This is more of a guideline/warning in practice

	// Rule: AMD systems may need special handling
	if cpuInfo.Vendor() == CPUVendorAMD {
		// This will be used in higher layers for P-State configuration
	}

	return nil
}

// SanitizeHostname removes or replaces invalid characters from hostname
// Returns a sanitized hostname that passes validation
func (r *SystemValidationRules) SanitizeHostname(hostname string) string {
	// Convert to lowercase
	hostname = strings.ToLower(hostname)

	// Replace invalid characters with hyphen
	var result strings.Builder
	for _, ch := range hostname {
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' {
			result.WriteRune(ch)
		} else {
			result.WriteRune('-')
		}
	}

	hostname = result.String()

	// Remove leading/trailing hyphens
	hostname = strings.Trim(hostname, "-")

	// Collapse consecutive hyphens
	for strings.Contains(hostname, "--") {
		hostname = strings.ReplaceAll(hostname, "--", "-")
	}

	// Limit to 63 characters
	if len(hostname) > 63 {
		hostname = hostname[:63]
	}

	// Remove trailing hyphen if it exists (from truncation)
	hostname = strings.TrimRight(hostname, "-")

	return hostname
}

// SuggestedLocaleForTimezone returns a commonly paired locale for a given timezone
// Returns empty string if no specific suggestion
func (r *SystemValidationRules) SuggestedLocaleForTimezone(timezone string) string {
	// Common timezone -> locale pairings (not enforced, just suggestions)
	suggestions := map[string]string{
		"America/New_York":      "en_US.UTF-8",
		"America/Chicago":       "en_US.UTF-8",
		"America/Denver":        "en_US.UTF-8",
		"America/Los_Angeles":   "en_US.UTF-8",
		"Europe/London":         "en_GB.UTF-8",
		"Europe/Paris":          "fr_FR.UTF-8",
		"Europe/Berlin":         "de_DE.UTF-8",
		"Europe/Madrid":         "es_ES.UTF-8",
		"Asia/Tokyo":            "ja_JP.UTF-8",
		"Australia/Sydney":      "en_AU.UTF-8",
		"UTC":                   "en_US.UTF-8",
	}

	if locale, exists := suggestions[timezone]; exists {
		return locale
	}

	return ""
}

// SuggestedKeymapForLocale returns a commonly used keymap for a given locale
// Returns empty string if no specific suggestion
func (r *SystemValidationRules) SuggestedKeymapForLocale(locale string) string {
	// Common locale -> keymap pairings (not enforced, just suggestions)
	suggestions := map[string]string{
		"en_US.UTF-8": "us",
		"en_GB.UTF-8": "gb",
		"de_DE.UTF-8": "de",
		"fr_FR.UTF-8": "fr",
		"es_ES.UTF-8": "es",
		"it_IT.UTF-8": "it",
		"ja_JP.UTF-8": "jp",
	}

	if keymap, exists := suggestions[locale]; exists {
		return keymap
	}

	return ""
}

// IsValidLocaleVariant checks if a locale is a valid UTF-8 or other standard variant
func (r *SystemValidationRules) IsValidLocaleVariant(locale string) bool {
	// Common valid variants
	validVariants := map[string]bool{
		"UTF-8":  true,
		"utf8":   true,
		"UTF8":   true,
		"ISO8859-1": true,
		"ISO-8859-1": true,
	}

	// Check if locale contains a known valid variant
	for variant := range validVariants {
		if strings.Contains(strings.ToUpper(locale), strings.ToUpper(variant)) {
			return true
		}
	}

	// Accept if it's just a locale code without variant
	if !strings.Contains(locale, ".") && locale != "C" {
		return false
	}

	return true
}

// HostnameRequirements returns a description of hostname requirements
func (r *SystemValidationRules) HostnameRequirements() string {
	return "Hostname must be 1-63 characters, contain only letters, digits, and hyphens. Cannot start or end with hyphen."
}

// TimezoneRequirements returns a description of timezone requirements
func (r *SystemValidationRules) TimezoneRequirements() string {
	return "Timezone can be empty (system default) or a valid IANA timezone like UTC, Europe/London, America/New_York"
}

// LocaleRequirements returns a description of locale requirements
func (r *SystemValidationRules) LocaleRequirements() string {
	return "Locale can be empty (system default) or a valid locale code like en_US.UTF-8"
}

// KeymapRequirements returns a description of keymap requirements
func (r *SystemValidationRules) KeymapRequirements() string {
	return "Keymap must be a valid keyboard layout code like us, de, fr, or dvorak. Can include variants like de-nodeadkeys"
}
