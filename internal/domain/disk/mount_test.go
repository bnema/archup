package disk

import (
	"testing"
)

// TestNewMountOptions tests MountOptions creation
func TestNewMountOptions(t *testing.T) {
	tests := []struct {
		name      string
		options   map[string]string
		shouldErr bool
	}{
		{"valid empty", map[string]string{}, false},
		{"valid with values", map[string]string{"compress": "zstd", "noatime": ""}, false},
		{"valid subvol", map[string]string{"subvol": "@"}, false},
		{"nil options", nil, true},
		{"empty key", map[string]string{"": "value"}, true},
		{"key with space", map[string]string{"no time": ""}, true},
		{"key with comma", map[string]string{"no,time": ""}, true},
		{"key with equals", map[string]string{"no=time": ""}, true},
		{"value with space", map[string]string{"compress": "zstd level"}, true},
		{"value with comma", map[string]string{"compress": "zstd,level"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := NewMountOptions(tt.options)

			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}

			if !tt.shouldErr && opts == nil {
				t.Error("expected non-nil options, got nil")
			}

			// Verify immutability - original map should not affect internal state
			if !tt.shouldErr && tt.options != nil {
				tt.options["test"] = "value"
				if opts.Has("test") {
					t.Error("modifying original map should not affect MountOptions")
				}
			}
		})
	}
}

// TestNewBtrfsMountOptions tests standard Btrfs mount options creation
func TestNewBtrfsMountOptions(t *testing.T) {
	tests := []struct {
		name      string
		subvolume string
	}{
		{"without subvolume", ""},
		{"with @ subvolume", "@"},
		{"with @home subvolume", "@home"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := NewBtrfsMountOptions(tt.subvolume)
			if err != nil {
				t.Fatalf("failed to create Btrfs mount options: %v", err)
			}

			if !opts.Has("noatime") {
				t.Error("Btrfs mount options should have noatime")
			}

			if !opts.Has("compress") {
				t.Error("Btrfs mount options should have compress")
			}

			compress, _ := opts.Get("compress")
			if compress != "zstd" {
				t.Errorf("compress should be zstd, got %q", compress)
			}

			if tt.subvolume != "" {
				if !opts.Has("subvol") {
					t.Error("should have subvol option when subvolume is specified")
				}
				subvol, _ := opts.Get("subvol")
				if subvol != tt.subvolume {
					t.Errorf("subvol should be %q, got %q", tt.subvolume, subvol)
				}
			} else {
				if opts.Has("subvol") {
					t.Error("should not have subvol option when subvolume is empty")
				}
			}
		})
	}
}

// TestMountOptionsHas tests Has method
func TestMountOptionsHas(t *testing.T) {
	opts, _ := NewMountOptions(map[string]string{
		"noatime":  "",
		"compress": "zstd",
	})

	if !opts.Has("noatime") {
		t.Error("should have noatime option")
	}

	if !opts.Has("compress") {
		t.Error("should have compress option")
	}

	if opts.Has("nonexistent") {
		t.Error("should not have nonexistent option")
	}
}

// TestMountOptionsGet tests Get method
func TestMountOptionsGet(t *testing.T) {
	opts, _ := NewMountOptions(map[string]string{
		"noatime":  "",
		"compress": "zstd",
	})

	// Get existing option with value
	value, exists := opts.Get("compress")
	if !exists {
		t.Error("compress option should exist")
	}
	if value != "zstd" {
		t.Errorf("compress value should be zstd, got %q", value)
	}

	// Get existing option without value
	value, exists = opts.Get("noatime")
	if !exists {
		t.Error("noatime option should exist")
	}
	if value != "" {
		t.Errorf("noatime value should be empty, got %q", value)
	}

	// Get nonexistent option
	_, exists = opts.Get("nonexistent")
	if exists {
		t.Error("nonexistent option should not exist")
	}
}

