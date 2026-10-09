import { formatDocumentSnippet } from './formatDocumentSnippet';

describe('formatDocumentSnippet', () => {
    it('wraps the snippet in ellipses and collapses whitespace', () => {
        expect(formatDocumentSnippet('  o contrato\n  foi assinado ')).toBe(
            '…o contrato foi assinado…'
        );
    });

    it.each([undefined, '', '   '])('returns empty text for %p', (emptySnippet) => {
        expect(formatDocumentSnippet(emptySnippet)).toBe('');
    });
});
