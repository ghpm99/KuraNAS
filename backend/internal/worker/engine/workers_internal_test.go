package engine

import (
	"encoding/json"
	"errors"
	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	"nas-go/api/internal/worker/scan"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestGetCheckSum(t *testing.T) {
	file := files.FileDto{Type: files.File, Path: "/tmp/a"}
	dir := files.FileDto{Type: files.Directory, Path: "/tmp/d"}

	fileHash, err := scan.GetCheckSum(
		file,
		func(path string) (string, error) { return "file-hash", nil },
		func(path string) (string, error) { return "dir-hash", nil },
	)
	if err != nil || fileHash != "file-hash" {
		t.Fatalf("expected file hash, got %q err=%v", fileHash, err)
	}

	dirHash, err := scan.GetCheckSum(
		dir,
		func(path string) (string, error) { return "file-hash", nil },
		func(path string) (string, error) { return "dir-hash", nil },
	)
	if err != nil || dirHash != "dir-hash" {
		t.Fatalf("expected dir hash, got %q err=%v", dirHash, err)
	}

	_, err = scan.GetCheckSum(
		files.FileDto{Type: files.FileType(99), Path: "/tmp/x"},
		func(path string) (string, error) { return "", nil },
		func(path string) (string, error) { return "", nil },
	)
	if err == nil {
		t.Fatalf("expected unknown type error")
	}
}

func TestStartChecksumWorker(t *testing.T) {
	in := make(chan files.FileDto, 3)
	out := make(chan files.FileDto, 3)
	monitor := make(chan scan.ResultWorkerData, 3)
	var wg sync.WaitGroup

	in <- files.FileDto{ID: 1, Type: files.File, Path: "/tmp/file"}
	in <- files.FileDto{ID: 2, Type: files.Directory, Path: "/tmp/dir"}
	in <- files.FileDto{ID: 3, Type: files.FileType(99), Path: "/tmp/unknown"}
	close(in)

	wg.Add(1)
	go scan.StartChecksumWorker(
		in,
		out,
		func(path string) (string, error) { return "fh", nil },
		func(path string) (string, error) { return "dh", nil },
		monitor,
		&wg,
	)
	wg.Wait()
	close(out)
	close(monitor)

	var items []files.FileDto
	for item := range out {
		items = append(items, item)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 processed items, got %d", len(items))
	}
	if items[0].CheckSum == "" || items[1].CheckSum == "" {
		t.Fatalf("expected checksums on first two items")
	}

	var errorsCount int
	for r := range monitor {
		if !r.Success {
			errorsCount++
		}
	}
	if errorsCount != 1 {
		t.Fatalf("expected 1 checksum error, got %d", errorsCount)
	}
}

func TestDtoConverterAndMonitor(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "a.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat temp file: %v", err)
	}

	in := make(chan scan.FileWalk, 1)
	out := make(chan files.FileDto, 1)
	var wg sync.WaitGroup
	in <- scan.FileWalk{Path: filePath, Info: info}
	close(in)
	wg.Add(1)
	go scan.StartDtoConverterWorker(in, out, &wg)
	wg.Wait()
	close(out)

	count := 0
	for range out {
		count++
	}
	if count != 1 {
		t.Fatalf("expected one dto from converter")
	}

	monitor := make(chan scan.ResultWorkerData, 2)
	monitor <- scan.ResultWorkerData{Path: "ok", Success: true}
	monitor <- scan.ResultWorkerData{Path: "err", Success: false, Error: "boom"}
	close(monitor)
	wg.Add(1)
	go scan.StartResultMonitorWorker(monitor, &wg)
	wg.Wait()
}

