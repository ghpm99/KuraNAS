package doctext

import (
	"strconv"
	"strings"
)

const kerningGapThousandths = -200

type contentStreamScanner struct {
	content        []byte
	position       int
	collector      *textCollector
	pendingText    strings.Builder
	isInsideArray  bool
	hasPendingText bool
}

func collectContentStreamText(content []byte, collector *textCollector) {
	scanner := &contentStreamScanner{content: content, collector: collector}
	for scanner.position < len(scanner.content) && !collector.isFull {
		scanner.scanNextToken()
	}
}

func (scanner *contentStreamScanner) scanNextToken() {
	current := scanner.content[scanner.position]
	switch {
	case current == '(':
		scanner.appendPendingText(scanner.readLiteralString())
	case current == '<' && scanner.peekIs(1, '<'):
		scanner.position += 2
	case current == '<':
		scanner.skipHexString()
	case current == '[':
		scanner.isInsideArray = true
		scanner.position++
	case current == ']':
		scanner.isInsideArray = false
		scanner.position++
	case isPDFWhitespace(current) || current == '>' || current == '/':
		scanner.skipNameOrSpace()
	default:
		scanner.handleWord()
	}
}

func (scanner *contentStreamScanner) peekIs(offset int, expected byte) bool {
	index := scanner.position + offset
	return index < len(scanner.content) && scanner.content[index] == expected
}

func (scanner *contentStreamScanner) skipNameOrSpace() {
	if scanner.content[scanner.position] != '/' {
		scanner.position++
		return
	}
	scanner.position++
	for scanner.position < len(scanner.content) && !isPDFDelimiter(scanner.content[scanner.position]) {
		scanner.position++
	}
}

func (scanner *contentStreamScanner) skipHexString() {
	for scanner.position < len(scanner.content) && scanner.content[scanner.position] != '>' {
		scanner.position++
	}
	scanner.position++
}

func (scanner *contentStreamScanner) handleWord() {
	wordStart := scanner.position
	for scanner.position < len(scanner.content) && !isPDFDelimiter(scanner.content[scanner.position]) {
		scanner.position++
	}
	if scanner.position == wordStart {
		scanner.position++
		return
	}
	word := string(scanner.content[wordStart:scanner.position])

	if kerning, err := strconv.ParseFloat(word, 64); err == nil {
		if scanner.isInsideArray && kerning <= kerningGapThousandths {
			scanner.appendPendingText(" ")
		}
		return
	}
	scanner.applyOperator(word)
}

func (scanner *contentStreamScanner) applyOperator(operator string) {
	switch operator {
	case "Tj", "TJ":
		scanner.flushPendingText()
	case "'", "\"":
		scanner.collector.add("\n")
		scanner.flushPendingText()
	case "T*", "Td", "TD", "Tm":
		scanner.discardPendingText()
		scanner.collector.add(" ")
	case "ET":
		scanner.discardPendingText()
		scanner.collector.add("\n")
	default:
		scanner.discardPendingText()
	}
}

func (scanner *contentStreamScanner) appendPendingText(fragment string) {
	scanner.pendingText.WriteString(fragment)
	scanner.hasPendingText = true
}

func (scanner *contentStreamScanner) flushPendingText() {
	if scanner.hasPendingText {
		scanner.collector.add(scanner.pendingText.String())
	}
	scanner.discardPendingText()
}

func (scanner *contentStreamScanner) discardPendingText() {
	scanner.pendingText.Reset()
	scanner.hasPendingText = false
}

func (scanner *contentStreamScanner) readLiteralString() string {
	scanner.position++
	var decoded strings.Builder
	nestingDepth := 1
	for scanner.position < len(scanner.content) {
		current := scanner.content[scanner.position]
		scanner.position++
		switch current {
		case '\\':
			decoded.WriteString(scanner.readEscapeSequence())
		case '(':
			nestingDepth++
			decoded.WriteByte(current)
		case ')':
			nestingDepth--
			if nestingDepth == 0 {
				return decoded.String()
			}
			decoded.WriteByte(current)
		default:
			decoded.WriteRune(rune(current))
		}
	}
	return decoded.String()
}

func (scanner *contentStreamScanner) readEscapeSequence() string {
	if scanner.position >= len(scanner.content) {
		return ""
	}
	escaped := scanner.content[scanner.position]
	scanner.position++
	switch escaped {
	case 'n':
		return "\n"
	case 'r':
		return "\r"
	case 't':
		return "\t"
	case 'b', 'f':
		return ""
	case '\r':
		if scanner.peekIs(0, '\n') {
			scanner.position++
		}
		return ""
	case '\n':
		return ""
	}
	if escaped >= '0' && escaped <= '7' {
		return scanner.readOctalEscape(escaped)
	}
	return string(rune(escaped))
}

func (scanner *contentStreamScanner) readOctalEscape(firstDigit byte) string {
	characterCode := int(firstDigit - '0')
	for digitCount := 1; digitCount < 3 && scanner.position < len(scanner.content); digitCount++ {
		nextDigit := scanner.content[scanner.position]
		if nextDigit < '0' || nextDigit > '7' {
			break
		}
		characterCode = characterCode*8 + int(nextDigit-'0')
		scanner.position++
	}
	return string(rune(characterCode & 0xFF))
}

func isPDFWhitespace(character byte) bool {
	switch character {
	case ' ', '\t', '\r', '\n', '\f', 0:
		return true
	default:
		return false
	}
}

func isPDFDelimiter(character byte) bool {
	if isPDFWhitespace(character) {
		return true
	}
	switch character {
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}
