package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const checksumsAssetName = "SHA256SUMS"

var (
	maxUpdatePackageBytes int64 = 512 << 20
	maxChecksumListBytes  int64 = 1 << 20
)

var (
	errChecksumMissing  = errors.New("release has no SHA256SUMS asset")
	errChecksumNoEntry  = errors.New("SHA256SUMS has no entry for the update package")
	errChecksumMismatch = errors.New("update package checksum does not match SHA256SUMS")
	errPackageTooLarge  = errors.New("update package exceeds the size limit")
)

func parseChecksumEntry(checksumList, fileName string) (string, bool) {
	for _, line := range strings.Split(checksumList, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if strings.TrimPrefix(fields[1], "*") == fileName {
			return strings.ToLower(fields[0]), true
		}
	}
	return "", false
}

func sha256OfFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func findAssetURL(release GitHubRelease, assetName string) (string, int64, bool) {
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return asset.BrowserDownloadURL, asset.Size, true
		}
	}
	return "", 0, false
}

func fetchExpectedChecksum(release GitHubRelease, packageName, workDir string) (string, error) {
	checksumsURL, _, found := findAssetURL(release, checksumsAssetName)
	if !found {
		return "", errChecksumMissing
	}

	checksumsPath := filepath.Join(workDir, checksumsAssetName)
	if err := downloadFileFunc(checksumsURL, checksumsPath, maxChecksumListBytes); err != nil {
		if errors.Is(err, errPackageTooLarge) {
			return "", fmt.Errorf("SHA256SUMS exceeds %d bytes", maxChecksumListBytes)
		}
		return "", err
	}

	checksumList, err := os.ReadFile(checksumsPath)
	if err != nil {
		return "", err
	}

	expectedChecksum, found := parseChecksumEntry(string(checksumList), packageName)
	if !found {
		return "", errChecksumNoEntry
	}
	return expectedChecksum, nil
}

func verifyPackageChecksum(packagePath, expectedChecksum string) error {
	actualChecksum, err := sha256OfFile(packagePath)
	if err != nil {
		return err
	}
	if actualChecksum != expectedChecksum {
		return errChecksumMismatch
	}
	return nil
}
