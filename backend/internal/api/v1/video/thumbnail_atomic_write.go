package video

import (
	"os"
	"path/filepath"
)

func createTemporaryFileBeside(finalPath string, namePattern string) (string, error) {
	temporaryFile, err := os.CreateTemp(filepath.Dir(finalPath), namePattern)
	if err != nil {
		return "", err
	}
	temporaryPath := temporaryFile.Name()
	if err := temporaryFile.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return "", err
	}
	return temporaryPath, nil
}

func publishTemporaryFile(temporaryPath, finalPath string) error {
	if err := os.Chmod(temporaryPath, 0644); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	if err := os.Rename(temporaryPath, finalPath); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	return nil
}

func writeFileAtomically(finalPath string, content []byte, namePattern string) error {
	temporaryPath, err := createTemporaryFileBeside(finalPath, namePattern)
	if err != nil {
		return err
	}
	if err := os.WriteFile(temporaryPath, content, 0644); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	return publishTemporaryFile(temporaryPath, finalPath)
}