// TestMountOptionsToString tests ToString method
func TestMountOptionsToString(t *testing.T) {
	tests := []struct {
		name     string
		options  map[string]string
		expected []string // all possible orderings
	}{
		{
			"empty",
			map[string]string{},
			[]string{""},
		},
		{
			"single flag",
			map[string]string{"noatime": ""},
			[]string{"noatime"},
		},
		{
			"single value",
			map[string]string{"compress": "zstd"},
			[]string{"compress=zstd"},
		},
		{
			"multiple options",
			map[string]string{"noatime": "", "compress": "zstd"},
			[]string{
				"noatime,compress=zstd",
				"compress=zstd,noatime",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, _ := NewMountOptions(tt.options)
			result := opts.ToString()

			// Check if result matches any expected ordering
			found := false
			for _, expected := range tt.expected {
				if result == expected {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("ToString() = %q, want one of %v", result, tt.expected)
			}
		})
	}
}

// TestMountOptionsEquals tests Equals method
func TestMountOptionsEquals(t *testing.T) {
	opts1, _ := NewMountOptions(map[string]string{"noatime": "", "compress": "zstd"})
	opts2, _ := NewMountOptions(map[string]string{"noatime": "", "compress": "zstd"})
	opts3, _ := NewMountOptions(map[string]string{"noatime": ""})
	opts4, _ := NewMountOptions(map[string]string{"noatime": "", "compress": "lzo"})

	if !opts1.Equals(opts2) {
		t.Error("identical options should be equal")
	}

	if opts1.Equals(opts3) {
		t.Error("options with different keys should not be equal")
	}

	if opts1.Equals(opts4) {
		t.Error("options with different values should not be equal")
	}

	if opts1.Equals(nil) {
		t.Error("options should not equal nil")
	}
}

// TestMountOptionsOptions tests Options method (immutability)
func TestMountOptionsOptions(t *testing.T) {
	opts, _ := NewMountOptions(map[string]string{"noatime": "", "compress": "zstd"})
	options := opts.Options()

	// Modify returned map
	options["test"] = "value"

	// Verify internal state is not affected
	if opts.Has("test") {
		t.Error("modifying returned options should not affect MountOptions")
	}
}

// TestParseMountOptions tests ParseMountOptions function
func TestParseMountOptions(t *testing.T) {
	tests := []struct {
		name       string
		optionsStr string
		shouldErr  bool
		expected   map[string]string
	}{
		{"empty", "", false, map[string]string{}},
		{"single flag", "noatime", false, map[string]string{"noatime": ""}},
		{"single value", "compress=zstd", false, map[string]string{"compress": "zstd"}},
		{"multiple options", "noatime,compress=zstd", false, map[string]string{"noatime": "", "compress": "zstd"}},
		{"with subvol", "noatime,compress=zstd,subvol=@", false, map[string]string{"noatime": "", "compress": "zstd", "subvol": "@"}},
		{"with spaces", " noatime , compress=zstd ", false, map[string]string{"noatime": "", "compress": "zstd"}},
		{"invalid format", "compress=", false, map[string]string{"compress": ""}},
		{"extra commas", "noatime,,compress=zstd", false, map[string]string{"noatime": "", "compress": "zstd"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseMountOptions(tt.optionsStr)

			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}

			if !tt.shouldErr && opts != nil {
				for key, expectedValue := range tt.expected {
					value, exists := opts.Get(key)
					if !exists {
						t.Errorf("expected key %q to exist", key)
					}
					if value != expectedValue {
						t.Errorf("key %q: got value %q, want %q", key, value, expectedValue)
					}
				}

				// Verify no extra keys
				gotOptions := opts.Options()
				if len(gotOptions) != len(tt.expected) {
					t.Errorf("got %d options, want %d", len(gotOptions), len(tt.expected))
				}
			}
		})
	}
}

