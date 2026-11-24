package doctext

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

const (
	odtContentEntry         = "content.xml"
	odtTextNamespace        = "urn:oasis:names:tc:opendocument:xmlns:text:1.0"
	maxOdtXMLBytes          = 64 << 20
	odtContextCheckInterval = 512
)

var errOdtContentMissing = errors.New("odt has no content.xml")

func extractOdtText(ctx context.Context, path string) (Result, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return Result{}, fmt.Errorf("open odt archive: %w", err)
	}
	defer archive.Close()

	for _, entry := range archive.File {
		if entry.Name != odtContentEntry {
			continue
		}
		entryReader, err := entry.Open()
		if err != nil {
			return Result{}, fmt.Errorf("open odt content: %w", err)
		}
		defer entryReader.Close()
		return collectOpenDocumentText(ctx, io.LimitReader(entryReader, maxOdtXMLBytes))
	}
	return Result{}, errOdtContentMissing
}

func collectOpenDocumentText(ctx context.Context, contentXML io.Reader) (Result, error) {
	collector := &textCollector{}
	decoder := xml.NewDecoder(contentXML)

	for tokenCount := 1; !collector.isFull; tokenCount++ {
		if tokenCount%odtContextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
		}

		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return Result{}, fmt.Errorf("parse odt content: %w", err)
		}

		switch element := token.(type) {
		case xml.StartElement:
			collector.add(openDocumentStartSeparator(element))
		case xml.EndElement:
			collector.add(openDocumentEndSeparator(element))
		case xml.CharData:
			collector.add(string(element))
		}
	}
	return finalizeText(collector.text(), collector.isFull), nil
}

func openDocumentStartSeparator(element xml.StartElement) string {
	if element.Name.Space != odtTextNamespace {
		return ""
	}
	switch element.Name.Local {
	case "tab":
		return " "
	case "line-break":
		return "\n"
	case "s":
		return " "
	}
	return ""
}

func openDocumentEndSeparator(element xml.EndElement) string {
	if element.Name.Space != odtTextNamespace {
		return ""
	}
	switch element.Name.Local {
	case "p", "h":
		return "\n"
	}
	return ""
}
