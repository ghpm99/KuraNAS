package worker_test

import (
	"database/sql"
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/logger"
	"path/filepath"
	"testing"
)

// --- Mocks ---

const (
	testDir                  = "testscan"
	expectedCreateFilesCount = 2
	expectedUpdateFilesCount = 10
)

var expectedCreateFiles = []files.FileDto{
	{
		Name:       "teste3.xml",
		Path:       filepath.Join(testDir, "teste3.xml"),
		ParentPath: testDir,
		Type:       files.File,
		Format:     ".xml",
		Size:       1205,
		CheckSum:   "8b20534d2b3576285f3a97a755d8fdaad753b0412b4bb61f619d028bd3215869",
	},
	{
		Name:       "teste1.txt",
		Path:       filepath.Join(testDir, "teste1.txt"),
		ParentPath: testDir,
		Type:       files.File,
		Format:     ".txt",
		Size:       1011,
		CheckSum:   "86ba1e506b0afcf353c862c56cdb8b2a2d76741b9f9db309fd2419300821fdd1",
	},
}

var expectedUpdateFiles = []files.FileDto{
	{
		Name:       testDir,
		Path:       filepath.Join(testDir),
		ParentPath: testDir,
		Type:       files.Directory,
		Size:       4096,
		CheckSum:   "8b24624d125d69cdfb82131067b6037895e5861dba3a52771dc60f7395f097c8",
	},
	{
		Name:       "documentos",
		Path:       filepath.Join(testDir, "documentos"),
		ParentPath: testDir,
		Size:       4096,
		CheckSum:   "8f45adb9504dcd3b37deea2c6ef5d13f1744df2f324b58619915b50bd55baa83",
	},
	{
		Name:       "image",
		Path:       filepath.Join(testDir, "image"),
		ParentPath: testDir,
		Size:       4096,
		CheckSum:   "9c8eebeac6df628d7ce274881af1d095333fd802b12d891031acb7b80548ba2f",
	},
	{
		Name:       "testepasta",
		Path:       filepath.Join(testDir, "testepasta"),
		ParentPath: testDir,
		Size:       4096,
		CheckSum:   "cd372fb85148700fa88095e3492d3f9f5beb43e555e5ff26d95f5a6adc36f8e6",
	},
	{
		Name:       "conteudo_teste_transparente.pdf",
		Path:       filepath.Join(testDir, "documentos", "conteudo_teste_transparente.pdf"),
		ParentPath: testDir,
		Format:     ".pdf",
		Size:       2215,
		CheckSum:   "4db0b21e1e5fda35099c98a5234a20f61072d7a8dbabb939c08bc6263037c715",
	},
	{
		Name:       "conteudo_teste.pdf",
		Path:       filepath.Join(testDir, "documentos", "conteudo_teste.pdf"),
		ParentPath: testDir,
		Format:     ".pdf",
		Size:       1773,
		CheckSum:   "c4eda693e6873ca4d8f65ee2a8a60338bcb70c3ecaec2a569aa91b7cb3d0babe",
	},
	{
		Name:       "teste2.pdf",
		Path:       filepath.Join(testDir, "documentos", "teste2.pdf"),
		ParentPath: testDir,
		Format:     ".pdf",
		CheckSum:   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	},
	{
		Name:       "ai-generated-8610368_1280.png",
		Path:       filepath.Join(testDir, "image", "ai-generated-8610368_1280.png"),
		ParentPath: testDir,
		Format:     ".png",
		Size:       2285478,
		CheckSum:   "e4b1e13fedd469166660cd153fbe44c5a874f263216181455c31233e297f241a",
		Metadata: files.ImageMetadataModel{
			ID:                0,
			FileId:            0,
			Path:              filepath.Join(testDir, "image", "ai-generated-8610368_1280.png"),
			Format:            "PNG",
			Mode:              "RGB",
			Width:             964,
			Height:            1280,
			DPIX:              95.9866,
			DPIY:              95.9866,
			XResolution:       0.0,
			YResolution:       0.0,
			ResolutionUnit:    0.0,
			Orientation:       0.0,
			Compression:       0.0,
			Photometric:       0.0,
			ColorSpace:        0.0,
			ComponentsConfig:  "",
			ICCProfile:        "",
			Make:              "",
			Model:             "",
			Software:          "",
			LensModel:         "",
			SerialNumber:      "",
			DateTime:          "",
			DateTimeOriginal:  "",
			DateTimeDigitized: "",
			SubSecTime:        "",
			ExposureTime:      0.0,
			FNumber:           0.0,
			ISO:               0.0,
			ShutterSpeed:      0.0,
			ApertureValue:     0.0,
			BrightnessValue:   0.0,
			ExposureBias:      0.0,
			MeteringMode:      0.0,
			Flash:             0,
			FocalLength:       0.0,
			WhiteBalance:      0.0,
			ExposureProgram:   0.0,
			MaxApertureValue:  0.0,
			GPSLatitude:       0,
			GPSLongitude:      0,
			GPSAltitude:       0,
			GPSDate:           "",
			GPSTime:           "",
			ImageDescription:  "",
			UserComment:       "",
			Copyright:         "",
			Artist:            "",
		},
	},
	{
		Name:       "ChatGPT Image 28 de mar. de 2025, 20_45_52.png",
		Path:       filepath.Join(testDir, "image", "ChatGPT Image 28 de mar. de 2025, 20_45_52.png"),
		ParentPath: testDir,
		Format:     ".png",
		Size:       3180838,
		CheckSum:   "6712dab1ff55592ef8052a362c005e6d82c50e594bec315e4f437c327d052bcc",
		Metadata: files.ImageMetadataModel{
			ID:                0,
			FileId:            0,
			Path:              filepath.Join(testDir, "image", "ChatGPT Image 28 de mar. de 2025, 20_45_52.png"),
			Format:            "PNG",
			Mode:              "RGB",
			Width:             1024,
			Height:            1536,
			DPIX:              0.0,
			DPIY:              0.0,
			XResolution:       0.0,
			YResolution:       0.0,
			ResolutionUnit:    0.0,
			Orientation:       0.0,
			Compression:       0.0,
			Photometric:       0.0,
			ColorSpace:        0.0,
			ComponentsConfig:  "",
			ICCProfile:        "",
			Make:              "",
			Model:             "",
			Software:          "",
			LensModel:         "",
			SerialNumber:      "",
			DateTime:          "",
			DateTimeOriginal:  "",
			DateTimeDigitized: "",
			SubSecTime:        "",
			ExposureTime:      0.0,
			FNumber:           0.0,
			ISO:               0.0,
			ShutterSpeed:      0.0,
			ApertureValue:     0.0,
			BrightnessValue:   0.0,
			ExposureBias:      0.0,
			MeteringMode:      0.0,
			Flash:             0,
			FocalLength:       0.0,
			WhiteBalance:      0.0,
			ExposureProgram:   0.0,
			MaxApertureValue:  0.0,
			GPSLatitude:       0,
			GPSLongitude:      0,
			GPSAltitude:       0,
			GPSDate:           "",
			GPSTime:           "",
			ImageDescription:  "",
			UserComment:       "",
			Copyright:         "",
			Artist:            "",
		},
	},
	{
		Name:       "teste4.mp3",
		Path:       filepath.Join(testDir, "testepasta", "teste4.mp3"),
		ParentPath: testDir,
		Format:     ".mp3",
		CheckSum:   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Metadata: files.AudioMetadataModel{
			ID:                  0,
			FileId:              0,
			Path:                filepath.Join(testDir, "testepasta", "teste4.mp3"),
			Mime:                "",
			Length:              0,
			Bitrate:             0,
			SampleRate:          0,
			Channels:            0,
			BitrateMode:         0,
			EncoderInfo:         "",
			BitDepth:            0,
			Title:               "",
			Artist:              "",
			Album:               "",
			AlbumArtist:         "",
			TrackNumber:         "",
			Genre:               "",
			Composer:            "",
			Year:                "",
			RecordingDate:       "",
			Encoder:             "",
			Publisher:           "",
			OriginalReleaseDate: "",
			OriginalArtist:      "",
			Lyricist:            "",
			Lyrics:              "",
		},
	},
}

