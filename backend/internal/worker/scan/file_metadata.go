package scan

import (
	"context"
	"encoding/json"
	"errors"
	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	musicdom "nas-go/api/internal/api/v1/music"
	videodom "nas-go/api/internal/api/v1/video"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/img"
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

var imageStillConverter img.StillImageConverter = img.NewFFmpegStillConverter()

func SetImageStillConverterForTesting(converter img.StillImageConverter) {
	if converter == nil {
		imageStillConverter = img.NewFFmpegStillConverter()
		return
	}
	imageStillConverter = converter
}

func getImageMetadata(fileDto files.FileDto, runner ScriptRunner, aiService ai.ServiceInterface) (imagedom.MetadataModel, error) {
	metadata := imagedom.MetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	contentPath := fileDto.ResolveContentPath()
	scriptErr := loadImageScriptMetadata(&metadata, runner, contentPath)
	if errors.Is(scriptErr, context.DeadlineExceeded) {
		return metadata, scriptErr
	}
	if scriptErr != nil {
		applog.Warn("scan: image metadata script unusable, falling back to decoded preview", "path", contentPath, "error", scriptErr.Error())
	}

	fillMissingImageDimensions(&metadata, contentPath, fileDto.Format)
	metadata.Classification = imagedom.ClassifyImageWithAI(fileDto, metadata, aiService)

	return metadata, nil
}

func loadImageScriptMetadata(metadata *imagedom.MetadataModel, runner ScriptRunner, contentPath string) error {
	scriptOutput, err := runner(utils.ImageMetadata, contentPath)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(scriptOutput), metadata)
}

func fillMissingImageDimensions(metadata *imagedom.MetadataModel, contentPath string, extension string) {
	if metadata.Width > 0 && metadata.Height > 0 {
		return
	}
	width, height, err := img.PreviewDimensions(contentPath, extension, imageStillConverter)
	if err != nil {
		applog.Warn("scan: could not read image dimensions from preview", "path", contentPath, "error", err.Error())
		return
	}
	metadata.Width, metadata.Height = width, height
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