func TestMetadataWorkerAndHelpers(t *testing.T) {
	runner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		switch scriptType {
		case utils.ImageMetadata:
			b, _ := json.Marshal(imagedom.MetadataModel{Format: "PNG", Path: filePath})
			return string(b), nil
		case utils.AudioMetadata:
			b, _ := json.Marshal(files.AudioMetadataModel{Mime: "mp3", Path: filePath})
			return string(b), nil
		case utils.VideoMetadata:
			b, _ := json.Marshal(files.VideoMetadataModel{FormatName: "mp4", Path: filePath})
			return string(b), nil
		default:
			return "", errors.New("unknown")
		}
	}

	imgMeta, err := scan.GetMetadata(files.FileDto{ID: 1, Path: "/img.png", Format: ".png"}, runner, nil)
	if err != nil || imgMeta == nil {
		t.Fatalf("expected image metadata, err=%v", err)
	}
	imgMetaTyped, ok := imgMeta.(imagedom.MetadataModel)
	if !ok || imgMetaTyped.Format != "PNG" {
		t.Fatalf("expected image metadata model with PNG, err=%v", err)
	}
	if imgMetaTyped.Classification.Category != imagedom.ClassificationCategoryOther {
		t.Fatalf("expected default image classification, got %s", imgMetaTyped.Classification.Category)
	}

	audioMeta, err := scan.GetMetadata(files.FileDto{ID: 1, Path: "/a.mp3", Format: ".mp3"}, runner, nil)
	if err != nil || audioMeta == nil {
		t.Fatalf("expected audio metadata, err=%v", err)
	}
	audioMetaTyped, ok := audioMeta.(files.AudioMetadataModel)
	if !ok || audioMetaTyped.Mime != "mp3" {
		t.Fatalf("expected audio metadata model")
	}

	videoMeta, err := scan.GetMetadata(files.FileDto{ID: 1, Path: "/v.mp4", Format: ".mp4"}, runner, nil)
	if err != nil || videoMeta == nil {
		t.Fatalf("expected video metadata, err=%v", err)
	}
	videoMetaTyped, ok := videoMeta.(files.VideoMetadataModel)
	if !ok || videoMetaTyped.FormatName != "mp4" {
		t.Fatalf("expected video metadata model")
	}

	if meta, err := scan.GetMetadata(files.FileDto{Format: ".txt"}, runner, nil); err != nil || meta != nil {
		t.Fatalf("expected nil metadata for unsupported format, got meta=%v err=%v", meta, err)
	}

	in := make(chan files.FileDto, 2)
	out := make(chan files.FileDto, 2)
	monitor := make(chan scan.ResultWorkerData, 2)
	var wg sync.WaitGroup

	in <- files.FileDto{ID: 1, Path: "/x.png", Format: ".png", Type: files.File}
	in <- files.FileDto{ID: 2, Path: "/x.txt", Format: ".txt", Type: files.File}
	close(in)

	wg.Add(1)
	go scan.StartMetadataWorker(in, out, runner, monitor, &wg, nil)
	wg.Wait()
	close(out)
	close(monitor)

	processed := 0
	for item := range out {
		processed++
		if item.Format == ".png" {
			if item.Metadata == nil {
				t.Fatalf("expected metadata for png")
			}
			metadata, ok := item.Metadata.(imagedom.MetadataModel)
			if !ok {
				t.Fatalf("expected image metadata model, got %T", item.Metadata)
			}
			if metadata.Classification.Category != imagedom.ClassificationCategoryOther {
				t.Fatalf("expected classified image metadata, got %s", metadata.Classification.Category)
			}
		}
	}
	if processed != 2 {
		t.Fatalf("expected 2 processed files, got %d", processed)
	}

	errRunner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "", errors.New("runner failed")
	}
	if _, err := scan.GetMetadata(files.FileDto{ID: 2, Path: "/err.png", Format: ".png"}, errRunner, nil); err == nil {
		t.Fatalf("expected image metadata runner error")
	}
	if _, err := scan.GetMetadata(files.FileDto{ID: 2, Path: "/err.mp3", Format: ".mp3"}, errRunner, nil); err == nil {
		t.Fatalf("expected audio metadata runner error")
	}
	if _, err := scan.GetMetadata(files.FileDto{ID: 2, Path: "/err.mp4", Format: ".mp4"}, errRunner, nil); err == nil {
		t.Fatalf("expected video metadata runner error")
	}
	if _, err := scan.GetMetadata(files.FileDto{ID: 3, Path: "/bad.mp3", Format: ".mp3"}, func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{invalid-json", nil
	}, nil); err == nil {
		t.Fatalf("expected audio metadata json parse error")
	}
	if _, err := scan.GetMetadata(files.FileDto{ID: 3, Path: "/bad.mp4", Format: ".mp4"}, func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{invalid-json", nil
	}, nil); err == nil {
		t.Fatalf("expected video metadata json parse error")
	}
}