type mockService struct {
	files.ServiceInterface
	newFilesList []string
	createFiles  []files.FileDto
	updateFiles  []files.FileDto
}

func (m *mockService) GetFileByNameAndPath(name, path string) (files.FileDto, error) {
	fmt.Println("Mock GetFileByNameAndPath called with name:", name, "and path:", path)
	for _, file := range expectedUpdateFiles {
		if file.Name == name && file.Path == path {
			return file, nil
		}
	}

	return files.FileDto{}, sql.ErrNoRows

}

func (m *mockService) CreateFile(file files.FileDto) (files.FileDto, error) {
	m.createFiles = append(m.createFiles, file)
	return file, nil
}

func (m *mockService) UpdateFile(file files.FileDto) (bool, error) {
	m.updateFiles = append(m.updateFiles, file)
	return true, nil
}

type mockLogger struct {
	logger.LoggerServiceInterface
	createLogCalled           bool
	completeWithSuccessCalled bool
}

func (m *mockLogger) CreateLog(model logger.LoggerModel, _ interface{}) (logger.LoggerModel, error) {
	m.createLogCalled = true
	return model, nil
}
func (m *mockLogger) CompleteWithSuccessLog(logger.LoggerModel) error {
	m.completeWithSuccessCalled = true
	return nil
}

// --- Test ---

