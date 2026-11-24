package video

import (
	"math"
	"strconv"
	"strings"
)

const (
	remuxKeyframeLookbackSeconds  = 30.0
	remuxKeyframeLookaheadSeconds = 1.0
	remuxKeyframeMatchTolerance   = 0.001
	remuxKeyframeSeekNudgeSeconds = 0.5
)

type RemuxStartDto struct {
	StartSeconds float64 `json:"start_seconds"`
}

func buildRemuxStartProbeArguments(sourcePath string, atSeconds float64) []string {
	windowStart := math.Max(0, atSeconds-remuxKeyframeLookbackSeconds)
	readInterval := formatSeconds(windowStart) + "%" + formatSeconds(atSeconds+remuxKeyframeLookaheadSeconds)
	return []string{
		"-v", "error",
		"-select_streams", "v:0",
		"-read_intervals", readInterval,
		"-show_entries", "packet=pts_time,flags",
		"-of", "csv=p=0",
		sourcePath,
	}
}

func parseRemuxStartSeconds(probeOutput []byte, atSeconds float64) float64 {
	keyframeSeconds := 0.0
	hasKeyframe := false
	for _, line := range strings.Split(string(probeOutput), "\n") {
		packetSeconds, isKeyframe := parseKeyframePacket(line)
		if !isKeyframe || packetSeconds > atSeconds+remuxKeyframeMatchTolerance {
			continue
		}
		if !hasKeyframe || packetSeconds > keyframeSeconds {
			keyframeSeconds = packetSeconds
			hasKeyframe = true
		}
	}
	if !hasKeyframe {
		return atSeconds
	}
	return keyframeSeconds
}

func parseKeyframePacket(line string) (packetSeconds float64, isKeyframe bool) {
	fields := strings.Split(strings.TrimSpace(line), ",")
	if len(fields) < 2 || !strings.Contains(fields[1], "K") {
		return 0, false
	}
	packetSeconds, err := strconv.ParseFloat(fields[0], 64)
	return packetSeconds, err == nil
}
