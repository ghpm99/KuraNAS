export type DocumentViewerKind = 'pdf' | 'plainText' | 'markdown' | 'csv' | 'extractedText';

const markdownFormats = ['.md', '.markdown'];

export const resolveDocumentViewerKind = (format: string | undefined): DocumentViewerKind => {
    const normalizedFormat = (format ?? '').toLowerCase();
    if (normalizedFormat === '.pdf') return 'pdf';
    if (normalizedFormat === '.txt') return 'plainText';
    if (markdownFormats.includes(normalizedFormat)) return 'markdown';
    if (normalizedFormat === '.csv') return 'csv';
    return 'extractedText';
};
