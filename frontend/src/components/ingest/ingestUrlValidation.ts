export const ingestUrlMaxLength = 2048;

export type IngestUrlErrorKey =
    | 'INGEST_FORM_URL_REQUIRED'
    | 'INGEST_FORM_URL_TOO_LONG'
    | 'INGEST_FORM_URL_INVALID';

export const validateIngestUrl = (rawUrl: string): IngestUrlErrorKey | null => {
    const url = rawUrl.trim();
    if (url === '') return 'INGEST_FORM_URL_REQUIRED';
    if (url.length > ingestUrlMaxLength) return 'INGEST_FORM_URL_TOO_LONG';
    try {
        const { protocol, hostname } = new URL(url);
        const isWebScheme = protocol === 'http:' || protocol === 'https:';
        return isWebScheme && hostname !== '' ? null : 'INGEST_FORM_URL_INVALID';
    } catch {
        return 'INGEST_FORM_URL_INVALID';
    }
};
