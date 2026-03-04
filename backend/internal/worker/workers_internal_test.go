package worker

import (
	"encoding/json"
	"errors"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestGetCheckSum(t *testing.T) {
	file := files.FileDto{Type: files.File, Path: "/tmp/a"}
	dir := files.FileDto{Type: files.Directory, Path: "/tmp/d"}

	fileHash, err := getCheckSum(
		file,
		func(path string) (string, error) { return "file-hash", nil },
		func(path string) (string, error) { return "dir-hash", nil },
	)
	if err != nil || fileHash != "file-hash" {
		t.Fatalf("expected file hash, got %q err=%v", fileHash, err)
	}

	dirHash, err := getCheckSum(
		dir,
		func(path string) (string, error) { return "file-hash", nil },
		func(path string) (string, error) { return "dir-hash", nil },
	)
	if err != nil || dirHash != "dir-hash" {
		t.Fatalf("expected dir hash, got %q err=%v", dirHash, err)
	}

	_, err = getCheckSum(
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
	monitor := make(chan ResultWorkerData, 3)
	var wg sync.WaitGroup

	in <- files.FileDto{ID: 1, Type: files.File, Path: "/tmp/file"}
	in <- files.FileDto{ID: 2, Type: files.Directory, Path: "/tmp/dir"}
	in <- files.FileDto{ID: 3, Type: files.FileType(99), Path: "/tmp/unknown"}
	close(in)

	wg.Add(1)
	go StartChecksumWorker(
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

	dto := convertToDto(FileWalk{Path: filePath, Info: info})
	if dto.Path != filePath || dto.Name == "" {
		t.Fatalf("unexpected dto conversion result: %+v", dto)
	}

	in := make(chan FileWalk, 1)
	out := make(chan files.FileDto, 1)
	var wg sync.WaitGroup
	in <- FileWalk{Path: filePath, Info: info}
	close(in)
	wg.Add(1)
	go StartDtoConverterWorker(in, out, &wg)
	wg.Wait()
	close(out)

	count := 0
	for range out {
		count++
	}
	if count != 1 {
		t.Fatalf("expected one dto from converter")
	}

	monitor := make(chan ResultWorkerData, 2)
	monitor <- ResultWorkerData{Path: "ok", Success: true}
	monitor <- ResultWorkerData{Path: "err", Success: false, Error: "boom"}
	close(monitor)
	wg.Add(1)
	go StartResultMonitorWorker(monitor, &wg)
	wg.Wait()
}

func TestMetadataWorkerAndHelpers(t *testing.T) {
	runner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		switch scriptType {
		case utils.ImageMetadata:
			b, _ := json.Marshal(files.ImageMetadataModel{Format: "PNG", Path: filePath})
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

	imgMeta, err := getImageMetadata(files.FileDto{ID: 1, Path: "/img.png"}, runner)
	if err != nil || imgMeta.Format != "PNG" {
		t.Fatalf("expected image metadata, err=%v", err)
	}
	audioMeta, err := getAudioMetadata(files.FileDto{ID: 1, Path: "/a.mp3"}, runner)
	if err != nil || audioMeta.Mime != "mp3" {
		t.Fatalf("expected audio metadata, err=%v", err)
	}
	videoMeta, err := getVideoMetadata(files.FileDto{ID: 1, Path: "/v.mp4"}, runner)
	if err != nil || videoMeta.FormatName != "mp4" {
		t.Fatalf("expected video metadata, err=%v", err)
	}

	if meta, err := getMetadata(files.FileDto{Format: ".txt"}, runner); err != nil || meta != nil {
		t.Fatalf("expected nil metadata for unsupported format, got meta=%v err=%v", meta, err)
	}

	in := make(chan files.FileDto, 2)
	out := make(chan files.FileDto, 2)
	monitor := make(chan ResultWorkerData, 2)
	var wg sync.WaitGroup

	in <- files.FileDto{ID: 1, Path: "/x.png", Format: ".png", Type: files.File}
	in <- files.FileDto{ID: 2, Path: "/x.txt", Format: ".txt", Type: files.File}
	close(in)

	wg.Add(1)
	go StartMetadataWorker(in, out, runner, monitor, &wg)
	wg.Wait()
	close(out)
	close(monitor)

	processed := 0
	for item := range out {
		processed++
		if item.Format == ".png" && item.Metadata == nil {
			t.Fatalf("expected metadata for png")
		}
	}
	if processed != 2 {
		t.Fatalf("expected 2 processed files, got %d", processed)
	}
}

func TestStartDirectoryWalker(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "f.txt")
	if err := os.WriteFile(filePath, []byte("abc"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	fileWalkChannel := make(chan FileWalk, 10)
	monitor := make(chan ResultWorkerData, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go StartDirectoryWalker(tmpDir, fileWalkChannel, monitor, &wg)
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
	errCh := make(chan FileWalk, 1)
	monErr := make(chan ResultWorkerData, 2)
	wg.Add(1)
	go StartDirectoryWalker(filepath.Join(tmpDir, "missing"), errCh, monErr, &wg)
	wg.Wait()
	close(errCh)
	close(monErr)
	receivedAny := false
	for range errCh {
		receivedAny = true
	}
	if !receivedAny {
		t.Fatalf("expected at least one callback item for missing path")
	}
}