func TestStartFileProcessingPipeline(t *testing.T) {
	// Setup test directory

	fmt.Println("Setting up test directory...")
	config.AppConfig.EntryPoint = testDir
	fmt.Println("Test directory set to:", testDir)

	mockSvc := &mockService{
		newFilesList: []string{"teste3.xml", "teste1.txt"},
	}
	mockLog := &mockLogger{}

	fmt.Println("Starting file processing pipeline...")
	worker.StartFileProcessingPipeline(mockSvc, mockLog)
	fmt.Println("File processing pipeline started.")

	if mockSvc.createFiles == nil && mockSvc.updateFiles == nil {
		t.Error("No files were processed, expected some files to be created or updated.")
	}

	if len(mockSvc.createFiles) == 0 && len(mockSvc.updateFiles) == 0 {
		t.Error("No files were created or updated, expected some files to be processed.")
	}

	fmt.Println("Files processed:")
	if len(mockSvc.createFiles) != expectedCreateFilesCount {
		t.Errorf("Expected %d files to be created, got: %d", expectedCreateFilesCount, len(mockSvc.createFiles))
	}

	if len(mockSvc.updateFiles) != expectedUpdateFilesCount {
		t.Errorf("Expected %d files to be updated, got: %d", expectedUpdateFilesCount, len(mockSvc.updateFiles))
	}

	compareFileDtoSlices(t, expectedCreateFiles, mockSvc.createFiles)
	compareFileDtoSlices(t, expectedUpdateFiles, mockSvc.updateFiles)

	if !mockLog.createLogCalled {
		t.Error("Logger.CreateLog was not called")
	}
	if !mockLog.completeWithSuccessCalled {
		t.Error("Logger.CompleteWithSuccessLog was not called")
	}
}

func compareFileDtoSlices(t *testing.T, expected, actual []files.FileDto) {
	if len(expected) != len(actual) {
		t.Errorf("Expected %d files, got %d", len(expected), len(actual))
		return
	}

	for _, expectedFile := range expected {
		found := false
		fileDtoCreated := files.FileDto{}
		for _, actualFile := range actual {
			if expectedFile.Name == actualFile.Name && expectedFile.Path == actualFile.Path {
				found = true
				fileDtoCreated = actualFile
				break
			}
		}
		if !found {
			t.Errorf("Expected file not found: %s", expectedFile.Name)
		}
		compareFileDto(t, expectedFile, fileDtoCreated)
	}
}

func compareFileDto(t *testing.T, expected, actual files.FileDto) {
	if expected.Name != actual.Name {
		t.Errorf("Expected file %s\nName %s, got %s", actual.Path, expected.Name, actual.Name)
	}
	if expected.Path != actual.Path {
		t.Errorf("Expected file %s\nPath %s, got %s", actual.Path, expected.Path, actual.Path)
	}
	if expected.ParentPath != actual.ParentPath {
		t.Errorf("Expected file %s\nParentPath %s, got %s", actual.Path, expected.ParentPath, actual.ParentPath)
	}
	if expected.Type != actual.Type {
		t.Errorf("Expected file %s\nType %d, got %d", actual.Path, expected.Type, actual.Type)
	}
	if expected.Format != actual.Format {
		t.Errorf("Expected file %s\nFormat %s, got %s", actual.Path, expected.Format, actual.Format)
	}
	if expected.Size != actual.Size {
		t.Errorf("Expected file %s\nSize %d, got %d", actual.Path, expected.Size, actual.Size)
	}
	if expected.CheckSum != actual.CheckSum {
		t.Errorf("Expected file %s\nCheckSum %s, got %s", actual.Path, expected.CheckSum, actual.CheckSum)
	}
	if expected.DirectoryContentCount != actual.DirectoryContentCount {
		t.Errorf("Expected file %s\nDirectoryContentCount %d, got %d", actual.Path, expected.DirectoryContentCount, actual.DirectoryContentCount)
	}
	if expected.Starred != actual.Starred {
		t.Errorf("Expected file %s\nStarred %t, got %t", actual.Path, expected.Starred, actual.Starred)
	}

	compareMetadata(t, expected.Metadata, actual.Metadata)

}

func compareMetadata(t *testing.T, expected, actual any) {
	if expected == nil && actual == nil {
		return
	}
	if expected == nil || actual == nil {
		t.Errorf("Expected metadata to be non-nil, got %v", actual)
		return
	}

	switch expected := expected.(type) {
	case files.ImageMetadataModel:
		actualMetadata, ok := actual.(files.ImageMetadataModel)
		if !ok {
			t.Errorf("Expected ImageMetadataModel, got %T", actual)
			return
		}
		if expected != actualMetadata {
			t.Errorf("Image metadata mismatch: expected %v, got %v", expected, actualMetadata)
		}
	case files.AudioMetadataModel:
		actualMetadata, ok := actual.(files.AudioMetadataModel)
		if !ok {
			t.Errorf("Expected AudioMetadataModel, got %T", actual)
			return
		}
		if expected != actualMetadata {
			t.Errorf("Audio metadata mismatch: expected %v, got %v", expected, actualMetadata)
		}
	case files.VideoMetadataModel:
		actualMetadata, ok := actual.(files.VideoMetadataModel)
		if !ok {
			t.Errorf("Expected VideoMetadataModel, got %T", actual)
			return
		}
		if expected != actualMetadata {
			t.Errorf("Video metadata mismatch: expected %v, got %v", expected, actualMetadata)
		}
	default:
		t.Errorf("Unknown metadata type: %T", expected)
	}
}
