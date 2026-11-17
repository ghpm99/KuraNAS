package takeout

import (
	"nas-go/api/pkg/applog"
	"os"
	"strconv"
	"strings"
	"time"
)

func (metadata TakeoutMetadata) resolveTakenTime() (time.Time, bool) {
	for _, timestamp := range []takeoutTimestamp{metadata.PhotoTakenTime, metadata.CreationTime} {
		if takenTime, isValid := timestamp.parseUnixSeconds(); isValid {
			return takenTime, true
		}
	}
	return time.Time{}, false
}

func (timestamp takeoutTimestamp) parseUnixSeconds() (time.Time, bool) {
	unixSeconds, err := strconv.ParseInt(strings.TrimSpace(timestamp.Timestamp), 10, 64)
	if err != nil || unixSeconds <= 0 {
		return time.Time{}, false
	}
	return time.Unix(unixSeconds, 0), true
}

func applyTakenTimeToFile(filePath string, metadata *TakeoutMetadata) {
	if metadata == nil {
		return
	}
	takenTime, hasTakenTime := metadata.resolveTakenTime()
	if !hasTakenTime {
		return
	}
	if err := os.Chtimes(filePath, time.Now(), takenTime); err != nil {
		applog.Warn("takeout: failed to set taken time on extracted file", "path", filePath, "error", err.Error())
	}
}
