import type { DocumentIndexState } from '@/service/documents';

export type DocumentIndexNotice = 'protected' | 'scanned';

export const resolveDocumentIndexNotice = (
    indexState: DocumentIndexState | undefined
): DocumentIndexNotice | null => {
    if (indexState?.state !== 'name_only') return null;
    if (indexState.reason === 'encrypted') return 'protected';
    if (indexState.reason === 'no_text') return 'scanned';
    return null;
};
