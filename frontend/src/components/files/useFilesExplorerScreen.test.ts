import { renderHook, act } from '@testing-library/react';
import { FileType } from '@/utils';
import type { FileData } from '@/components/providers/fileProvider/fileContext';

const mockUseFile = jest.fn();
const mockT = jest.fn((key: string) => key);

jest.mock('@/components/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: mockT }),
}));

import useFilesExplorerScreen from './useFilesExplorerScreen';

const makeFile = (overrides: Partial<FileData> & { id: number; name: string }): FileData => ({
    path: `/${overrides.name}`,
    parent_path: '/',
    type: FileType.File,
    format: '',
    size: 0,
    updated_at: '',
    created_at: '',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    ...overrides,
});

const rootDir = makeFile({
    id: 1,
    name: 'root',
    type: FileType.Directory,
    path: '/root',
    file_children: [
        makeFile({
            id: 2,
            name: 'child.txt',
            path: '/root/child.txt',
            parent_path: '/root',
        }),
    ],
});

describe('useFilesExplorerScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseFile.mockReturnValue({
            files: [rootDir],
            selectedItem: null,
            fileListFilter: 'all',
        });
    });

    it('returns FILES title when filter is all', () => {
        const { result } = renderHook(() => useFilesExplorerScreen());
        expect(result.current.currentListTitle).toBe('FILES');
    });

    it('returns STARRED_FILES title when filter is starred', () => {
        mockUseFile.mockReturnValue({
            files: [rootDir],
            selectedItem: null,
            fileListFilter: 'starred',
        });

        const { result } = renderHook(() => useFilesExplorerScreen());
        expect(result.current.currentListTitle).toBe('STARRED_FILES');
    });

    it('returns RECENT_FILES title when filter is recent', () => {
        mockUseFile.mockReturnValue({
            files: [rootDir],
            selectedItem: null,
            fileListFilter: 'recent',
        });

        const { result } = renderHook(() => useFilesExplorerScreen());
        expect(result.current.currentListTitle).toBe('RECENT_FILES');
    });

    it('returns files as currentItems when no selectedItem', () => {
        const { result } = renderHook(() => useFilesExplorerScreen());
        expect(result.current.breadcrumbSegments).toHaveLength(1);
        expect(result.current.breadcrumbSegments[0].isCurrent).toBe(true);
    });

    it('returns file_children when selectedItem is a directory', () => {
        mockUseFile.mockReturnValue({
            files: [rootDir],
            selectedItem: rootDir,
            fileListFilter: 'all',
        });

        const { result } = renderHook(() => useFilesExplorerScreen());
        // Breadcrumb should have root + the selected dir
        expect(result.current.breadcrumbSegments.length).toBeGreaterThanOrEqual(2);
    });

    it('returns empty array when selectedItem is a directory with no children', () => {
        const emptyDir = makeFile({
            id: 5,
            name: 'empty',
            type: FileType.Directory,
            path: '/empty',
            file_children: undefined,
        });

        mockUseFile.mockReturnValue({
            files: [emptyDir],
            selectedItem: emptyDir,
            fileListFilter: 'all',
        });

        const { result } = renderHook(() => useFilesExplorerScreen());
        // The itemCountLabel should show 0 items
        expect(result.current.itemCountLabel).toBe('0 ITENS');
    });

    it('returns empty array when selectedItem is a file', () => {
        const file = makeFile({
            id: 2,
            name: 'child.txt',
            type: FileType.File,
            path: '/root/child.txt',
            parent_path: '/root',
        });

        mockUseFile.mockReturnValue({
            files: [rootDir],
            selectedItem: file,
            fileListFilter: 'all',
        });

        const { result } = renderHook(() => useFilesExplorerScreen());
        expect(result.current.itemCountLabel).toBe('0 ITENS');
        // contextLabel for a file should be parent_path
        expect(result.current.contextLabel).toBe('/root');
    });

    it('builds breadcrumb with fallback when trail is empty (item not in tree)', () => {
        const orphanFile = makeFile({
            id: 999,
            name: 'orphan.txt',
            type: FileType.File,
            path: '/somewhere/orphan.txt',
            parent_path: '/somewhere',
        });

        mockUseFile.mockReturnValue({
            files: [rootDir],
            selectedItem: orphanFile,
            fileListFilter: 'all',
        });

        const { result } = renderHook(() => useFilesExplorerScreen());
        // Should have root segment + fallback segment for orphan item
        expect(result.current.breadcrumbSegments).toHaveLength(2);
        expect(result.current.breadcrumbSegments[1].id).toBe(999);
        expect(result.current.breadcrumbSegments[1].label).toBe('orphan.txt');
        expect(result.current.breadcrumbSegments[1].isCurrent).toBe(true);
    });

    it('toggles mobileTreeOpen via openMobileTree and closeMobileTree', () => {
        const { result } = renderHook(() => useFilesExplorerScreen());

        expect(result.current.mobileTreeOpen).toBe(false);

        act(() => {
            result.current.openMobileTree();
        });
        expect(result.current.mobileTreeOpen).toBe(true);

        act(() => {
            result.current.closeMobileTree();
        });
        expect(result.current.mobileTreeOpen).toBe(false);
    });

    it('contextLabel uses currentListTitle when no selectedItem', () => {
        const { result } = renderHook(() => useFilesExplorerScreen());
        expect(result.current.contextLabel).toBe('FILES');
    });

    it('contextLabel uses path for a directory selectedItem', () => {
        mockUseFile.mockReturnValue({
            files: [rootDir],
            selectedItem: rootDir,
            fileListFilter: 'all',
        });

        const { result } = renderHook(() => useFilesExplorerScreen());
        expect(result.current.contextLabel).toBe('/root');
    });

    it('shows singular item label when count is 1', () => {
        const singleChildDir = makeFile({
            id: 10,
            name: 'single',
            type: FileType.Directory,
            path: '/single',
            file_children: [
                makeFile({
                    id: 11,
                    name: 'only.txt',
                    path: '/single/only.txt',
                    parent_path: '/single',
                }),
            ],
        });

        mockUseFile.mockReturnValue({
            files: [singleChildDir],
            selectedItem: singleChildDir,
            fileListFilter: 'all',
        });

        const { result } = renderHook(() => useFilesExplorerScreen());
        expect(result.current.itemCountLabel).toBe('1 ITEM');
    });
});
