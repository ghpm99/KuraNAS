package doctext

import (
	"bytes"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	maxInflatedStreamBytes = 16 << 20
	pdfSignature           = "%PDF-"
	pdfSignatureWindow     = 1024
)

var errNotPDF = errors.New("file is not a pdf")

var unsupportedPDFFilters = [][]byte{
	[]byte("ASCII85Decode"), []byte("ASCIIHexDecode"), []byte("LZWDecode"), []byte("RunLengthDecode"),
	[]byte("DCTDecode"), []byte("CCITTFaxDecode"), []byte("JBIG2Decode"), []byte("JPXDecode"), []byte("Crypt"),
}

var nonContentStreamMarkers = [][]byte{
	[]byte("/ObjStm"), []byte("/XRef"), []byte("/Length1"), []byte("/EmbeddedFile"), []byte("/Metadata"),
}

func extractPDFText(ctx context.Context, path string) (Result, error) {
	pdfBytes, err := os.ReadFile(path)
	if err != nil {
		return Result{}, err
	}
	signatureWindow := pdfBytes[:min(len(pdfBytes), pdfSignatureWindow)]
	if !bytes.Contains(signatureWindow, []byte(pdfSignature)) {
		return Result{}, errNotPDF
	}
	if bytes.Contains(pdfBytes, []byte("/Encrypt")) {
		return Result{}, ErrEncrypted
	}

	collector := &textCollector{}
	searchFrom := 0
	for !collector.isFull {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		stream, nextSearchFrom, hasStream := nextPDFStream(pdfBytes, searchFrom)
		if !hasStream {
			break
		}
		searchFrom = nextSearchFrom
		if !stream.isTextContentCandidate() {
			continue
		}
		content, err := stream.decode()
		if err != nil || !bytes.Contains(content, []byte("BT")) {
			continue
		}
		collectContentStreamText(content, collector)
	}
	return finalizeText(collector.text(), collector.isFull), nil
}

type pdfStream struct {
	dictionary []byte
	data       []byte
}

func nextPDFStream(pdfBytes []byte, searchFrom int) (pdfStream, int, bool) {
	keywordOffset := bytes.Index(pdfBytes[searchFrom:], []byte("stream"))
	for keywordOffset >= 0 {
		keywordPosition := searchFrom + keywordOffset
		isEndKeyword := keywordPosition >= 3 && string(pdfBytes[keywordPosition-3:keywordPosition]) == "end"
		if !isEndKeyword {
			break
		}
		searchFrom = keywordPosition + len("stream")
		keywordOffset = bytes.Index(pdfBytes[searchFrom:], []byte("stream"))
	}
	if keywordOffset < 0 {
		return pdfStream{}, len(pdfBytes), false
	}

	keywordPosition := searchFrom + keywordOffset
	dataStart := skipStreamLineBreak(pdfBytes, keywordPosition+len("stream"))
	dataEndOffset := bytes.Index(pdfBytes[dataStart:], []byte("endstream"))
	if dataEndOffset < 0 {
		return pdfStream{}, len(pdfBytes), false
	}
	dataEnd := dataStart + dataEndOffset

	dictionaryStart := bytes.LastIndex(pdfBytes[searchFrom:keywordPosition], []byte("obj"))
	dictionary := pdfBytes[searchFrom:keywordPosition]
	if dictionaryStart >= 0 {
		dictionary = dictionary[dictionaryStart+len("obj"):]
	}
	return pdfStream{dictionary: dictionary, data: pdfBytes[dataStart:dataEnd]}, dataEnd + len("endstream"), true
}

func skipStreamLineBreak(pdfBytes []byte, position int) int {
	if position < len(pdfBytes) && pdfBytes[position] == '\r' {
		position++
	}
	if position < len(pdfBytes) && pdfBytes[position] == '\n' {
		position++
	}
	return position
}

func (stream pdfStream) isTextContentCandidate() bool {
	for _, marker := range nonContentStreamMarkers {
		if bytes.Contains(stream.dictionary, marker) {
			return false
		}
	}
	for _, filter := range unsupportedPDFFilters {
		if bytes.Contains(stream.dictionary, filter) {
			return false
		}
	}
	hasSubtype := bytes.Contains(stream.dictionary, []byte("/Subtype"))
	return !hasSubtype || bytes.Contains(stream.dictionary, []byte("/Form"))
}

func (stream pdfStream) decode() ([]byte, error) {
	if !bytes.Contains(stream.dictionary, []byte("/Flate")) && !bytes.Contains(stream.dictionary, []byte("/Fl")) {
		return stream.data, nil
	}
	inflater, err := zlib.NewReader(bytes.NewReader(stream.data))
	if err != nil {
		return nil, fmt.Errorf("open flate stream: %w", err)
	}
	defer inflater.Close()
	inflated, err := io.ReadAll(io.LimitReader(inflater, maxInflatedStreamBytes))
	if err != nil && len(inflated) == 0 {
		return nil, fmt.Errorf("inflate stream: %w", err)
	}
	return inflated, nil
}
