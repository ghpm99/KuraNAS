package files

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// UploadOffsetMismatchError is returned when a chunk does not start where the
// server stopped receiving; ReceivedSize tells the client where to realign.
type UploadOffsetMismatchError struct {
	ReceivedSize int64
	RequestedAt  int64
}

func (e *UploadOffsetMismatchError) Error() string {
	return fmt.Sprintf("chunk offset %d differs from received size %d", e.RequestedAt, e.ReceivedSize)
}

// UploadChunk writes one chunk at its exact offset of the session partial file
// and returns the new received size. Nothing is written when the offset does
// not match the received size or the chunk would exceed the declared size.
func (s *Service) UploadChunk(uploadID string, offset int64, chunk *multipart.FileHeader) (int64, error) {
	if err := requireValidUploadID(uploadID); err != nil {
		return 0, err
	}
	unlock := lockUploadSession(uploadID)
	defer unlock()

	session, err := loadUploadSession(uploadID)
	if err != nil {
		return 0, err
	}
	if offset != session.ReceivedSize {
		return 0, &UploadOffsetMismatchError{ReceivedSize: session.ReceivedSize, RequestedAt: offset}
	}
	if session.ReceivedSize+chunk.Size > session.Size {
		return 0, newFileOperationError(
			http.StatusBadRequest,
			"ERROR_UPLOAD_CHUNK_EXCEEDS_SIZE",
			fmt.Errorf("chunk of %d bytes at %d exceeds declared size %d", chunk.Size, offset, session.Size),
		)
	}

	writtenBytes, err := appendChunkToPartialFile(uploadID, session.ReceivedSize, chunk)
	if err != nil {
		return 0, uploadInternalError("write upload chunk", err)
	}

	session.ReceivedSize += writtenBytes
	session.LastActivityAt = time.Now()
	if err := saveUploadSession(session); err != nil {
		return 0, uploadInternalError("save upload session", err)
	}
	return session.ReceivedSize, nil
}

func appendChunkToPartialFile(uploadID string, offset int64, chunk *multipart.FileHeader) (int64, error) {
	source, err := chunk.Open()
	if err != nil {
		return 0, err
	}
	defer source.Close()

	partialFile, err := os.OpenFile(uploadSessionPartialPath(uploadID), os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer partialFile.Close()

	if err := partialFile.Truncate(offset); err != nil {
		return 0, err
	}
	if _, err := partialFile.Seek(offset, io.SeekStart); err != nil {
		return 0, err
	}
	writtenBytes, err := io.Copy(partialFile, source)
	if err != nil {
		return 0, err
	}
	return writtenBytes, partialFile.Close()
}
