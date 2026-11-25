import { ingestUrlMaxLength, validateIngestUrl } from './ingestUrlValidation';

describe('validateIngestUrl', () => {
    it.each(['http://a.com/x', 'https://a.com/x?y=1'])('accepts %s', (url) => {
        expect(validateIngestUrl(url)).toBeNull();
    });

    it.each(['ftp://a.com', 'javascript:alert(1)', 'not a url', 'https://'])(
        'rejects %s',
        (url) => {
            expect(validateIngestUrl(url)).toBe('INGEST_FORM_URL_INVALID');
        }
    );

    it('requires a value', () => {
        expect(validateIngestUrl('   ')).toBe('INGEST_FORM_URL_REQUIRED');
    });

    it('rejects urls over the limit and accepts the exact limit', () => {
        const prefix = 'https://a.com/';
        expect(validateIngestUrl(prefix + 'a'.repeat(ingestUrlMaxLength))).toBe(
            'INGEST_FORM_URL_TOO_LONG'
        );
        expect(
            validateIngestUrl(prefix + 'a'.repeat(ingestUrlMaxLength - prefix.length))
        ).toBeNull();
    });
});
