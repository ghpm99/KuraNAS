package updater

import (
	"slices"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"a less than b major", "1.0.0", "2.0.0", -1},
		{"a greater than b major", "2.0.0", "1.0.0", 1},
		{"a less than b minor", "1.1.0", "1.2.0", -1},
		{"a greater than b minor", "1.2.0", "1.1.0", 1},
		{"a less than b patch", "1.0.1", "1.0.2", -1},
		{"a greater than b patch", "1.0.2", "1.0.1", 1},
		{"dev treated as 0.0.0", "dev", "1.0.0", -1},
		{"both dev equal", "dev", "dev", 0},
		{"complex versions", "1.10.2", "1.9.3", 1},
		{"two digit minor", "0.12.0", "0.3.0", 1},
		// CalVer4 (4-segment) tags
		{"calver4 equal", "26.1.2.3", "26.1.2.3", 0},
		{"calver4 patch bump", "26.1.2.3", "26.1.2.4", -1},
		{"calver4 minor bump resets patch", "26.1.2.3", "26.1.3.0", -1},
		{"calver4 major bump", "26.1.2.3", "26.2.0.0", -1},
		{"calver4 year bump", "26.5.9.9", "27.0.0.0", -1},
		{"calver4 greater patch", "26.1.2.4", "26.1.2.3", 1},
		// Cross-scheme: any CalVer4 tag is newer than any SemVer tag
		{"semver older than calver4", "1.2.3", "26.0.0.0", -1},
		{"high semver still older than calver4", "99.99.99", "26.0.0.0", -1},
		{"calver4 newer than semver", "26.0.0.0", "1.2.3", 1},
		{"dev older than calver4", "dev", "26.0.0.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected []int
	}{
		{"standard semver", "1.2.3", []int{1, 2, 3}},
		{"calver4", "26.1.2.3", []int{26, 1, 2, 3}},
		{"zero version", "0.0.0", []int{0, 0, 0}},
		{"calver4 zero", "26.0.0.0", []int{26, 0, 0, 0}},
		{"dev version", "dev", []int{0}},
		{"partial version", "1.2", []int{1, 2}},
		{"major only", "5", []int{5}},
		{"trailing non-numeric collapses", "1.2.x", []int{0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersion(tt.version)
			if !slices.Equal(result, tt.expected) {
				t.Errorf("parseVersion(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestGetAssetName(t *testing.T) {
	original := runtimeGOOS
	t.Cleanup(func() { runtimeGOOS = original })

	runtimeGOOS = "linux"
	if name := getAssetName(); name != "kuranas-linux.zip" {
		t.Errorf("getAssetName() linux = %q, want kuranas-linux.zip", name)
	}

	runtimeGOOS = "windows"
	if name := getAssetName(); name != "kuranas-windows.zip" {
		t.Errorf("getAssetName() windows = %q, want kuranas-windows.zip", name)
	}
}
