import type { DocumentListItem, DocumentSearchResult } from '@/service/documents';
import { FileType } from '@/utils';
import { documentToFileData, searchResultToFileData } from './documentFileAdapter';

const listItem: DocumentListItem = {
    id: 4,
    name: 'escritura.pdf',
    format: '.pdf',
    parent_path: '/Casa',
    size: 2048,
    modified_at: '2024-03-05T10:00:00Z',
    created_at: '2024-03-01T10:00:00Z',
    starred: true,
    index_state: 'indexed',
};

describe('documentFileAdapter', () => {
    it('builds the logical path from the parent path and the name', () => {
        const file = documentToFileData(listItem);

        expect(file).toMatchObject({
            id: 4,
            name: 'escritura.pdf',
            path: '/Casa/escritura.pdf',
            parent_path: '/Casa',
            type: FileType.File,
            format: '.pdf',
            size: 2048,
            updated_at: '2024-03-05T10:00:00Z',
            created_at: '2024-03-01T10:00:00Z',
            starred: true,
        });
    });

    it.each(['/', ''])('handles a document at the root with parent path %j', (parentPath) => {
        expect(documentToFileData({ ...listItem, parent_path: parentPath }).path).toBe(
            '/escritura.pdf'
        );
    });

    it('survives a partial item', () => {
        const file = documentToFileData({ id: 9 } as DocumentListItem);

        expect(file).toMatchObject({ id: 9, type: FileType.File, starred: false, size: 0 });
    });

    it('adapts a search result keeping its logical path', () => {
        const result: DocumentSearchResult = {
            file_id: 8,
            name: 'iptu.docx',
            path: '/Casa/iptu.docx',
            parent_path: '/Casa',
            format: '.docx',
            size: 10,
            updated_at: '2024-03-05T10:00:00Z',
            snippet: '',
            page: null,
        };

        expect(searchResultToFileData(result)).toMatchObject({
            id: 8,
            path: '/Casa/iptu.docx',
            parent_path: '/Casa',
            type: FileType.File,
            starred: false,
        });
    });

    it('falls back to the parent path when a search result has no path', () => {
        expect(
            searchResultToFileData({
                file_id: 1,
                name: 'a.txt',
                parent_path: '/x',
            } as DocumentSearchResult).path
        ).toBe('/x/a.txt');
    });
});
