import { QueryClient, QueryClientProvider, type InfiniteData } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { apiBase } from '@/service';
import type { ImageLibraryFilters, ImageLibraryPage } from '@/types/imageLibrary';
import { buildImageLibraryItem } from './imageLibraryTestFixtures';
import { useImageStarToggle } from './useImageStarToggle';

const enqueueSnackbar = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { post: jest.fn() },
}));

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApiPost = apiBase.post as jest.Mock;

const emptyFilters: ImageLibraryFilters = {
    nameQuery: '',
    categories: [],
    isStarredOnly: false,
    formats: [],
    takenFrom: '',
    takenTo: '',
    folder: '',
};
const ordering = { sort: 'taken_at', order: 'desc' };

const buildCache = (starredIds: number[]): InfiniteData<ImageLibraryPage, unknown> => ({
    pageParams: [{ page: 1 }],
    pages: [
        {
            items: [1, 2].map((fileId) =>
                buildImageLibraryItem({ file_id: fileId, starred: starredIds.includes(fileId) })
            ),
            next_cursor: '',
            has_next: false,
            page_size: 60,
        },
    ],
});

const libraryKey = (filters: ImageLibraryFilters) => ['images', 'library', filters, ordering, ''];

const setup = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    queryClient.setQueryData(libraryKey(emptyFilters), buildCache([2]));
    queryClient.setQueryData(
        libraryKey({ ...emptyFilters, isStarredOnly: true }),
        buildCache([1, 2])
    );
    const wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    const hook = renderHook(() => useImageStarToggle(), { wrapper });
    const readItems = (filters: ImageLibraryFilters) =>
        queryClient
            .getQueryData<InfiniteData<ImageLibraryPage>>(libraryKey(filters))!
            .pages.flatMap((page) => page.items);
    return { queryClient, hook, readItems };
};

describe('useImageStarToggle', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('stars optimistically across every cached gallery page before the server answers', async () => {
        let resolveRequest: () => void = () => undefined;
        mockedApiPost.mockReturnValue(new Promise<void>((resolve) => (resolveRequest = resolve)));
        const { hook, readItems } = setup();

        act(() => hook.result.current.toggleStar(1, false));

        await waitFor(() =>
            expect(readItems(emptyFilters).find((item) => item.file_id === 1)?.starred).toBe(true)
        );
        expect(mockedApiPost).toHaveBeenCalledWith('/files/starred/1');
        expect(hook.result.current.isStarTogglePending).toBe(true);

        await act(async () => resolveRequest());
        await waitFor(() =>
            expect(enqueueSnackbar).toHaveBeenCalledWith('IMAGES_VIEWER_FAVORITE_ADDED', {
                variant: 'success',
            })
        );
    });

    it('drops the image from starred-only caches when it is unstarred', async () => {
        mockedApiPost.mockResolvedValue({});
        const { hook, readItems } = setup();

        act(() => hook.result.current.toggleStar(2, true));

        await waitFor(() =>
            expect(
                readItems({ ...emptyFilters, isStarredOnly: true }).map((item) => item.file_id)
            ).toEqual([1])
        );
        expect(readItems(emptyFilters).find((item) => item.file_id === 2)?.starred).toBe(false);
        await waitFor(() =>
            expect(enqueueSnackbar).toHaveBeenCalledWith('IMAGES_VIEWER_FAVORITE_REMOVED', {
                variant: 'success',
            })
        );
    });

    it('rolls the cache back and reports the failure when the request fails', async () => {
        mockedApiPost.mockRejectedValue(new Error('boom'));
        const { hook, readItems } = setup();

        act(() => hook.result.current.toggleStar(1, false));

        await waitFor(() =>
            expect(enqueueSnackbar).toHaveBeenCalledWith('IMAGES_VIEWER_FAVORITE_ERROR', {
                variant: 'error',
            })
        );
        expect(readItems(emptyFilters).find((item) => item.file_id === 1)?.starred).toBe(false);
    });
});
