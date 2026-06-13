package scan

import (
	"encoding/json"
	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	musicdom "nas-go/api/internal/api/v1/music"
	videodom "nas-go/api/internal/api/v1/video"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/utils"
)

type ScriptRunner func(scriptType utils.ScriptType, filePath string) (string, error)

// PythonScriptRunner is the production ScriptRunner used by the metadata step.
var PythonScriptRunner = func(scriptType utils.ScriptType, filePath string) (string, error) {
	return utils.RunPythonScript(scriptType, filePath)
}

func SetPythonScriptRunnerForTesting(runner func(scriptType utils.ScriptType, filePath string) (string, error)) {
	if runner == nil {
		PythonScriptRunner = func(scriptType utils.ScriptType, filePath string) (string, error) {
			return utils.RunPythonScript(scriptType, filePath)
		}
		return
	}

	PythonScriptRunner = runner
}

func GetMetadata(fileDto files.FileDto, runner ScriptRunner, aiService ai.ServiceInterface) (any, error) {
	formatType := utils.GetFormatTypeByExtension(fileDto.Format)

	switch formatType.Type {
	case utils.FormatTypeImage:
		return getImageMetadata(fileDto, runner, aiService)
	case utils.FormatTypeAudio:
		return getAudioMetadata(fileDto, runner)
	case utils.FormatTypeVideo:
		return getVideoMetadata(fileDto, runner)
	default:
		return nil, nil
	}
}

func getImageMetadata(fileDto files.FileDto, runner ScriptRunner, aiService ai.ServiceInterface) (imagedom.MetadataModel, error) {
	metadata := imagedom.MetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	result, err := runner(utils.ImageMetadata, fileDto.ResolveContentPath())
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	metadata.Classification = imagedom.ClassifyImageWithAI(fileDto, metadata, aiService)

	return metadata, nil
}

func getAudioMetadata(fileDto files.FileDto, runner ScriptRunner) (musicdom.AudioMetadataModel, error) {
	metadata := musicdom.AudioMetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	result, err := runner(utils.AudioMetadata, fileDto.ResolveContentPath())
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}

func getVideoMetadata(fileDto files.FileDto, runner ScriptRunner) (videodom.VideoMetadataModel, error) {
	metadata := videodom.VideoMetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	result, err := runner(utils.VideoMetadata, fileDto.ResolveContentPath())
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}
