jest.mock('.', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
    },
}));

import { readFileSync } from 'fs';
import path from 'path';
import { apiBase } from '.';
import { getNextPageNumber } from '@/shared/pagination/getNextPageNumber';
import {
    countDocuments,
    getDocumentIndexState,
    getDocumentText,
    listDocuments,
    reindexDocument,
    searchDocuments,
} from './documents';
import type { DocumentFilters, DocumentListResponse } from './documents';

const mockedApiGet = apiBase.get as jest.Mock;
const mockedApiPost = apiBase.post as jest.Mock;

const fullFilters: DocumentFilters = {
    format: 'pdf',
    folderId: 4,
    modifiedFrom: '2024-03-01',
    modifiedTo: '2024-03-31',
    favoritesOnly: true,
    indexState: 'queued',
};

const fullFilterParams = {
    format: 'pdf',
    folder_id: 4,
    modified_from: '2024-03-01',
    modified_to: '2024-03-31',
    favorites_only: 'true',
    index_state: 'queued',
};

describe('service/documents', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('requests the document content search endpoint with the backend param names', async () => {
        const response = {
            items: [],
            pagination: { page: 2, page_size: 20, has_next: false, has_prev: true },
        };
        mockedApiGet.mockResolvedValue({ data: response });

        const result = await searchDocuments({ q: 'contrato', page: 2, pageSize: 20 });

        expect(mockedApiGet).toHaveBeenCalledWith('/documents/search', {
            params: { q: 'contrato', page: 2, page_size: 20 },
            signal: undefined,
        });
        expect(result).toEqual(response);
    });

    it('maps the camelCase filters to snake_case query params on search', async () => {
        mockedApiGet.mockResolvedValue({ data: { items: [], pagination: {} } });

        await searchDocuments({ q: 'iptu', filters: fullFilters, page: 1, pageSize: 20 });

        expect(mockedApiGet).toHaveBeenCalledWith('/documents/search', {
            params: { q: 'iptu', ...fullFilterParams, page: 1, page_size: 20 },
            signal: undefined,
        });
    });

    it('lists documents with filters, sort and pagination mapped once', async () => {
        const response = { items: [], pagination: { page: 3, page_size: 50 } };
        mockedApiGet.mockResolvedValue({ data: response });
        const signal = new AbortController().signal;

        const result = await listDocuments({
            filters: fullFilters,
            sort: { sort: 'size', order: 'asc' },
            page: 3,
            pageSize: 50,
            signal,
        });

        expect(mockedApiGet).toHaveBeenCalledWith('/documents', {
            params: { ...fullFilterParams, sort: 'size', order: 'asc', page: 3, page_size: 50 },
            signal,
        });
        expect(result).toEqual(response);
    });

    it('omits unset filters and sort from the list params', async () => {
        mockedApiGet.mockResolvedValue({ data: { items: [], pagination: {} } });

        await listDocuments({ filters: { favoritesOnly: false }, page: 1, pageSize: 50 });

        const [, requestConfig] = mockedApiGet.mock.calls[0];
        expect(requestConfig.params).toStrictEqual({ page: 1, page_size: 50 });
    });

    it('counts documents with the same filter mapping', async () => {
        mockedApiGet.mockResolvedValue({ data: { total: 1234 } });

        const result = await countDocuments(fullFilters);

        expect(mockedApiGet).toHaveBeenCalledWith('/documents/count', { params: fullFilterParams });
        expect(result).toEqual({ total: 1234 });
    });

    it('reads the index state of one document', async () => {
        const indexState = {
            file_id: 5,
            state: 'partial',
            indexed_bytes: 10,
            reason: '',
            error: '',
        };
        mockedApiGet.mockResolvedValue({ data: indexState });

        const result = await getDocumentIndexState(5);

        expect(mockedApiGet).toHaveBeenCalledWith('/documents/5/index-state');
        expect(result).toEqual(indexState);
    });

    it('queues the reindex of one document', async () => {
        mockedApiPost.mockResolvedValue({ data: { job_id: 77 } });

        const result = await reindexDocument(5);

        expect(mockedApiPost).toHaveBeenCalledWith('/documents/5/reindex');
        expect(result).toEqual({ job_id: 77 });
    });

    it('reads the extracted text of one document', async () => {
        mockedApiGet.mockResolvedValue({ data: { text: 'olá', is_partial: true } });

        const result = await getDocumentText(5);

        expect(mockedApiGet).toHaveBeenCalledWith('/documents/5/text');
        expect(result).toEqual({ text: 'olá', is_partial: true });
    });
});

describe('backend golden files of GET /documents', () => {
    const goldenDirectory = path.resolve(
        __dirname,
        '../../../backend/internal/api/v1/documenttext/testdata'
    );

    const readGoldenPage = async (fileName: string): Promise<DocumentListResponse> => {
        mockedApiGet.mockResolvedValue({
            data: JSON.parse(readFileSync(path.join(goldenDirectory, fileName), 'utf8')),
        });
        return listDocuments({ page: 1, pageSize: 2 });
    };

    it('feeds the service with the real page that has a next page', async () => {
        const page = await readGoldenPage('documents_page_with_next.json');

        expect(page.items).toHaveLength(2);
        expect(Object.keys(page.items[0] ?? {}).sort()).toEqual(
            [
                'created_at',
                'format',
                'id',
                'index_state',
                'modified_at',
                'name',
                'parent_path',
                'size',
                'starred',
            ].sort()
        );
        expect(page.items[0]).toMatchObject({ id: 1, starred: true, index_state: 'indexed' });
        expect(page.pagination).toMatchObject({ page: 1, page_size: 2, has_next: true });
        expect(getNextPageNumber(page)).toBe(2);
    });

    it('feeds the service with the real last page', async () => {
        const page = await readGoldenPage('documents_last_page.json');

        expect(page.items).toHaveLength(1);
        expect(page.pagination).toMatchObject({ page: 2, has_next: false });
        expect(getNextPageNumber(page)).toBeUndefined();
    });
});
