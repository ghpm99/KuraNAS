import { act, renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { apiBase } from '@/service';
import { buildImageLibraryItem } from './imageLibraryTestFixtures';
import { useImageBulkActions } from './useImageBulkActions';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockEnqueueSnackbar = jest.fn();
jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock; delete: jest.Mock };

const images = [
    buildImageLibraryItem({ file_id: 11 }),
    buildImageLibraryItem({ file_id: 12 }),
    buildImageLibraryItem({ file_id: 13, starred: true }),
];

const buildWrapper = (queryClient: QueryClient) => {
    const Wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    return Wrapper;
};

describe('useImageBulkActions (seam)', () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const onSucceeded = jest.fn();

    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.post.mockResolvedValue({ data: { path: '/moved' } });
        mockedApi.delete.mockResolvedValue({ data: undefined });
    });

    const renderActions = () =>
        renderHook(() => useImageBulkActions(onSucceeded), { wrapper: buildWrapper(queryClient) });

    it('DELETEs /files/path once per image with the id in the body', async () => {
        const { result } = renderActions();

        await act(async () => {
            await result.current.deleteImages(images.slice(0, 2), false);
        });

        expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', { data: { id: 11 } });
        expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', { data: { id: 12 } });
        expect(onSucceeded).toHaveBeenCalledWith(images.slice(0, 2));
    });

    it('DELETEs /files/path with permanent=true as a query param', async () => {
        const { result } = renderActions();

        await act(async () => {
            await result.current.deleteImages([images[0]!], true);
        });

        expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', {
            data: { id: 11 },
            params: { permanent: true },
        });
    });

    it('POSTs /files/move with the source and destination for each image', async () => {
        const { result } = renderActions();

        await act(async () => {
            await result.current.moveImages(images.slice(0, 2), { folderId: 9, path: '/dst' });
        });

        expect(mockedApi.post).toHaveBeenCalledWith('/files/move', {
            source_id: 11,
            destination_folder_id: 9,
            destination_path: '/dst',
        });
        expect(mockedApi.post).toHaveBeenCalledWith('/files/move', {
            source_id: 12,
            destination_folder_id: 9,
            destination_path: '/dst',
        });
    });

    it('POSTs /files/starred/:id only for images that are not starred yet', async () => {
        const { result } = renderActions();

        await act(async () => {
            await result.current.toggleFavorites(images);
        });

        expect(mockedApi.post).toHaveBeenCalledTimes(2);
        expect(mockedApi.post).toHaveBeenCalledWith('/files/starred/11');
        expect(mockedApi.post).toHaveBeenCalledWith('/files/starred/12');
    });

    it('POSTs /files/starred/:id for every image when all are starred', async () => {
        const { result } = renderActions();
        const starredImages = images.map((image) => ({ ...image, starred: true }));

        await act(async () => {
            await result.current.toggleFavorites(starredImages);
        });

        expect(mockedApi.post).toHaveBeenCalledTimes(3);
    });

    it('keeps failed images selected and shows the first backend error verbatim', async () => {
        mockedApi.delete.mockImplementation((_url: string, config: { data: { id: number } }) =>
            config.data.id === 12
                ? Promise.reject({ response: { data: { error: 'sem permissão' } } })
                : Promise.resolve({ data: undefined })
        );
        const { result } = renderActions();

        await act(async () => {
            await result.current.deleteImages(images.slice(0, 2), false);
        });

        expect(onSucceeded).toHaveBeenCalledWith([images[0]]);
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('FILES_BULK_PARTIAL_SUMMARY', {
            variant: 'warning',
        });
    });

    it('shows the backend error verbatim when the only image fails', async () => {
        mockedApi.delete.mockRejectedValue({ response: { data: { error: 'sem permissão' } } });
        const { result } = renderActions();

        await act(async () => {
            await result.current.deleteImages([images[1]!], false);
        });

        expect(onSucceeded).toHaveBeenCalledWith([]);
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('sem permissão', { variant: 'error' });
    });

    it('invalidates the image library and the files queries after a mutation', async () => {
        const invalidateSpy = jest.spyOn(queryClient, 'invalidateQueries');
        const { result } = renderActions();

        await act(async () => {
            await result.current.deleteImages([images[0]!], false);
        });

        ['library', 'count', 'timeline', 'folders'].forEach((segment) =>
            expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['images', segment] })
        );
        expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['files'] });
    });
});
