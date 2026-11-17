package video

import (
	"os"
	"testing"
)

func TestHasVideoMasterFrameDetectsGeneratedFrame(t *testing.T) {
	const fileID = 987654322
	if HasVideoMasterFrame(fileID) {
		t.Fatalf("expected no master frame before it is generated")
	}
	framePath := masterFramePath(fileID)
	if err := os.MkdirAll(videoThumbnailCacheDir(), 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}
	if err := os.WriteFile(framePath, []byte("jpeg"), 0o644); err != nil {
		t.Fatalf("write master frame: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(framePath) })

	if !HasVideoMasterFrame(fileID) {
		t.Fatalf("expected the generated master frame to be found")
	}
}
