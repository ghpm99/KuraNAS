package doctext

import (
	"bytes"
	"context"
	"regexp"
	"strconv"
)

const (
	maxPageTreeDepth = 64
	pageBreak        = "\f"
)

var (
	objectHeaderPattern   = regexp.MustCompile(`(\d+)\s+\d+\s+obj\b`)
	referencePattern      = regexp.MustCompile(`(\d+)\s+\d+\s+R\b`)
	catalogTypePattern    = regexp.MustCompile(`/Type\s*/Catalog\b`)
	objectStreamPattern   = regexp.MustCompile(`/Type\s*/ObjStm\b`)
	pageTypePattern       = regexp.MustCompile(`/Type\s*/Page\b`)
	pagesReferencePattern = regexp.MustCompile(`/Pages\s+(\d+)\s+\d+\s+R\b`)
	kidsArrayPattern      = regexp.MustCompile(`/Kids\s*\[([^\]]*)\]`)
	contentsPattern       = regexp.MustCompile(`/Contents\s*(\[[^\]]*\]|\d+\s+\d+\s+R\b)`)
	objectCountPattern    = regexp.MustCompile(`/N\s+(\d+)`)
	firstOffsetPattern    = regexp.MustCompile(`/First\s+(\d+)`)
)

type pdfObject struct {
	dictionary []byte
	stream     pdfStream
	hasStream  bool
}

type pdfObjectIndex struct {
	objects       map[int]pdfObject
	catalogNumber int
}

func indexPDFObjects(pdfBytes []byte) pdfObjectIndex {
	index := pdfObjectIndex{objects: map[int]pdfObject{}}
	cursor := 0
	for cursor < len(pdfBytes) {
		header := objectHeaderPattern.FindSubmatchIndex(pdfBytes[cursor:])
		if header == nil {
			break
		}
		objectNumber, _ := strconv.Atoi(string(pdfBytes[cursor+header[2] : cursor+header[3]]))
		bodyStart := cursor + header[1]
		bodyLength := bytes.Index(pdfBytes[bodyStart:], []byte("endobj"))
		if bodyLength < 0 {
			break
		}
		index.add(objectNumber, pdfBytes[bodyStart:bodyStart+bodyLength])
		cursor = bodyStart + bodyLength
	}
	return index
}

func (index *pdfObjectIndex) add(objectNumber int, body []byte) {
	object := pdfObject{dictionary: body}
	if stream, _, hasStream := nextPDFStream(body, 0); hasStream {
		object = pdfObject{dictionary: stream.dictionary, stream: stream, hasStream: true}
	}
	index.objects[objectNumber] = object
	if catalogTypePattern.Match(object.dictionary) {
		index.catalogNumber = objectNumber
	}
	if object.hasStream && objectStreamPattern.Match(object.dictionary) {
		index.addObjectStreamMembers(object.stream)
	}
}

func (index *pdfObjectIndex) addObjectStreamMembers(stream pdfStream) {
	memberCount, hasCount := numberAfter(objectCountPattern, stream.dictionary)
	firstOffset, hasFirst := numberAfter(firstOffsetPattern, stream.dictionary)
	content, err := stream.decode()
	if err != nil || !hasCount || !hasFirst || firstOffset > len(content) {
		return
	}
	headerFields := bytes.Fields(content[:firstOffset])
	for member := 0; member < memberCount && member*2+1 < len(headerFields); member++ {
		objectNumber, numberErr := strconv.Atoi(string(headerFields[member*2]))
		memberStart, offsetErr := strconv.Atoi(string(headerFields[member*2+1]))
		if numberErr != nil || offsetErr != nil {
			continue
		}
		memberEnd := len(content)
		if member*2+3 < len(headerFields) {
			if nextStart, nextErr := strconv.Atoi(string(headerFields[member*2+3])); nextErr == nil {
				memberEnd = min(firstOffset+nextStart, len(content))
			}
		}
		if firstOffset+memberStart > memberEnd {
			continue
		}
		memberBody := content[firstOffset+memberStart : memberEnd]
		index.objects[objectNumber] = pdfObject{dictionary: memberBody}
		if catalogTypePattern.Match(memberBody) {
			index.catalogNumber = objectNumber
		}
	}
}

