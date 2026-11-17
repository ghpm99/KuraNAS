package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const checksumTestAssetName = "kuranas-linux.zip"

func fourBytesChecksumList(assetName string) string {
	return checksumListFor("1234", assetName)
}

func checksumListFor(content, assetName string) string {
	digest := sha256.Sum256([]byte(content))
	return hex.EncodeToString(digest[:]) + "  " + assetName + "\n"
}

type checksumScenario struct {
	assets       []GitHubAsset
	checksumList string
	packageBody  string
}

type checksumOutcome struct {
	err           error
	applyCalls    int
	updateDirGone bool
}

func runChecksumScenario(t *testing.T, scenario checksumScenario) checksumOutcome {
	t.Helper()
	resetServiceFns()
	t.Cleanup(resetServiceFns)

	getAssetNameFunc = func() string { return checksumTestAssetName }
	fetchLatestReleaseFunc = func() (GitHubRelease, error) {
		return GitHubRelease{TagName: "v2.0.0", Assets: scenario.assets}, nil
	}
	downloadFileFunc = func(url, dest string, maxBytes int64) error {
		if url == "sums-url" {
			return os.WriteFile(dest, []byte(scenario.checksumList), 0644)
		}
		return os.WriteFile(dest, []byte(scenario.packageBody), 0644)
	}
	extractAllFunc = func(zipPath, destDir string) error { return os.MkdirAll(destDir, 0755) }

	outcome := checksumOutcome{}
	applyFullUpdateFunc = func(extractedDir string) error {
		outcome.applyCalls++
		return nil
	}

	outcome.err = NewService().DownloadAndApply()
	_, statErr := os.Stat(filepath.Join(os.TempDir(), "kuranas-update"))
	outcome.updateDirGone = os.IsNotExist(statErr)
	return outcome
}

func packageAndSumsAssets() []GitHubAsset {
	return []GitHubAsset{
		{Name: checksumTestAssetName, Size: 4, BrowserDownloadURL: "pkg-url"},
		{Name: checksumsAssetName, BrowserDownloadURL: "sums-url"},
	}
}

func TestDownloadAndApply_MatchingChecksumApplies(t *testing.T) {
	outcome := runChecksumScenario(t, checksumScenario{
		assets:       packageAndSumsAssets(),
		checksumList: strings.ToUpper(checksumListFor("1234", checksumTestAssetName)[:64]) + "  " + checksumTestAssetName,
		packageBody:  "1234",
	})

	if outcome.err != nil || outcome.applyCalls != 1 || !outcome.updateDirGone {
		t.Fatalf("expected applied and cleaned, got err=%v applyCalls=%d dirGone=%v", outcome.err, outcome.applyCalls, outcome.updateDirGone)
	}
}

func TestDownloadAndApply_RefusalsApplyNothingAndRemoveTempDir(t *testing.T) {
	cases := []struct {
		name     string
		scenario checksumScenario
		expected error
	}{
		{
			name:     "different hash",
			scenario: checksumScenario{assets: packageAndSumsAssets(), checksumList: checksumListFor("abcd", checksumTestAssetName), packageBody: "1234"},
			expected: errChecksumMismatch,
		},
		{
			name:     "release without SHA256SUMS",
			scenario: checksumScenario{assets: packageAndSumsAssets()[:1], packageBody: "1234"},
			expected: errChecksumMissing,
		},
		{
			name:     "SHA256SUMS without package entry",
			scenario: checksumScenario{assets: packageAndSumsAssets(), checksumList: checksumListFor("1234", "other.zip"), packageBody: "1234"},
			expected: errChecksumNoEntry,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outcome := runChecksumScenario(t, tc.scenario)

			if !errors.Is(outcome.err, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, outcome.err)
			}
			if outcome.applyCalls != 0 || !outcome.updateDirGone {
				t.Fatalf("expected nothing applied and temp dir removed, got applyCalls=%d dirGone=%v", outcome.applyCalls, outcome.updateDirGone)
			}
		})
	}
}

func TestDownloadAndApply_PackageAboveLimitIsRefused(t *testing.T) {
	resetServiceFns()
	t.Cleanup(resetServiceFns)
	originalLimit := maxUpdatePackageBytes
	maxUpdatePackageBytes = 8
	t.Cleanup(func() { maxUpdatePackageBytes = originalLimit })

	oversizedBody := strings.Repeat("x", 100)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sums" {
			_, _ = w.Write([]byte(checksumListFor(oversizedBody, checksumTestAssetName)))
			return
		}
		_, _ = w.Write([]byte(oversizedBody))
	}))
	t.Cleanup(server.Close)

	getAssetNameFunc = func() string { return checksumTestAssetName }
	fetchLatestReleaseFunc = func() (GitHubRelease, error) {
		return GitHubRelease{Assets: []GitHubAsset{
			{Name: checksumTestAssetName, Size: 4, BrowserDownloadURL: server.URL + "/pkg"},
			{Name: checksumsAssetName, BrowserDownloadURL: server.URL + "/sums"},
		}}, nil
	}
	applyCalls := 0
	applyFullUpdateFunc = func(string) error { applyCalls++; return nil }

	err := NewService().DownloadAndApply()

	if !errors.Is(err, errPackageTooLarge) {
		t.Fatalf("expected errPackageTooLarge, got %v", err)
	}
	if applyCalls != 0 {
		t.Fatalf("expected nothing applied, got %d calls", applyCalls)
	}
	if _, statErr := os.Stat(filepath.Join(os.TempDir(), "kuranas-update")); !os.IsNotExist(statErr) {
		t.Fatalf("expected temp dir removed")
	}
}

func TestDownloadFile_AboveLimitRemovesFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 50)))
	}))
	t.Cleanup(server.Close)
	dest := filepath.Join(t.TempDir(), "package.zip")

	err := downloadFile(server.URL, dest, 10)

	if !errors.Is(err, errPackageTooLarge) {
		t.Fatalf("expected errPackageTooLarge, got %v", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("expected oversized file removed")
	}
}

func TestParseChecksumEntry(t *testing.T) {
	list := fmt.Sprintf("%s  kuranas-linux.zip\r\n%s *kuranas-windows.zip\nAAAA other.zip extra\n%s one-space.zip\n",
		strings.Repeat("a", 64), strings.Repeat("B", 64), strings.Repeat("c", 64))

	cases := []struct {
		name     string
		fileName string
		expected string
		found    bool
	}{
		{"two spaces", "kuranas-linux.zip", strings.Repeat("a", 64), true},
		{"binary marker and uppercase hash", "kuranas-windows.zip", strings.Repeat("b", 64), true},
		{"one space", "one-space.zip", strings.Repeat("c", 64), true},
		{"malformed line ignored", "other.zip", "", false},
		{"absent file", "missing.zip", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hash, found := parseChecksumEntry(list, tc.fileName)
			if hash != tc.expected || found != tc.found {
				t.Fatalf("expected (%q,%v), got (%q,%v)", tc.expected, tc.found, hash, found)
			}
		})
	}
}
