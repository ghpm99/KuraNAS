package video

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func strongVideoETag(fileID int, fileSize int64, modifiedAt time.Time) string {
	return fmt.Sprintf(`"%d-%d-%x"`, fileID, fileSize, modifiedAt.UnixNano())
}

func ifRangeAllowsPartialContent(ifRangeHeader, eTag string, modifiedAt time.Time) bool {
	ifRangeHeader = strings.TrimSpace(ifRangeHeader)
	if ifRangeHeader == "" {
		return true
	}
	if strings.HasPrefix(ifRangeHeader, `"`) {
		return ifRangeHeader == eTag
	}
	validatorTime, err := http.ParseTime(ifRangeHeader)
	return err == nil && modifiedAt.Truncate(time.Second).Equal(validatorTime)
}

func respondRangeNotSatisfiable(c *gin.Context, fileSize int64) {
	c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
	c.Status(http.StatusRequestedRangeNotSatisfiable)
	c.Writer.WriteHeaderNow()
}

type transferErrorRecorder struct {
	firstError error
}

func (recorder *transferErrorRecorder) record(err error) {
	if err != nil && recorder.firstError == nil {
		recorder.firstError = err
	}
}

type recordingResponseWriter struct {
	gin.ResponseWriter
	recorder *transferErrorRecorder
}

func (writer recordingResponseWriter) Write(payload []byte) (int, error) {
	written, err := writer.ResponseWriter.Write(payload)
	writer.recorder.record(err)
	return written, err
}

type recordingReadSeeker struct {
	io.ReadSeeker
	recorder *transferErrorRecorder
}

func (reader recordingReadSeeker) Read(buffer []byte) (int, error) {
	read, err := reader.ReadSeeker.Read(buffer)
	if err != io.EOF {
		reader.recorder.record(err)
	}
	return read, err
}