func numberAfter(pattern *regexp.Regexp, source []byte) (int, bool) {
	match := pattern.FindSubmatch(source)
	if match == nil {
		return 0, false
	}
	number, err := strconv.Atoi(string(match[1]))
	return number, err == nil
}

type pdfPageTree struct {
	index          pdfObjectIndex
	visitedNodes   map[int]bool
	pageContentIDs [][]int
}

func resolvePDFPages(index pdfObjectIndex) ([][]int, bool) {
	catalog, hasCatalog := index.objects[index.catalogNumber]
	if !hasCatalog {
		return nil, false
	}
	rootNumber, hasRoot := numberAfter(pagesReferencePattern, catalog.dictionary)
	if !hasRoot {
		return nil, false
	}
	tree := &pdfPageTree{index: index, visitedNodes: map[int]bool{}}
	if !tree.collectPages(rootNumber, 0) || len(tree.pageContentIDs) == 0 {
		return nil, false
	}
	return tree.pageContentIDs, true
}

func (tree *pdfPageTree) collectPages(nodeNumber int, depth int) bool {
	node, isKnown := tree.index.objects[nodeNumber]
	if !isKnown || tree.visitedNodes[nodeNumber] || depth > maxPageTreeDepth {
		return false
	}
	tree.visitedNodes[nodeNumber] = true

	if kids := kidsArrayPattern.FindSubmatch(node.dictionary); kids != nil {
		for _, child := range referencePattern.FindAllSubmatch(kids[1], -1) {
			childNumber, _ := strconv.Atoi(string(child[1]))
			if !tree.collectPages(childNumber, depth+1) {
				return false
			}
		}
		return true
	}
	if !pageTypePattern.Match(node.dictionary) {
		return false
	}
	tree.pageContentIDs = append(tree.pageContentIDs, tree.contentObjectNumbers(node.dictionary))
	return true
}

func (tree *pdfPageTree) contentObjectNumbers(pageDictionary []byte) []int {
	contents := contentsPattern.FindSubmatch(pageDictionary)
	if contents == nil {
		return nil
	}
	if contents[1][0] == '[' {
		return referencedObjectNumbers(contents[1])
	}
	contentsNumber, _ := numberAfter(referencePattern, contents[1])
	referenced, isKnown := tree.index.objects[contentsNumber]
	if isKnown && !referenced.hasStream && bytes.HasPrefix(bytes.TrimSpace(referenced.dictionary), []byte("[")) {
		return referencedObjectNumbers(referenced.dictionary)
	}
	return []int{contentsNumber}
}

func referencedObjectNumbers(arraySource []byte) []int {
	numbers := []int{}
	for _, reference := range referencePattern.FindAllSubmatch(arraySource, -1) {
		number, _ := strconv.Atoi(string(reference[1]))
		numbers = append(numbers, number)
	}
	return numbers
}

func collectPagedPDFText(ctx context.Context, pdfBytes []byte) (*textCollector, bool, error) {
	index := indexPDFObjects(pdfBytes)
	pages, isResolved := resolvePDFPages(index)
	if !isResolved {
		return nil, false, nil
	}
	collector := &textCollector{}
	for pageIndex, contentNumbers := range pages {
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}
		if collector.isFull {
			break
		}
		if pageIndex > 0 {
			collector.add(pageBreak)
		}
		for _, contentNumber := range contentNumbers {
			if content, hasContent := index.objects[contentNumber]; hasContent && content.hasStream {
				collectStreamText(content.stream, collector)
			}
		}
	}
	return collector, true, nil
}
