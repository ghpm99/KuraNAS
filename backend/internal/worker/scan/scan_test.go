package scan

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	musicdom "nas-go/api/internal/api/v1/music"
	videodom "nas-go/api/internal/api/v1/video"
	"nas-go/api/pkg/utils"
)

// --- mocks ---

type scanFilesServiceMock struct {
	files.ServiceInterface
	getFilesFn          func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	getFileByNamePathFn func(name, path string) (files.FileDto, error)
	createFileFn        func(fileDto files.FileDto) (files.FileDto, error)
	updateFileFn        func(file files.FileDto) (bool, error)
}

func (m *scanFilesServiceMock) GetFiles(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	if m.getFilesFn != nil {
		return m.getFilesFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[files.FileDto]{}, nil
}
func (m *scanFilesServiceMock) GetFileByNameAndPath(name, path string) (files.FileDto, error) {
	if m.getFileByNamePathFn != nil {
		return m.getFileByNamePathFn(name, path)
	}
	return files.FileDto{}, sql.ErrNoRows
}
func (m *scanFilesServiceMock) CreateFile(fileDto files.FileDto) (files.FileDto, error) {
	if m.createFileFn != nil {
		return m.createFileFn(fileDto)
	}
	return fileDto, nil
}
func (m *scanFilesServiceMock) UpdateFile(file files.FileDto) (bool, error) {
	if m.updateFileFn != nil {
		return m.updateFileFn(file)
	}
	return true, nil
}

// --- checksum ---

func TestGetCheckSum(t *testing.T) {
	file := files.FileDto{Type: files.File, Path: "/tmp/a"}
	dir := files.FileDto{Type: files.Directory, Path: "/tmp/d"}

	fileHash, err := GetCheckSum(
		file,
		func(path string) (string, error) { return "file-hash", nil },
		func(path string) (string, error) { return "dir-hash", nil },
	)
	if err != nil || fileHash != "file-hash" {
		t.Fatalf("expected file hash, got %q err=%v", fileHash, err)
	}

	dirHash, err := GetCheckSum(
		dir,
		func(path string) (string, error) { return "file-hash", nil },
		func(path string) (string, error) { return "dir-hash", nil },
	)
	if err != nil || dirHash != "dir-hash" {
		t.Fatalf("expected dir hash, got %q err=%v", dirHash, err)
	}

	_, err = GetCheckSum(
		files.FileDto{Type: files.FileType(99), Path: "/tmp/x"},
		func(path string) (string, error) { return "", nil },
		func(path string) (string, error) { return "", nil },
	)
	if err == nil {
		t.Fatalf("expected unknown type error")
	}
}

// --- metadata ---

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

	imgMeta, err := GetMetadata(files.FileDto{ID: 1, Path: "/x.jpg", Format: ".jpg"}, runner, nil)
	if err != nil || imgMeta == nil {
		t.Fatalf("expected image metadata dispatch success, err=%v", err)
	}

	audioMeta, err := GetMetadata(files.FileDto{ID: 1, Path: "/x.mp3", Format: ".mp3"}, runner, nil)
	if err != nil || audioMeta == nil {
		t.Fatalf("expected audio metadata dispatch success, err=%v", err)
	}

	videoMeta, err := GetMetadata(files.FileDto{ID: 1, Path: "/x.mp4", Format: ".mp4"}, runner, nil)
	if err != nil || videoMeta == nil {
		t.Fatalf("expected video metadata dispatch success, err=%v", err)
	}

	nilMeta, err := GetMetadata(files.FileDto{Format: ".txt"}, runner, nil)
	if err != nil || nilMeta != nil {
		t.Fatalf("expected nil metadata for unsupported format, got meta=%v err=%v", nilMeta, err)
	}
}

func TestMetadataHelpers(t *testing.T) {
	runner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		switch scriptType {
		case utils.ImageMetadata:
			b, _ := json.Marshal(imagedom.MetadataModel{Format: "PNG", Path: filePath})
			return string(b), nil
		case utils.AudioMetadata:
			b, _ := json.Marshal(musicdom.AudioMetadataModel{Mime: "mp3", Path: filePath})
			return string(b), nil
		case utils.VideoMetadata:
			b, _ := json.Marshal(videodom.VideoMetadataModel{FormatName: "mp4", Path: filePath})
			return string(b), nil
		default:
			return "", errors.New("unknown")
		}
	}

	imgMeta, err := getImageMetadata(files.FileDto{ID: 1, Path: "/img.png"}, runner, nil)
	if err != nil || imgMeta.Format != "PNG" {
		t.Fatalf("expected image metadata, err=%v", err)
	}
	if imgMeta.Classification.Category != imagedom.ClassificationCategoryOther {
		t.Fatalf("expected default image classification, got %s", imgMeta.Classification.Category)
	}
	audioMeta, err := getAudioMetadata(files.FileDto{ID: 1, Path: "/a.mp3"}, runner)
	if err != nil || audioMeta.Mime != "mp3" {
		t.Fatalf("expected audio metadata, err=%v", err)
	}
	videoMeta, err := getVideoMetadata(files.FileDto{ID: 1, Path: "/v.mp4"}, runner)
	if err != nil || videoMeta.FormatName != "mp4" {
		t.Fatalf("expected video metadata, err=%v", err)
	}

	errRunner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "", errors.New("runner failed")
	}
	if _, err := getImageMetadata(files.FileDto{ID: 2, Path: "/err.png"}, errRunner, nil); err == nil {
		t.Fatalf("expected image metadata runner error")
	}
	if _, err := getAudioMetadata(files.FileDto{ID: 2, Path: "/err.mp3"}, errRunner); err == nil {
		t.Fatalf("expected audio metadata runner error")
	}
	if _, err := getVideoMetadata(files.FileDto{ID: 2, Path: "/err.mp4"}, errRunner); err == nil {
		t.Fatalf("expected video metadata runner error")
	}
	if _, err := getAudioMetadata(files.FileDto{ID: 3, Path: "/bad.mp3"}, func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{invalid-json", nil
	}); err == nil {
		t.Fatalf("expected audio metadata json parse error")
	}
	if _, err := getVideoMetadata(files.FileDto{ID: 3, Path: "/bad.mp4"}, func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{invalid-json", nil
	}); err == nil {
		t.Fatalf("expected video metadata json parse error")
	}
}

func TestSetPythonScriptRunnerForTesting(t *testing.T) {
	custom := func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "custom", nil
	}
	SetPythonScriptRunnerForTesting(custom)
	if out, _ := PythonScriptRunner(utils.ImageMetadata, "/x"); out != "custom" {
		t.Fatalf("expected custom runner to be installed")
	}
	SetPythonScriptRunnerForTesting(nil)
}

// --- database persistence helpers ---

func TestCreateAndUpdateFileRecord(t *testing.T) {
	svc := &scanFilesServiceMock{
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			fileDto.ID = 10
			return fileDto, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) { return true, nil },
	}

	created, err := CreateFileRecord(svc, files.FileDto{Name: "a.txt"})
	if err != nil || created.ID != 10 {
		t.Fatalf("CreateFileRecord failed, created=%+v err=%v", created, err)
	}

	ok, err := UpdateFileRecord(svc, files.FileDto{Name: "a", Format: ".txt", CheckSum: "abc"}, files.FileDto{ID: 1})
	if err != nil || !ok {
		t.Fatalf("UpdateFileRecord failed, ok=%v err=%v", ok, err)
	}
}
