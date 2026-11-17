package files

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"
)

type UploadCompleteResult struct {
	File  UploadFileResult
	JobID int
}

// CompleteUpload publishes a fully received upload at its destination,
// registers it in the index and enqueues its processing. A size or checksum
// that differs from what init declared discards the session and answers 400.
func (s *Service) CompleteUpload(uploadID string) (UploadCompleteResult, error) {
	if err := requireValidUploadID(uploadID); err != nil {
		return UploadCompleteResult{}, err
	}
	unlock := lockUploadSession(uploadID)
	defer unlock()

	session, err := loadUploadSession(uploadID)
	if err != nil {
		return UploadCompleteResult{}, err
	}
	if err := verifyReceivedUpload(session); err != nil {
		if deleteErr := deleteUploadSession(uploadID); deleteErr != nil {
			return UploadCompleteResult{}, deleteErr
		}
		return UploadCompleteResult{}, err
	}

	fileResult, err := s.publishUploadSession(session)
	if err != nil {
		return UploadCompleteResult{}, err
	}
	if err := deleteUploadSession(uploadID); err != nil {
		return UploadCompleteResult{}, err
	}
	if !isWrittenStatus(fileResult.Status) {
		return UploadCompleteResult{File: fileResult}, nil
	}

	jobID, err := s.createUploadJobWhenNeeded([]string{fileResult.Path})
	if err != nil {
		return UploadCompleteResult{}, err
	}
	return UploadCompleteResult{File: fileResult, JobID: jobID}, nil
}

func verifyReceivedUpload(session UploadSession) error {
	partialPath := uploadSessionPartialPath(session.UploadID)
	partialInfo, err := os.Stat(partialPath)
	if err != nil {
		return uploadInternalError("stat upload partial", err)
	}
	if session.ReceivedSize != session.Size || partialInfo.Size() != session.Size {
		return newFileOperationError(
			http.StatusBadRequest,
			"ERROR_UPLOAD_INCOMPLETE",
			fmt.Errorf("received %d of %d bytes", partialInfo.Size(), session.Size),
		)
	}
	if session.Checksum == "" {
		return nil
	}
	receivedChecksum, err := sha256HexOfFile(partialPath)
	if err != nil {
		return uploadInternalError("hash upload partial", err)
	}
	if receivedChecksum != session.Checksum {
		return newFileOperationError(
			http.StatusBadRequest,
			"ERROR_UPLOAD_CHECKSUM_MISMATCH",
			fmt.Errorf("checksum %s differs from declared %s", receivedChecksum, session.Checksum),
		)
	}
	return nil
}

func (s *Service) publishUploadSession(session UploadSession) (UploadFileResult, error) {
	directorySegments, err := parseRelativeDirectorySegments(session.RelativePath)
	if err != nil {
		return UploadFileResult{}, err
	}
	pending := pendingUpload{
		fileName:          session.FileName,
		directorySegments: directorySegments,
		publish: func(destinationPath string) error {
			return s.publishPartialFile(uploadSessionPartialPath(session.UploadID), destinationPath, session)
		},
	}
	return s.storeUploadedFile(session.TargetFolderPath, pending, session.OnConflict)
}

func (s *Service) publishPartialFile(partialPath string, destinationPath string, session UploadSession) error {
	if err := s.movePartialFile(partialPath, destinationPath); err != nil {
		return err
	}
	if session.ModifiedAt == nil {
		return nil
	}
	return os.Chtimes(destinationPath, *session.ModifiedAt, *session.ModifiedAt)
}

func (s *Service) movePartialFile(partialPath string, destinationPath string) error {
	rename := s.RenamePartialFile
	if rename == nil {
		rename = os.Rename
	}
	err := rename(partialPath, destinationPath)
	if !errors.Is(err, syscall.EXDEV) {
		return err
	}
	return copyThroughTemporaryFile(partialPath, destinationPath)
}

func copyThroughTemporaryFile(partialPath string, destinationPath string) error {
	temporaryPath := uploadTemporaryPathFor(destinationPath)
	if err := copyFileContent(partialPath, temporaryPath); err != nil {
		os.Remove(temporaryPath)
		return err
	}
	if err := os.Rename(temporaryPath, destinationPath); err != nil {
		os.Remove(temporaryPath)
		return err
	}
	return nil
}

func copyFileContent(sourcePath string, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(destination, source); err != nil {
		destination.Close()
		return err
	}
	return destination.Close()
}
