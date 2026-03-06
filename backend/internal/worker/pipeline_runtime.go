package worker

import (
	"nas-go/api/pkg/utils"
	"os"
)

type FileWalk struct {
	Path string
	Info os.FileInfo
}

type ResultWorkerData struct {
	Path    string
	Success bool
	Error   string
}

var pythonScriptRunner = func(scriptType utils.ScriptType, filePath string) (string, error) {
	return utils.RunPythonScript(scriptType, filePath)
}

func SetPythonScriptRunnerForTesting(runner func(scriptType utils.ScriptType, filePath string) (string, error)) {
	if runner == nil {
		pythonScriptRunner = func(scriptType utils.ScriptType, filePath string) (string, error) {
			return utils.RunPythonScript(scriptType, filePath)
		}
		return
	}

	pythonScriptRunner = runner
}
