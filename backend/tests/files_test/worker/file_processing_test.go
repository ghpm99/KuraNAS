package worker_test

import (
	"database/sql"
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/logger"
	"slices"
	"testing"
)

// --- Mocks ---

type mockService struct {
	files.ServiceInterface
	newFilesList []string
	createFiles  []files.FileDto
	updateFiles  []files.FileDto
}

func (m *mockService) GetFileByNameAndPath(name, path string) (files.FileDto, error) {

	if m.newFilesList != nil {
		if slices.Contains(m.newFilesList, name) {
			return files.FileDto{}, sql.ErrNoRows
		}
	}
	return files.FileDto{
		Name: name,
		Path: path,
	}, nil
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

	const (
		testDir                  = "testscan"
		expectedCreateFilesCount = 2
		expectedUpdateFilesCount = 10
	)

	expectedCreateFiles := []files.FileDto{
		{
			Name:       "teste3.xml",
			Path:       testDir + "/teste3.xml",
			ParentPath: testDir,
			Type:       files.File,
			Format:     ".xml",
			Size:       1205,
			CheckSum:   "8b20534d2b3576285f3a97a755d8fdaad753b0412b4bb61f619d028bd3215869",
		},
		{
			Name:       "teste1.txt",
			Path:       testDir + "/teste1.txt",
			ParentPath: testDir,
			Type:       files.File,
			Format:     ".txt",
			Size:       1011,
			CheckSum:   "86ba1e506b0afcf353c862c56cdb8b2a2d76741b9f9db309fd2419300821fdd1",
		},
	}

	expectedUpdateFiles := []files.FileDto{
		{
			Name:       testDir,
			Path:       testDir,
			ParentPath: testDir,
			Type:       files.Directory,
		},
		{
			Name:       "documentos",
			Path:       testDir + "/documentos/",
			ParentPath: testDir,
		},
		{
			Name:       "image",
			Path:       testDir + "/image/",
			ParentPath: testDir,
		},
		{
			Name:       "testepasta",
			Path:       testDir + "/testepasta/",
			ParentPath: testDir,
		},
		{
			Name:       "testepasta",
			Path:       testDir + "/testepasta/",
			ParentPath: testDir,
		},
		{
			Name:       "conteudo_teste_transparente.pdf",
			Path:       testDir + "/documentos/conteudo_teste_transparente.pdf",
			ParentPath: testDir,
		},
		{
			Name:       "conteudo_teste.pdf",
			Path:       testDir + "/documentos/conteudo_teste.pdf",
			ParentPath: testDir,
		},
		{
			Name:       "teste2.pdf",
			Path:       testDir + "/documentos/teste2.pdf",
			ParentPath: testDir,
		},
		{
			Name:       "ai-generated-8610368_1280.png",
			Path:       testDir + "/image/ai-generated-8610368_1280.png",
			ParentPath: testDir,
		},
		{
			Name:       "ChatGPT Image 28 de mar. de 2025, 20_45_52.png",
			Path:       testDir + "/image/ChatGPT Image 28 de mar. de 2025, 20_45_52.png",
			ParentPath: testDir,
		},
	}

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
		t.Errorf("Expected file Name %s, got %s", expected.Name, actual.Name)
	}
	if expected.Path != actual.Path {
		t.Errorf("Expected file Path %s, got %s", expected.Path, actual.Path)
	}
	if expected.ParentPath != actual.ParentPath {
		t.Errorf("Expected file ParentPath %s, got %s", expected.ParentPath, actual.ParentPath)
	}
	if expected.Type != actual.Type {
		t.Errorf("Expected file Type %d, got %d", expected.Type, actual.Type)
	}
	if expected.Format != actual.Format {
		t.Errorf("Expected file Format %s, got %s", expected.Format, actual.Format)
	}
	if expected.Size != actual.Size {
		t.Errorf("Expected file Size %d, got %d", expected.Size, actual.Size)
	}
	if expected.CheckSum != actual.CheckSum {
		t.Errorf("Expected file CheckSum %s, got %s", expected.CheckSum, actual.CheckSum)
	}
	if expected.DirectoryContentCount != actual.DirectoryContentCount {
		t.Errorf("Expected file DirectoryContentCount %d, got %d", expected.DirectoryContentCount, actual.DirectoryContentCount)
	}
	if expected.Starred != actual.Starred {
		t.Errorf("Expected file Starred %t, got %t", expected.Starred, actual.Starred)
	}

	if expected.Metadata != nil && actual.Metadata != nil {
		if expected.Metadata != actual.Metadata {
			t.Errorf("Expected file Metadata %v, got %v", expected.Metadata, actual.Metadata)
		}
	}

}