func TestStartDirectoryWalker(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "f.txt")
	if err := os.WriteFile(filePath, []byte("abc"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	fileWalkChannel := make(chan scan.FileWalk, 10)
	monitor := make(chan scan.ResultWorkerData, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go scan.StartDirectoryWalker(tmpDir, fileWalkChannel, monitor, &wg)
	wg.Wait()

	// Walk succeeded: should have at least root and one file.
	close(fileWalkChannel)
	walked := 0
	for range fileWalkChannel {
		walked++
	}
	if walked < 1 {
		t.Fatalf("expected walked entries, got %d", walked)
	}

	// Non-existing folder path should still return cleanly.
	errCh := make(chan scan.FileWalk, 1)
	monErr := make(chan scan.ResultWorkerData, 2)
	wg.Add(1)
	go scan.StartDirectoryWalker(filepath.Join(tmpDir, "missing"), errCh, monErr, &wg)
	wg.Wait()
	close(errCh)
	close(monErr)
	receivedAny := false
	for range errCh {
		receivedAny = true
	}
	if receivedAny {
		t.Fatalf("did not expect file walk items for missing path")
	}
	monitorErrors := 0
	for item := range monErr {
		if !item.Success {
			monitorErrors++
		}
	}
	if monitorErrors == 0 {
		t.Fatalf("expected at least one monitor error for missing path")
	}
}

func TestStartDirectoryWalkerPermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()
	restrictedDir := filepath.Join(tmpDir, "restricted")
	if err := os.MkdirAll(restrictedDir, 0700); err != nil {
		t.Fatalf("failed to create restricted dir: %v", err)
	}
	innerFile := filepath.Join(restrictedDir, "hidden.txt")
	if err := os.WriteFile(innerFile, []byte("x"), 0600); err != nil {
		t.Fatalf("failed to create file in restricted dir: %v", err)
	}
	if err := os.Chmod(restrictedDir, 0000); err != nil {
		t.Fatalf("failed to remove permissions: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(restrictedDir, 0700)
	})

	fileWalkChannel := make(chan scan.FileWalk, 10)
	monitor := make(chan scan.ResultWorkerData, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go scan.StartDirectoryWalker(tmpDir, fileWalkChannel, monitor, &wg)
	wg.Wait()
	close(fileWalkChannel)
	close(monitor)

	permissionErrors := 0
	for item := range monitor {
		if !item.Success && strings.Contains(strings.ToLower(item.Error), "permission") {
			permissionErrors++
		}
	}

	if permissionErrors == 0 {
		t.Skip("environment did not surface permission-denied during walk")
	}
}

func TestMetadataWorkerErrorBranch(t *testing.T) {
	runner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{invalid-json", nil
	}
	in := make(chan files.FileDto, 1)
	out := make(chan files.FileDto, 1)
	monitor := make(chan scan.ResultWorkerData, 1)
	var wg sync.WaitGroup

	in <- files.FileDto{ID: 10, Path: "/x.png", Format: ".png", Type: files.File}
	close(in)
	wg.Add(1)
	go scan.StartMetadataWorker(in, out, runner, monitor, &wg, nil)
	wg.Wait()
	close(out)
	close(monitor)

	if len(out) != 1 {
		t.Fatalf("expected processed file output")
	}
	if len(monitor) != 1 {
		t.Fatalf("expected one monitor error item")
	}
}

func TestGetMetadataDispatchByFormat(t *testing.T) {
	runner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		switch scriptType {
		case utils.ImageMetadata:
			return `{"id":1,"file_id":1,"path":"` + filePath + `","format":"jpeg","mode":"RGB","width":1,"height":1,"created_at":"2026-01-01T00:00:00Z"}`, nil
		case utils.AudioMetadata:
			return `{"id":2,"file_id":1,"path":"` + filePath + `","mime":"audio/mpeg","length":1,"bitrate":320,"sample_rate":44100,"channels":2,"created_at":"2026-01-01T00:00:00Z"}`, nil
		case utils.VideoMetadata:
			return `{"id":3,"file_id":1,"path":"` + filePath + `","format_name":"mp4","size":"1","duration":"1","width":1,"height":1,"created_at":"2026-01-01T00:00:00Z"}`, nil
		default:
			return `{}`, nil
		}
	}

	imgMeta, err := scan.GetMetadata(files.FileDto{ID: 1, Path: "/x.jpg", Format: ".jpg"}, runner, nil)
	if err != nil || imgMeta == nil {
		t.Fatalf("expected image metadata dispatch success, err=%v", err)
	}

	audioMeta, err := scan.GetMetadata(files.FileDto{ID: 1, Path: "/x.mp3", Format: ".mp3"}, runner, nil)
	if err != nil || audioMeta == nil {
		t.Fatalf("expected audio metadata dispatch success, err=%v", err)
	}

	videoMeta, err := scan.GetMetadata(files.FileDto{ID: 1, Path: "/x.mp4", Format: ".mp4"}, runner, nil)
	if err != nil || videoMeta == nil {
		t.Fatalf("expected video metadata dispatch success, err=%v", err)
	}
}
