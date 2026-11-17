package files

import "nas-go/api/internal/roots"

func toClientUploadFileResult(fileResult UploadFileResult) UploadFileResult {
	fileResult.Path = roots.ToRelativePath(fileResult.Path)
	return fileResult
}

func toClientUploadFileResults(fileResults []UploadFileResult) []UploadFileResult {
	clientResults := make([]UploadFileResult, 0, len(fileResults))
	for _, fileResult := range fileResults {
		clientResults = append(clientResults, toClientUploadFileResult(fileResult))
	}
	return clientResults
}

func toClientPaths(absolutePaths []string) []string {
	clientPaths := make([]string, 0, len(absolutePaths))
	for _, absolutePath := range absolutePaths {
		clientPaths = append(clientPaths, roots.ToRelativePath(absolutePath))
	}
	return clientPaths
}
