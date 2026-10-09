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
	docxMainDocumentEntry     = "word/document.xml"
	wordprocessingMLNamespace = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	maxDocxXMLBytes           = 64 << 20
	docxContextCheckInterval  = 512
)

var errDocxMainDocumentMissing = errors.New("docx has no word/document.xml")

func extractDocxText(ctx context.Context, path string) (Result, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return Result{}, fmt.Errorf("open docx archive: %w", err)
	}
	defer archive.Close()

	for _, entry := range archive.File {
		if entry.Name != docxMainDocumentEntry {
			continue
		}
		entryReader, err := entry.Open()
		if err != nil {
			return Result{}, fmt.Errorf("open docx main document: %w", err)
		}
		defer entryReader.Close()
		return collectWordprocessingText(ctx, io.LimitReader(entryReader, maxDocxXMLBytes))
	}
	return Result{}, errDocxMainDocumentMissing
}

func collectWordprocessingText(ctx context.Context, documentXML io.Reader) (Result, error) {
	collector := &textCollector{}
	decoder := xml.NewDecoder(documentXML)
	isInsideText := false

	for tokenCount := 1; !collector.isFull; tokenCount++ {
		if tokenCount%docxContextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
		}

		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return Result{}, fmt.Errorf("parse docx main document: %w", err)
		}

		switch element := token.(type) {
		case xml.StartElement:
			isInsideText = handleWordprocessingStart(collector, element, isInsideText)
		case xml.EndElement:
			isInsideText = handleWordprocessingEnd(collector, element, isInsideText)
		case xml.CharData:
			if isInsideText {
				collector.add(string(element))
			}
		}
	}
	return finalizeText(collector.text(), collector.isFull), nil
}

func handleWordprocessingStart(collector *textCollector, element xml.StartElement, isInsideText bool) bool {
	if element.Name.Space != wordprocessingMLNamespace {
		return isInsideText
	}
	switch element.Name.Local {
	case "t":
		return true
	case "tab":
		collector.add(" ")
	case "br", "cr":
		collector.add("\n")
	}
	return isInsideText
}

func handleWordprocessingEnd(collector *textCollector, element xml.EndElement, isInsideText bool) bool {
	if element.Name.Space != wordprocessingMLNamespace {
		return isInsideText
	}
	switch element.Name.Local {
	case "t":
		return false
	case "p":
		collector.add("\n")
	}
	return isInsideText
}
