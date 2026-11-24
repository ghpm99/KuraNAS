import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import type { DocumentListItem, DocumentSearchResult } from '@/service/documents';
import { FileType } from '@/utils';

const joinLogicalPath = (parentPath: string | undefined, name: string | undefined): string => {
    const folderPath = (parentPath ?? '').replace(/\/+$/, '');
    return `${folderPath}/${name ?? ''}`;
};

const emptyFileFields = {
    type: FileType.File,
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
};

export const documentToFileData = (document: DocumentListItem): FileData => ({
    ...emptyFileFields,
    id: document.id,
    name: document.name,
    path: joinLogicalPath(document.parent_path, document.name),
    parent_path: document.parent_path ?? '',
    format: document.format ?? '',
    size: document.size ?? 0,
    updated_at: document.modified_at ?? '',
    created_at: document.created_at ?? '',
    starred: Boolean(document.starred),
});

export const searchResultToFileData = (result: DocumentSearchResult): FileData => ({
    ...emptyFileFields,
    id: result.file_id,
    name: result.name,
    path: result.path || joinLogicalPath(result.parent_path, result.name),
    parent_path: result.parent_path ?? '',
    format: result.format ?? '',
    size: result.size ?? 0,
    updated_at: result.updated_at ?? '',
    created_at: '',
    starred: false,
});
