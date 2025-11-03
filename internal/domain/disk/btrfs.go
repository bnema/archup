package disk

import (
	"errors"
	"fmt"
	"strings"
)

// BtrfsSubvolume is an immutable value object representing a Btrfs subvolume
type BtrfsSubvolume struct {
	name       string // e.g., "@", "@home", "@snapshots"
	mountPoint string // e.g., "/", "/home", "/.snapshots"
}

// NewBtrfsSubvolume creates a new Btrfs subvolume with validation
func NewBtrfsSubvolume(name, mountPoint string) (*BtrfsSubvolume, error) {
	// Validate name
	if err := ValidateSubvolumeName(name); err != nil {
		return nil, err
	}

	// Validate mount point
	if mountPoint != "" && !strings.HasPrefix(mountPoint, "/") {
		return nil, errors.New("mount point must start with / or be empty")
	}

	return &BtrfsSubvolume{
		name:       name,
		mountPoint: mountPoint,
	}, nil
}

// Name returns the subvolume name
func (s *BtrfsSubvolume) Name() string {
	return s.name
}

// MountPoint returns the mount point
func (s *BtrfsSubvolume) MountPoint() string {
	return s.mountPoint
}

// String returns human-readable representation
func (s *BtrfsSubvolume) String() string {
	if s.mountPoint == "" {
		return fmt.Sprintf("BtrfsSubvolume(name=%s)", s.name)
	}
	return fmt.Sprintf("BtrfsSubvolume(name=%s, mount=%s)", s.name, s.mountPoint)
}

// Equals checks if two subvolumes are equal
func (s *BtrfsSubvolume) Equals(other *BtrfsSubvolume) bool {
	if other == nil {
		return false
	}
	return s.name == other.name && s.mountPoint == other.mountPoint
}

// BtrfsLayout is an entity that manages a collection of Btrfs subvolumes
type BtrfsLayout struct {
	subvolumes []*BtrfsSubvolume
}

// NewBtrfsLayout creates a new empty Btrfs layout
func NewBtrfsLayout() *BtrfsLayout {
	return &BtrfsLayout{
		subvolumes: []*BtrfsSubvolume{},
	}
}

// NewStandardBtrfsLayout creates a standard layout with @ and @home subvolumes
// This matches the archup installer's default layout
func NewStandardBtrfsLayout() (*BtrfsLayout, error) {
	layout := NewBtrfsLayout()

	// Create @ (root) subvolume
	root, err := NewBtrfsSubvolume("@", "/")
	if err != nil {
		return nil, err
	}

	// Create @home subvolume
	home, err := NewBtrfsSubvolume("@home", "/home")
	if err != nil {
		return nil, err
	}

	if err := layout.AddSubvolume(root); err != nil {
		return nil, err
	}

	if err := layout.AddSubvolume(home); err != nil {
		return nil, err
	}

	return layout, nil
}

// AddSubvolume adds a subvolume to the layout with validation
func (l *BtrfsLayout) AddSubvolume(subvolume *BtrfsSubvolume) error {
	if subvolume == nil {
		return errors.New("subvolume cannot be nil")
	}

	// Check for duplicate names
	for _, sv := range l.subvolumes {
		if sv.name == subvolume.name {
			return fmt.Errorf("subvolume %s already exists", subvolume.name)
		}
	}

	// Check for duplicate mount points (except empty mount points)
	if subvolume.mountPoint != "" {
		for _, sv := range l.subvolumes {
			if sv.mountPoint == subvolume.mountPoint {
				return fmt.Errorf("mount point %s is already in use by subvolume %s", subvolume.mountPoint, sv.name)
			}
		}
	}

	l.subvolumes = append(l.subvolumes, subvolume)
	return nil
}

// Subvolumes returns a copy of the subvolumes list
func (l *BtrfsLayout) Subvolumes() []*BtrfsSubvolume {
	subvolumes := make([]*BtrfsSubvolume, len(l.subvolumes))
	copy(subvolumes, l.subvolumes)
	return subvolumes
}

// SubvolumeCount returns the number of subvolumes
func (l *BtrfsLayout) SubvolumeCount() int {
	return len(l.subvolumes)
}

// FindSubvolumeByName finds a subvolume by its name
func (l *BtrfsLayout) FindSubvolumeByName(name string) *BtrfsSubvolume {
	for _, sv := range l.subvolumes {
		if sv.name == name {
			return sv
		}
	}
	return nil
}

// FindSubvolumeByMountPoint finds a subvolume by its mount point
func (l *BtrfsLayout) FindSubvolumeByMountPoint(mountPoint string) *BtrfsSubvolume {
	for _, sv := range l.subvolumes {
		if sv.mountPoint == mountPoint {
			return sv
		}
	}
	return nil
}

// HasRootSubvolume returns true if layout has a root (/) subvolume
func (l *BtrfsLayout) HasRootSubvolume() bool {
	return l.FindSubvolumeByMountPoint("/") != nil
}

// GetRootSubvolume returns the root subvolume if it exists
func (l *BtrfsLayout) GetRootSubvolume() *BtrfsSubvolume {
	return l.FindSubvolumeByMountPoint("/")
}

// ValidateLayout validates the Btrfs layout according to business rules
func (l *BtrfsLayout) ValidateLayout() error {
	// Must have at least one subvolume
	if len(l.subvolumes) == 0 {
		return errors.New("Btrfs layout must have at least one subvolume")
	}

	// Must have a root subvolume
	if !l.HasRootSubvolume() {
		return errors.New("Btrfs layout must have a root (/) subvolume")
	}

	// Root subvolume should be named @ by convention
	rootSv := l.GetRootSubvolume()
	if rootSv.name != "@" {
		return errors.New("root subvolume should be named @")
	}

	return nil
}

// String returns human-readable representation
func (l *BtrfsLayout) String() string {
	return fmt.Sprintf("BtrfsLayout(subvolumes=%d)", len(l.subvolumes))
}

// Validation functions

var (
	// ErrInvalidSubvolumeName is returned when subvolume name validation fails
	ErrInvalidSubvolumeName = errors.New("invalid subvolume name")

	// ErrDuplicateSubvolume is returned when a subvolume already exists
	ErrDuplicateSubvolume = errors.New("duplicate subvolume")
)

// ValidateSubvolumeName validates a Btrfs subvolume name
// Valid names: start with @, alphanumeric plus hyphen/underscore
// Examples: @, @home, @snapshots, @cache, @log
func ValidateSubvolumeName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidSubvolumeName)
	}

	// Must start with @
	if !strings.HasPrefix(name, "@") {
		return fmt.Errorf("%w: name must start with @", ErrInvalidSubvolumeName)
	}

	// Must be at least @ (length 1)
	if len(name) < 1 {
		return fmt.Errorf("%w: name too short", ErrInvalidSubvolumeName)
	}

	// Maximum length (reasonable limit)
	if len(name) > 64 {
		return fmt.Errorf("%w: name too long (max 64 chars)", ErrInvalidSubvolumeName)
	}

	// Check characters after @
	for i, ch := range name[1:] { // Skip the @
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_') {
			return fmt.Errorf("%w: invalid character at position %d: %c", ErrInvalidSubvolumeName, i+1, ch)
		}
	}

	// No spaces allowed
	if strings.Contains(name, " ") {
		return fmt.Errorf("%w: spaces not allowed", ErrInvalidSubvolumeName)
	}

	return nil
}
