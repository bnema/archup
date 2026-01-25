package disk

import (
	"testing"
)

// TestNewBtrfsSubvolume tests BtrfsSubvolume creation
func TestNewBtrfsSubvolume(t *testing.T) {
	tests := []struct {
		name       string
		svName     string
		mountPoint string
		shouldErr  bool
		errMsg     string
	}{
		{"valid root subvolume", "@", "/", false, ""},
		{"valid home subvolume", "@home", "/home", false, ""},
		{"valid snapshots subvolume", "@snapshots", "/.snapshots", false, ""},
		{"valid cache subvolume", "@cache", "/var/cache", false, ""},
		{"valid log subvolume", "@log", "/var/log", false, ""},
		{"valid with underscore", "@my_data", "/data", false, ""},
		{"valid with hyphen", "@my-data", "/data", false, ""},
		{"valid with numbers", "@data123", "/data", false, ""},
		{"empty name", "", "/", true, "name cannot be empty"},
		{"no @ prefix", "home", "/home", true, "must start with @"},
		{"space in name", "@ home", "/home", true, "invalid character"},
		{"invalid char in name", "@home!", "/home", true, "invalid character"},
		{"mount point no slash", "@home", "home", true, "must start with /"},
		{"empty mount point", "@tmp", "", false, ""},
		{"name too long", "@" + string(make([]byte, 64)), "/", true, "too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv, err := NewBtrfsSubvolume(tt.svName, tt.mountPoint)

			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}

			if tt.shouldErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errMsg)
				}
			}

			if !tt.shouldErr && sv == nil {
				t.Error("expected non-nil subvolume, got nil")
			}

			if !tt.shouldErr && sv != nil {
				if sv.Name() != tt.svName {
					t.Errorf("got name %q, want %q", sv.Name(), tt.svName)
				}
				if sv.MountPoint() != tt.mountPoint {
					t.Errorf("got mount point %q, want %q", sv.MountPoint(), tt.mountPoint)
				}
			}
		})
	}
}

// TestBtrfsSubvolumeEquals tests subvolume equality
func TestBtrfsSubvolumeEquals(t *testing.T) {
	sv1, _ := NewBtrfsSubvolume("@", "/")
	sv2, _ := NewBtrfsSubvolume("@", "/")
	sv3, _ := NewBtrfsSubvolume("@home", "/home")

	if !sv1.Equals(sv2) {
		t.Error("identical subvolumes should be equal")
	}

	if sv1.Equals(sv3) {
		t.Error("different subvolumes should not be equal")
	}

	if sv1.Equals(nil) {
		t.Error("subvolume should not equal nil")
	}
}

