package updater

import "testing"

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

func TestParseSemVer(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected [3]int
	}{
		{"standard version", "1.2.3", [3]int{1, 2, 3}},
		{"zero version", "0.0.0", [3]int{0, 0, 0}},
		{"dev version", "dev", [3]int{0, 0, 0}},
		{"partial version", "1.2", [3]int{1, 2, 0}},
		{"major only", "5", [3]int{5, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSemVer(tt.version)
			if result != tt.expected {
				t.Errorf("parseSemVer(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestGetAssetName(t *testing.T) {
	name := getAssetName()
	if name != "kuranas-linux.zip" && name != "kuranas-windows.zip" {
		t.Errorf("getAssetName() = %q, want kuranas-linux.zip or kuranas-windows.zip", name)
	}
}
