import { act, renderHook } from '@testing-library/react';
import useFavoritesScreen from './useFavoritesScreen';

const mockUseFile = jest.fn();

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

describe('useFavoritesScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('filters starred scope by folders, files, and media', () => {
        mockUseFile.mockReturnValue({
            files: [
                {
                    id: 1,
                    name: 'Projects',
                    path: '/projects',
                    parent_path: '/',
                    type: 1,
                    format: '',
                    size: 0,
                    updated_at: '',
                    created_at: '',
                    deleted_at: '',
                    last_interaction: '',
                    last_backup: '',
                    check_sum: '',
                    directory_content_count: 2,
                    starred: true,
                },
                {
                    id: 2,
                    name: 'Invoice.pdf',
                    path: '/docs/invoice.pdf',
                    parent_path: '/docs',
                    type: 2,
                    format: '.pdf',
                    size: 42,
                    updated_at: '',
                    created_at: '',
                    deleted_at: '',
                    last_interaction: '',
                    last_backup: '',
                    check_sum: '',
                    directory_content_count: 0,
                    starred: true,
                },
                {
                    id: 3,
                    name: 'Song.mp3',
                    path: '/music/song.mp3',
                    parent_path: '/music',
                    type: 2,
                    format: '.mp3',
                    size: 84,
                    updated_at: '',
                    created_at: '',
                    deleted_at: '',
                    last_interaction: '',
                    last_backup: '',
                    check_sum: '',
                    directory_content_count: 0,
                    starred: true,
                },
            ],
            selectedItem: null,
            handleSelectItem: jest.fn(),
        });

        const { result } = renderHook(() => useFavoritesScreen());

        expect(result.current.currentTitle).toBe('STARRED_FILES');
        expect(result.current.contextPath).toBe('FAVORITES_CONTEXT_ALL');
        expect(result.current.itemCountLabel).toBe('3 ITENS');
        expect(result.current.filterOptions).toEqual([
            { value: 'all', label: 'FAVORITES_FILTER_ALL', count: 3 },
            { value: 'folders', label: 'FAVORITES_FILTER_FOLDERS', count: 1 },
            { value: 'files', label: 'FAVORITES_FILTER_FILES', count: 1 },
            { value: 'media', label: 'FAVORITES_FILTER_MEDIA', count: 1 },
        ]);

        act(() => {
            result.current.setActiveFilter('folders');
        });
        expect(result.current.filteredItems.map((item) => item.name)).toEqual(['Projects']);
        expect(result.current.activeFilterLabel).toBe('FAVORITES_FILTER_FOLDERS');

        act(() => {
            result.current.setActiveFilter('files');
        });
        expect(result.current.filteredItems.map((item) => item.name)).toEqual(['Invoice.pdf']);

        act(() => {
            result.current.setActiveFilter('media');
        });
        expect(result.current.filteredItems.map((item) => item.name)).toEqual(['Song.mp3']);
    });

    it('uses the selected folder as the active scope and keeps breadcrumb trail', () => {
        const folder = {
            id: 10,
            name: 'Captured',
            path: '/captured',
            parent_path: '/',
            type: 1,
            format: '',
            size: 0,
            updated_at: '',
            created_at: '',
            deleted_at: '',
            last_interaction: '',
            last_backup: '',
            check_sum: '',
            directory_content_count: 2,
            starred: true,
            file_children: [
                {
                    id: 11,
                    name: 'frame.jpg',
                    path: '/captured/frame.jpg',
                    parent_path: '/captured',
                    type: 2,
                    format: '.jpg',
                    size: 30,
                    updated_at: '',
                    created_at: '',
                    deleted_at: '',
                    last_interaction: '',
                    last_backup: '',
                    check_sum: '',
                    directory_content_count: 0,
                    starred: true,
                },
                {
                    id: 12,
                    name: 'notes.txt',
                    path: '/captured/notes.txt',
                    parent_path: '/captured',
                    type: 2,
                    format: '.txt',
                    size: 12,
                    updated_at: '',
                    created_at: '',
                    deleted_at: '',
                    last_interaction: '',
                    last_backup: '',
                    check_sum: '',
                    directory_content_count: 0,
                    starred: true,
                },
            ],
        };

        mockUseFile.mockReturnValue({
            files: [folder],
            selectedItem: folder,
            handleSelectItem: jest.fn(),
        });

        const { result } = renderHook(() => useFavoritesScreen());

        expect(result.current.currentTitle).toBe('Captured');
        expect(result.current.contextPath).toBe('/captured');
        expect(result.current.breadcrumbSegments.map((segment) => segment.label)).toEqual([
            'STARRED_FILES',
            'Captured',
        ]);

        act(() => {
            result.current.setActiveFilter('media');
        });

        expect(result.current.filteredItems.map((item) => item.name)).toEqual(['frame.jpg']);
        expect(result.current.itemCountLabel).toBe('1 ITEM');
    });

    it('shows an open-ended count and exposes pagination while more pages exist', () => {
        const fetchNextPage = jest.fn();
        mockUseFile.mockReturnValue({
            files: [],
            selectedItem: null,
            hasNextPage: true,
            isFetchingNextPage: false,
            fetchNextPage,
        });

        const { result } = renderHook(() => useFavoritesScreen());

        expect(result.current.itemCountLabel).toBe('0+ ITENS');
        expect(result.current.hasNextPage).toBe(true);
        expect(result.current.fetchNextPage).toBe(fetchNextPage);
    });

    const buildEntry = (id: number, name: string, path: string, overrides = {}) => ({
        id,
        name,
        path,
        parent_path: path.slice(0, path.lastIndexOf('/')) || '/',
        type: 2,
        format: '.txt',
        size: 1,
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

    it('starts the breadcrumb at the starred ancestor and lists sibling files when a file is selected', () => {
        const handleSelectItem = jest.fn();
        const report = buildEntry(30, 'report.txt', '/home/docs/sub/report.txt');
        const notes = buildEntry(31, 'notes.txt', '/home/docs/sub/notes.txt');
        const subFolder = buildEntry(20, 'sub', '/home/docs/sub', {
            type: 1,
            file_children: [report, notes],
        });
        const starredFolder = buildEntry(10, 'docs', '/home/docs', {
            type: 1,
            starred: true,
            file_children: [subFolder],
        });
        const homeFolder = buildEntry(1, 'home', '/home', {
            type: 1,
            file_children: [starredFolder],
        });
        mockUseFile.mockReturnValue({
            files: [homeFolder],
            selectedItem: report,
            handleSelectItem,
        });

        const { result } = renderHook(() => useFavoritesScreen());

        expect(result.current.breadcrumbSegments.map((segment) => segment.label)).toEqual([
            'STARRED_FILES',
            'docs',
            'sub',
            'report.txt',
        ]);
        expect(result.current.filteredItems.map((entry) => entry.name)).toEqual([
            'report.txt',
            'notes.txt',
        ]);

        act(() => {
            result.current.closeSelectedItem();
        });
        expect(handleSelectItem).toHaveBeenCalledWith(subFolder);
    });

    it('lists the starred roots and closes back to the root when a starred root file is selected', () => {
        const handleSelectItem = jest.fn();
        const starredFile = buildEntry(5, 'a.txt', '/a.txt', { starred: true });
        const otherStarredFile = buildEntry(6, 'b.txt', '/b.txt', { starred: true });
        mockUseFile.mockReturnValue({
            files: [starredFile, otherStarredFile],
            selectedItem: starredFile,
            handleSelectItem,
        });

        const { result } = renderHook(() => useFavoritesScreen());

        expect(result.current.breadcrumbSegments.map((segment) => segment.label)).toEqual([
            'STARRED_FILES',
            'a.txt',
        ]);
        expect(result.current.filteredItems).toHaveLength(2);

        act(() => {
            result.current.closeSelectedItem();
        });
        expect(handleSelectItem).toHaveBeenCalledWith(null);
    });

    it('falls back to the selected item alone when it is missing from the loaded tree', () => {
        const orphan = buildEntry(77, 'orphan', '/orphan', { type: 1, file_children: [] });
        mockUseFile.mockReturnValue({
            files: [],
            selectedItem: orphan,
            handleSelectItem: jest.fn(),
        });

        const { result } = renderHook(() => useFavoritesScreen());

        expect(result.current.breadcrumbSegments.map((segment) => segment.label)).toEqual([
            'STARRED_FILES',
            'orphan',
        ]);
        expect(result.current.filteredItems).toEqual([]);
    });
});
