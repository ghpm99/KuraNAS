package captures

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	isolatedThumbnailRoot, err := os.MkdirTemp("", "captures-thumbnails-")
	if err != nil {
		panic(err)
	}
	thumbnailRootPath = func() string { return isolatedThumbnailRoot }

	exitCode := m.Run()

	_ = os.RemoveAll(isolatedThumbnailRoot)
	os.Exit(exitCode)
}
