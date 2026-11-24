import { resolveDocumentViewerKind } from './documentViewerKind';

describe('resolveDocumentViewerKind', () => {
    it.each([
        ['.pdf', 'pdf'],
        ['.PDF', 'pdf'],
        ['.txt', 'plainText'],
        ['.md', 'markdown'],
        ['.markdown', 'markdown'],
        ['.csv', 'csv'],
        ['.docx', 'extractedText'],
        ['', 'extractedText'],
        [undefined, 'extractedText'],
    ])('%s -> %s', (format, expected) => {
        expect(resolveDocumentViewerKind(format)).toBe(expected);
    });
});