// TestNewMountPoint tests MountPoint creation
func TestNewMountPoint(t *testing.T) {
	validOpts, _ := NewMountOptions(map[string]string{"noatime": ""})

	tests := []struct {
		name      string
		path      string
		device    string
		options   *MountOptions
		shouldErr bool
	}{
		{"valid root", "/", "/dev/sda1", validOpts, false},
		{"valid home", "/home", "/dev/sda2", validOpts, false},
		{"valid boot", "/boot", "/dev/sda1", validOpts, false},
		{"empty path", "", "/dev/sda1", validOpts, true},
		{"path no slash", "home", "/dev/sda1", validOpts, true},
		{"empty device", "/home", "", validOpts, true},
		{"nil options", "/home", "/dev/sda1", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp, err := NewMountPoint(tt.path, tt.device, tt.options)

			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}

			if !tt.shouldErr && mp == nil {
				t.Error("expected non-nil mount point, got nil")
			}

			if !tt.shouldErr && mp != nil {
				if mp.Path() != tt.path {
					t.Errorf("got path %q, want %q", mp.Path(), tt.path)
				}
				if mp.Device() != tt.device {
					t.Errorf("got device %q, want %q", mp.Device(), tt.device)
				}
				if !mp.Options().Equals(tt.options) {
					t.Error("options should match")
				}
			}
		})
	}
}

// TestMountPointIsRoot tests IsRoot method
func TestMountPointIsRoot(t *testing.T) {
	opts, _ := NewMountOptions(map[string]string{"noatime": ""})

	root, _ := NewMountPoint("/", "/dev/sda1", opts)
	if !root.IsRoot() {
		t.Error("/ should be root mount point")
	}

	home, _ := NewMountPoint("/home", "/dev/sda2", opts)
	if home.IsRoot() {
		t.Error("/home should not be root mount point")
	}
}

// TestMountPointEquals tests Equals method
func TestMountPointEquals(t *testing.T) {
	opts1, _ := NewMountOptions(map[string]string{"noatime": ""})
	opts2, _ := NewMountOptions(map[string]string{"noatime": ""})
	opts3, _ := NewMountOptions(map[string]string{"compress": "zstd"})

	mp1, _ := NewMountPoint("/", "/dev/sda1", opts1)
	mp2, _ := NewMountPoint("/", "/dev/sda1", opts2)
	mp3, _ := NewMountPoint("/home", "/dev/sda2", opts1)
	mp4, _ := NewMountPoint("/", "/dev/sda1", opts3)

	if !mp1.Equals(mp2) {
		t.Error("identical mount points should be equal")
	}

	if mp1.Equals(mp3) {
		t.Error("mount points with different paths should not be equal")
	}

	if mp1.Equals(mp4) {
		t.Error("mount points with different options should not be equal")
	}

	if mp1.Equals(nil) {
		t.Error("mount point should not equal nil")
	}
}

// TestValidateMountPoint tests ValidateMountPoint function
func TestValidateMountPoint(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{"valid root", "/", false},
		{"valid home", "/home", false},
		{"valid nested", "/var/log", false},
		{"empty", "", true},
		{"no slash", "home", true},
		{"double slash", "/home//data", true},
		{"dot dot", "/home/../data", true},
		{"too long", "/" + string(make([]byte, 256)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMountPoint(tt.path)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

// TestMountOptionsString tests String method
func TestMountOptionsString(t *testing.T) {
	opts, _ := NewMountOptions(map[string]string{"noatime": ""})
	str := opts.String()
	if !contains(str, "MountOptions") {
		t.Errorf("String() should contain MountOptions: %s", str)
	}
}

// TestMountPointString tests String method
func TestMountPointString(t *testing.T) {
	opts, _ := NewMountOptions(map[string]string{"noatime": ""})
	mp, _ := NewMountPoint("/", "/dev/sda1", opts)
	str := mp.String()
	if !contains(str, "MountPoint") || !contains(str, "/") || !contains(str, "/dev/sda1") {
		t.Errorf("String() should contain path and device: %s", str)
	}
}
