package engine

import (
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	ingestapi "nas-go/api/internal/api/v1/ingest"
)

const downloadLineMarker = "[download]"

var (
	downloadPercentRe     = regexp.MustCompile(`(\d{1,3}(?:\.\d+)?)%`)
	downloadSpeedRe       = regexp.MustCompile(`\bat\s+(\d+(?:\.\d+)?)\s*([KMGT]?i?B)/s`)
	downloadETARe         = regexp.MustCompile(`\bETA\s+((?:\d+:)?\d{1,2}:\d{2})\b`)
	downloadDestinationRe = regexp.MustCompile(`^\[download\]\s+Destination:\s+(.+)$`)
)

var byteUnitMultipliers = map[string]float64{
	"B":   1,
	"KB":  1000,
	"MB":  1000 * 1000,
	"GB":  1000 * 1000 * 1000,
	"TB":  1000 * 1000 * 1000 * 1000,
	"KiB": 1 << 10,
	"MiB": 1 << 20,
	"GiB": 1 << 30,
	"TiB": 1 << 40,
}

// parseDownloadProgressLine reads a yt-dlp `[download]` progress line such as
// "[download]  42.3% of ~12.34MiB at 1.23MiB/s ETA 00:12". Speed and ETA stay
// nil when the line reports them as unknown or omits them.
func parseDownloadProgressLine(line string) (ingestapi.DownloadProgress, bool) {
	if !strings.Contains(line, downloadLineMarker) {
		return ingestapi.DownloadProgress{}, false
	}
	percentMatch := downloadPercentRe.FindStringSubmatch(line)
	if percentMatch == nil {
		return ingestapi.DownloadProgress{}, false
	}
	percent, err := strconv.ParseFloat(percentMatch[1], 64)
	if err != nil {
		return ingestapi.DownloadProgress{}, false
	}
	return ingestapi.DownloadProgress{
		Percent:             int(math.Min(math.Max(percent, 0), 100)),
		SpeedBytesPerSecond: parseDownloadSpeed(line),
		ETASeconds:          parseDownloadETA(line),
	}, true
}

func parseDownloadSpeed(line string) *int64 {
	speedMatch := downloadSpeedRe.FindStringSubmatch(line)
	if speedMatch == nil {
		return nil
	}
	amount, err := strconv.ParseFloat(speedMatch[1], 64)
	if err != nil {
		return nil
	}
	bytesPerSecond := int64(amount * byteUnitMultipliers[speedMatch[2]])
	return &bytesPerSecond
}

func parseDownloadETA(line string) *int {
	etaMatch := downloadETARe.FindStringSubmatch(line)
	if etaMatch == nil {
		return nil
	}
	etaSeconds := 0
	for _, clockPart := range strings.Split(etaMatch[1], ":") {
		clockValue, err := strconv.Atoi(clockPart)
		if err != nil {
			return nil
		}
		etaSeconds = etaSeconds*60 + clockValue
	}
	return &etaSeconds
}

// parseDownloadDestinationTitle reads the file name yt-dlp announced with
// "[download] Destination: <path>" and returns it without directory or extension.
func parseDownloadDestinationTitle(line string) (string, bool) {
	destinationMatch := downloadDestinationRe.FindStringSubmatch(strings.TrimSpace(line))
	if destinationMatch == nil {
		return "", false
	}
	title := titleFromFileName(filepath.Base(destinationMatch[1]))
	return title, title != ""
}

func titleFromFileName(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
