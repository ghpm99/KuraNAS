import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

export const createTestFile = (id: number, overrides: Partial<FileData> = {}): FileData => ({
    id,
    name: `file-${id}.txt`,
    path: `/library/file-${id}.txt`,
    parent_path: '/library',
    type: 2,
    format: '.txt',
    size: 1024,
    updated_at: '2026-03-10T10:00:00Z',
    created_at: '2026-03-10T10:00:00Z',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    ...overrides,
});
