import type { DocumentFilters, DocumentSort } from '@/service/documents';
import {
    defaultDocumentsSort,
    parseDocumentsFilters,
    parseDocumentsFolderId,
    parseDocumentsSort,
    toDocumentFilters,
    toDocumentSort,
} from './documentsFilters';
import { loadDocumentsSort } from './documentsSortPreference';

export type DocumentsListScope = {
    filters: DocumentFilters;
    sort: DocumentSort;
    searchQuery: string;
};

export const resolveDocumentsListScope = (listSearch: string): DocumentsListScope => {
    const listParams = new URLSearchParams(listSearch);
    const sortState = parseDocumentsSort(listParams) ?? loadDocumentsSort() ?? defaultDocumentsSort;
    return {
        filters: {
            ...toDocumentFilters(parseDocumentsFilters(listParams)),
            folderId: parseDocumentsFolderId(listParams),
        },
        sort: toDocumentSort(sortState),
        searchQuery: (listParams.get('q') ?? '').trim(),
    };
};
