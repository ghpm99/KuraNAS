import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import useFavoritesScreen from './useFavoritesScreen';

const mockUseFile = jest.fn();
const mockGetStarredCounts = jest.fn();

jest.mock('@/service/files', () => ({
    getStarredCounts: () => mockGetStarredCounts(),
}));

const renderFavoritesHook = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: ReactNode }) =>
        createElement(QueryClientProvider, { client: queryClient }, children);
    return renderHook(() => useFavoritesScreen(), { wrapper });
};

const rootFileContext = {
    starredKind: 'all',
    setStarredKind: jest.fn(),
    starredSort: 'name',
    setStarredSort: jest.fn(),
};

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
        mockGetStarredCounts.mockResolvedValue({ total: 10, folders: 3, files: 4, media: 3 });
    });

    it('renders without server counts and keeps chips countless', async () => {
        mockGetStarredCounts.mockRejectedValue(new Error('offline'));
        mockUseFile.mockReturnValue({ ...rootFileContext, files: [], selectedItem: null });

        const { result } = renderFavoritesHook();

        expect(result.current.isRootListing).toBe(true);
        expect(result.current.filterOptions.map((option) => option.count)).toEqual([
            null,
            null,
            null,
            null,
        ]);
        expect(result.current.itemCountLabel).toBe('0 ITENS');
    });

    it('takes chip counts and the item total from the counts endpoint', async () => {
        mockUseFile.mockReturnValue({
            ...rootFileContext,
            files: [{ id: 1, name: 'a', type: 2 }],
            selectedItem: null,
        });

        const { result } = renderFavoritesHook();

        await waitFor(() =>
            expect(result.current.filterOptions).toEqual([
                { value: 'all', label: 'FAVORITES_FILTER_ALL', count: 10 },
                { value: 'folders', label: 'FAVORITES_FILTER_FOLDERS', count: 3 },
                { value: 'files', label: 'FAVORITES_FILTER_FILES', count: 4 },
                { value: 'media', label: 'FAVORITES_FILTER_MEDIA', count: 3 },
            ])
        );
        expect(result.current.itemCountLabel).toBe('10 ITENS');
        expect(result.current.listedItems).toHaveLength(1);
    });

    it('delegates the chip to the server kind without filtering the listed items', async () => {
        const setStarredKind = jest.fn();
        mockUseFile.mockReturnValue({
            ...rootFileContext,
            setStarredKind,
            starredKind: 'folders',
            files: [
                { id: 1, name: 'a.pdf', type: 2, format: '.pdf' },
                { id: 2, name: 'dir', type: 1, format: '' },
            ],
            selectedItem: null,
        });

        const { result } = renderFavoritesHook();
        await waitFor(() => expect(result.current.filterOptions[1]?.count).toBe(3));

        expect(result.current.activeFilter).toBe('folders');
        expect(result.current.activeFilterLabel).toBe('FAVORITES_FILTER_FOLDERS');
        expect(result.current.itemCountLabel).toBe('3 ITENS');
        expect(result.current.listedItems).toHaveLength(2);

        act(() => {
            result.current.setActiveFilter('media');
        });
        expect(setStarredKind).toHaveBeenCalledWith('media');
    });

    it('exposes the sort options and delegates the sort change', () => {
        const setStarredSort = jest.fn();
        mockUseFile.mockReturnValue({
            ...rootFileContext,
            setStarredSort,
            starredSort: 'starred_at',
            files: [],
            selectedItem: null,
        });

        const { result } = renderFavoritesHook();

        expect(result.current.starredSort).toBe('starred_at');
        expect(result.current.sortOptions).toEqual([
            { value: 'name', label: 'FAVORITES_SORT_NAME' },
            { value: 'starred_at', label: 'FAVORITES_SORT_STARRED_AT' },
        ]);
        act(() => {
            result.current.setStarredSort('name');
        });
        expect(setStarredSort).toHaveBeenCalledWith('name');
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
            ...rootFileContext,
            files: [folder],
            selectedItem: folder,
            handleSelectItem: jest.fn(),
        });

        const { result } = renderFavoritesHook();

        expect(result.current.currentTitle).toBe('Captured');
        expect(result.current.contextPath).toBe('/captured');
        expect(result.current.breadcrumbSegments.map((segment) => segment.label)).toEqual([
            'STARRED_FILES',
            'Captured',
        ]);

        expect(result.current.isRootListing).toBe(false);
        expect(result.current.activeFilter).toBe('all');
        expect(result.current.listedItems.map((item) => item.name)).toEqual([
            'frame.jpg',
            'notes.txt',
        ]);
        expect(result.current.itemCountLabel).toBe('2 ITENS');
        expect(mockGetStarredCounts).not.toHaveBeenCalled();
    });

    it('shows an open-ended count and exposes pagination while more pages exist and counts are missing', () => {
        mockGetStarredCounts.mockRejectedValue(new Error('offline'));
        const fetchNextPage = jest.fn();
        mockUseFile.mockReturnValue({
            ...rootFileContext,
            files: [],
            selectedItem: null,
            hasNextPage: true,
            isFetchingNextPage: false,
            fetchNextPage,
        });

        const { result } = renderFavoritesHook();

        expect(result.current.itemCountLabel).toBe('FAVORITES_ITEMS_OPEN_ENDED');
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
            ...rootFileContext,
            files: [homeFolder],
            selectedItem: report,
            handleSelectItem,
        });

        const { result } = renderFavoritesHook();

        expect(result.current.breadcrumbSegments.map((segment) => segment.label)).toEqual([
            'STARRED_FILES',
            'docs',
            'sub',
            'report.txt',
        ]);
        expect(result.current.listedItems.map((entry) => entry.name)).toEqual([
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
            ...rootFileContext,
            files: [starredFile, otherStarredFile],
            selectedItem: starredFile,
            handleSelectItem,
        });

        const { result } = renderFavoritesHook();

        expect(result.current.breadcrumbSegments.map((segment) => segment.label)).toEqual([
            'STARRED_FILES',
            'a.txt',
        ]);
        expect(result.current.listedItems).toHaveLength(2);

        act(() => {
            result.current.closeSelectedItem();
        });
        expect(handleSelectItem).toHaveBeenCalledWith(null);
    });

    it('falls back to the selected item alone when it is missing from the loaded tree', () => {
        const orphan = buildEntry(77, 'orphan', '/orphan', { type: 1, file_children: [] });
        mockUseFile.mockReturnValue({
            ...rootFileContext,
            files: [],
            selectedItem: orphan,
            handleSelectItem: jest.fn(),
        });

        const { result } = renderFavoritesHook();

        expect(result.current.breadcrumbSegments.map((segment) => segment.label)).toEqual([
            'STARRED_FILES',
            'orphan',
        ]);
        expect(result.current.listedItems).toEqual([]);
    });
});
