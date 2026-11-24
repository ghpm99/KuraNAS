import type { DocumentFilters, DocumentSort } from '@/service/documents';
import { documentsQueryKey } from '@/shared/queryKeys/fileQueryKeys';

export const documentsQueryKeyPrefix = [documentsQueryKey] as const;

export const documentsListQueryKey = (filters: DocumentFilters, sort: DocumentSort) =>
    [...documentsQueryKeyPrefix, 'list', filters, sort] as const;

export const documentsCountQueryKey = (filters: DocumentFilters) =>
    [...documentsQueryKeyPrefix, 'count', filters] as const;

export const documentsSearchQueryKey = (query: string, filters: DocumentFilters) =>
    [...documentsQueryKeyPrefix, 'search', query, filters] as const;

export const documentIndexStateQueryKey = (fileId: number) =>
    [...documentsQueryKeyPrefix, 'index-state', fileId] as const;

export const documentTextQueryKey = (fileId: number) =>
    [...documentsQueryKeyPrefix, 'text', fileId] as const;

export const documentViewerFileQueryKey = (fileId: number) =>
    [...documentsQueryKeyPrefix, 'viewer-file', fileId] as const;
