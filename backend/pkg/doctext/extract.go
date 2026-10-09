package doctext

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"

	"nas-go/api/pkg/utils"
)

const (
	MaxFileSizeBytes   = 50 << 20
	MaxReadBytes       = 2 << 20
	MaxStoredTextBytes = 200 << 10
)

const (
	ErrorCodeTooLarge          = "too_large"
	ErrorCodeBinary            = "binary"
	ErrorCodeUnsupportedFormat = "unsupported_format"
	ErrorCodeEncrypted         = "encrypted"
	ErrorCodeTimeout           = "timeout"
	ErrorCodeUnreadable        = "unreadable"
	ErrorCodeExtractFailed     = "extract_failed"
)

var (
	ErrTooLarge          = errors.New(ErrorCodeTooLarge)
	ErrBinary            = errors.New(ErrorCodeBinary)
	ErrUnsupportedFormat = errors.New(ErrorCodeUnsupportedFormat)
	ErrEncrypted         = errors.New(ErrorCodeEncrypted)
)

type Result struct {
	Text      string
	Truncated bool
}

// Extract reads the file at path and returns its searchable text, at most
// MaxStoredTextBytes long. format is the lowercase extension with the leading
// dot. It fails with ErrTooLarge above MaxFileSizeBytes, ErrBinary for binary
// content, ErrUnsupportedFormat for formats it does not know and ErrEncrypted
// for protected PDFs. Cancelling ctx aborts long extractions.
func Extract(ctx context.Context, path string, format string) (Result, error) {
	normalizedFormat := utils.NormalizeExtension(format)
	if !utils.IsDocumentTextFormat(normalizedFormat) {
		return Result{}, ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return Result{}, fmt.Errorf("stat document: %w", err)
	}
	if fileInfo.Size() > MaxFileSizeBytes {
		return Result{}, ErrTooLarge
	}

	switch {
	case normalizedFormat == ".pdf":
		return extractPDFText(ctx, path)
	case normalizedFormat == ".docx":
		return extractDocxText(ctx, path)
	case slices.Contains(utils.MarkupTextFormats, normalizedFormat):
		return extractMarkupText(path, fileInfo.Size())
	default:
		return extractPlainText(path, fileInfo.Size())
	}
}

// ErrorCode maps an extraction error to the short code stored in
// document_text.error.
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrTooLarge):
		return ErrorCodeTooLarge
	case errors.Is(err, ErrBinary):
		return ErrorCodeBinary
	case errors.Is(err, ErrUnsupportedFormat):
		return ErrorCodeUnsupportedFormat
	case errors.Is(err, ErrEncrypted):
		return ErrorCodeEncrypted
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return ErrorCodeTimeout
	case errors.Is(err, fs.ErrNotExist), errors.Is(err, fs.ErrPermission):
		return ErrorCodeUnreadable
	default:
		return ErrorCodeExtractFailed
	}
}
