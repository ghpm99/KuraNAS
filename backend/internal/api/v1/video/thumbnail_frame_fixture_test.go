package video

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const (
	fixtureMasterWidth  = 160
	fixtureMasterHeight = 120
)

var fixtureMasterColor = color.RGBA{R: 250, G: 250, B: 250, A: 255}

func buildFixtureMasterImage(width, height int) image.Image {
	fixtureImage := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(fixtureImage, fixtureImage.Bounds(), image.NewUniform(fixtureMasterColor), image.Point{}, draw.Src)
	return fixtureImage
}

func writeFrameFixture(outputPath string) error {
	if strings.HasSuffix(outputPath, ".gif") {
		return os.WriteFile(outputPath, []byte("frame"), 0644)
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	return jpeg.Encode(outputFile, buildFixtureMasterImage(fixtureMasterWidth, fixtureMasterHeight), &jpeg.Options{Quality: 95})
}

func removeCachedFilesOfVideo(fileID int) {
	cachedFiles, _ := filepath.Glob(filepath.Join(videoThumbnailCacheDir(), fmt.Sprintf("%d_*", fileID)))
	for _, cachedFile := range cachedFiles {
		_ = os.Remove(cachedFile)
	}
}

type countingFrameRunner struct {
	executions   atomic.Int32
	delay        time.Duration
	shouldFail   bool
	leavePartial bool
}

func (runner *countingFrameRunner) run(_ context.Context, arguments []string) error {
	runner.executions.Add(1)
	time.Sleep(runner.delay)
	outputPath := arguments[len(arguments)-1]
	if runner.leavePartial {
		_ = os.WriteFile(outputPath, []byte("partial"), 0644)
	}
	if runner.shouldFail {
		return fmt.Errorf("ffmpeg failed")
	}
	return writeFrameFixture(outputPath)
}