// TestValidateSubvolumeName tests subvolume name validation
func TestValidateSubvolumeName(t *testing.T) {
	tests := []struct {
		name      string
		svName    string
		shouldErr bool
	}{
		{"valid @", "@", false},
		{"valid @home", "@home", false},
		{"valid @snapshots", "@snapshots", false},
		{"valid @cache", "@cache", false},
		{"valid @log", "@log", false},
		{"valid with underscore", "@my_data", false},
		{"valid with hyphen", "@my-data", false},
		{"valid with numbers", "@data123", false},
		{"empty", "", true},
		{"no @ prefix", "home", true},
		{"space in name", "@ home", true},
		{"invalid char !", "@home!", true},
		{"invalid char #", "@home#test", true},
		{"too long", "@" + string(make([]byte, 64)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubvolumeName(tt.svName)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

// TestNewBtrfsLayout tests BtrfsLayout creation
func TestNewBtrfsLayout(t *testing.T) {
	layout := NewBtrfsLayout()
	if layout == nil {
		t.Fatal("expected non-nil layout, got nil")
	}

	if layout.SubvolumeCount() != 0 {
		t.Errorf("new layout should have 0 subvolumes, got %d", layout.SubvolumeCount())
	}
}

// TestNewStandardBtrfsLayout tests standard layout creation
func TestNewStandardBtrfsLayout(t *testing.T) {
	layout, err := NewStandardBtrfsLayout()
	if err != nil {
		t.Fatalf("failed to create standard layout: %v", err)
	}

	if layout.SubvolumeCount() != 2 {
		t.Errorf("standard layout should have 2 subvolumes, got %d", layout.SubvolumeCount())
	}

	// Check for @ subvolume
	root := layout.GetRootSubvolume()
	if root == nil {
		t.Error("standard layout should have root (@) subvolume")
	} else {
		if root.Name() != "@" {
			t.Errorf("root subvolume should be named @, got %q", root.Name())
		}
		if root.MountPoint() != "/" {
			t.Errorf("root subvolume should mount at /, got %q", root.MountPoint())
		}
	}

	// Check for @home subvolume
	home := layout.FindSubvolumeByName("@home")
	if home == nil {
		t.Error("standard layout should have @home subvolume")
	} else {
		if home.MountPoint() != "/home" {
			t.Errorf("@home should mount at /home, got %q", home.MountPoint())
		}
	}
}

// TestBtrfsLayoutAddSubvolume tests adding subvolumes to layout
func TestBtrfsLayoutAddSubvolume(t *testing.T) {
	layout := NewBtrfsLayout()

	// Add root subvolume
	root, _ := NewBtrfsSubvolume("@", "/")
	err := layout.AddSubvolume(root)
	if err != nil {
		t.Fatalf("failed to add root subvolume: %v", err)
	}

	if layout.SubvolumeCount() != 1 {
		t.Errorf("expected 1 subvolume, got %d", layout.SubvolumeCount())
	}

	// Try adding duplicate name
	dup, _ := NewBtrfsSubvolume("@", "/other")
	err = layout.AddSubvolume(dup)
	if err == nil {
		t.Error("should not allow duplicate subvolume names")
	}

	// Try adding duplicate mount point
	other, _ := NewBtrfsSubvolume("@other", "/")
	err = layout.AddSubvolume(other)
	if err == nil {
		t.Error("should not allow duplicate mount points")
	}

	// Add valid @home
	home, _ := NewBtrfsSubvolume("@home", "/home")
	err = layout.AddSubvolume(home)
	if err != nil {
		t.Fatalf("failed to add @home subvolume: %v", err)
	}

	if layout.SubvolumeCount() != 2 {
		t.Errorf("expected 2 subvolumes, got %d", layout.SubvolumeCount())
	}

	// Try adding nil subvolume
	err = layout.AddSubvolume(nil)
	if err == nil {
		t.Error("should not allow nil subvolume")
	}
}

// TestBtrfsLayoutFindSubvolume tests finding subvolumes
func TestBtrfsLayoutFindSubvolume(t *testing.T) {
	layout, _ := NewStandardBtrfsLayout()

	// Find by name
	sv := layout.FindSubvolumeByName("@")
	if sv == nil {
		t.Error("should find @ subvolume")
	}

	sv = layout.FindSubvolumeByName("@home")
	if sv == nil {
		t.Error("should find @home subvolume")
	}

	sv = layout.FindSubvolumeByName("@nonexistent")
	if sv != nil {
		t.Error("should not find nonexistent subvolume")
	}

	// Find by mount point
	sv = layout.FindSubvolumeByMountPoint("/")
	if sv == nil {
		t.Error("should find root subvolume by mount point")
	}

	sv = layout.FindSubvolumeByMountPoint("/home")
	if sv == nil {
		t.Error("should find @home by mount point")
	}

	sv = layout.FindSubvolumeByMountPoint("/nonexistent")
	if sv != nil {
		t.Error("should not find nonexistent mount point")
	}
}

// TestBtrfsLayoutHasRootSubvolume tests root subvolume detection
func TestBtrfsLayoutHasRootSubvolume(t *testing.T) {
	layout := NewBtrfsLayout()

	if layout.HasRootSubvolume() {
		t.Error("empty layout should not have root subvolume")
	}

	root, _ := NewBtrfsSubvolume("@", "/")
	if err := layout.AddSubvolume(root); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !layout.HasRootSubvolume() {
		t.Error("layout with @ subvolume should have root subvolume")
	}
}

// TestBtrfsLayoutValidateLayout tests layout validation
func TestBtrfsLayoutValidateLayout(t *testing.T) {
	// Empty layout should fail
	layout := NewBtrfsLayout()
	err := layout.ValidateLayout()
	if err == nil {
		t.Error("empty layout should fail validation")
	}

	// Layout without root should fail
	home, _ := NewBtrfsSubvolume("@home", "/home")
	if err := layout.AddSubvolume(home); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err = layout.ValidateLayout()
	if err == nil {
		t.Error("layout without root subvolume should fail validation")
	}

	// Layout with wrong root name should fail
	wrongRoot, _ := NewBtrfsSubvolume("@root", "/")
	layout2 := NewBtrfsLayout()
	if err := layout2.AddSubvolume(wrongRoot); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err = layout2.ValidateLayout()
	if err == nil {
		t.Error("layout with wrong root name should fail validation")
	}

	// Standard layout should pass
	standardLayout, _ := NewStandardBtrfsLayout()
	err = standardLayout.ValidateLayout()
	if err != nil {
		t.Errorf("standard layout should pass validation: %v", err)
	}
}

// TestBtrfsLayoutSubvolumes tests getting subvolumes list
func TestBtrfsLayoutSubvolumes(t *testing.T) {
	layout, _ := NewStandardBtrfsLayout()
	subvolumes := layout.Subvolumes()

	if len(subvolumes) != 2 {
		t.Errorf("expected 2 subvolumes, got %d", len(subvolumes))
	}

	// Verify immutability - modifying returned slice shouldn't affect layout
	subvolumes[0] = nil
	if layout.SubvolumeCount() != 2 {
		t.Error("modifying returned subvolumes should not affect layout")
	}
}

// TestBtrfsSubvolumeString tests String method
func TestBtrfsSubvolumeString(t *testing.T) {
	sv1, _ := NewBtrfsSubvolume("@", "/")
	str := sv1.String()
	if !contains(str, "@") || !contains(str, "/") {
		t.Errorf("String() output should contain name and mount point: %s", str)
	}

	sv2, _ := NewBtrfsSubvolume("@temp", "")
	str = sv2.String()
	if !contains(str, "@temp") {
		t.Errorf("String() output should contain name: %s", str)
	}
}

// TestBtrfsLayoutString tests String method
func TestBtrfsLayoutString(t *testing.T) {
	layout, _ := NewStandardBtrfsLayout()
	str := layout.String()
	if !contains(str, "BtrfsLayout") {
		t.Errorf("String() output should contain BtrfsLayout: %s", str)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
