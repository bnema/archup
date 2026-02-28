package packages

import "errors"

// AURHelper represents AUR helper tool choice
type AURHelper int

const (
	// AURHelperParu is the paru AUR helper
	AURHelperParu AURHelper = iota

	// AURHelperYay is the yay AUR helper
	AURHelperYay
)

// String returns the AUR helper package name
func (a AURHelper) String() string {
	switch a {
	case AURHelperYay:
		return "yay"
	default:
		return "paru"
	}
}

// Repository is an immutable value object for repository configuration.
// Chaotic-AUR is always enabled.
type Repository struct {
	enableMultilib bool
	aurHelper      AURHelper
}

// NewRepository creates a new Repository value object
func NewRepository(enableMultilib bool, aurHelper AURHelper) (*Repository, error) {
	if aurHelper < AURHelperParu || aurHelper > AURHelperYay {
		return nil, errors.New("invalid AUR helper")
	}

	return &Repository{
		enableMultilib: enableMultilib,
		aurHelper:      aurHelper,
	}, nil
}

// EnableMultilib returns true if multilib is enabled
func (r *Repository) EnableMultilib() bool {
	return r.enableMultilib
}

// AURHelper returns the AUR helper choice
func (r *Repository) AURHelper() AURHelper {
	return r.aurHelper
}

// String returns human-readable representation
func (r *Repository) String() string {
	return "Repository(multilib=" + boolToStr(r.enableMultilib) +
		", chaotic=true" +
		", aur=" + r.aurHelper.String() + ")"
}

// Equals checks if two Repository objects are equal
func (r *Repository) Equals(other *Repository) bool {
	if other == nil {
		return false
	}
	return r.enableMultilib == other.enableMultilib &&
		r.aurHelper == other.aurHelper
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
