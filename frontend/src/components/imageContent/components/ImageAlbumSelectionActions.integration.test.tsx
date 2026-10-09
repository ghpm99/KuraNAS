import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, renderHook, screen, waitFor, within } from '@testing-library/react';
import { apiBase } from '@/service';
import { buildImageLibraryItem } from '../imageLibraryTestFixtures';
import { useImageSelection, type ImageSelection } from '../useImageSelection';
import ImageSelectionToolbar from './ImageSelectionToolbar';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params?.name ? `${key}:${params.name}` : key,
    }),
}));

const mockEnqueueSnackbar = jest.fn();
jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

const mockedApi = apiBase as unknown as Record<'get' | 'post' | 'put' | 'delete', jest.Mock>;

const trip = {
    id: 4,
    name: 'Trip',
    cover_file_id: null,
    item_count: 3,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
};
const holiday = { ...trip, id: 6, name: 'Holiday' };

const buildSelection = (fileIds: number[]): ImageSelection => {
    const { result } = renderHook(() => useImageSelection('scope'));
    const selectedItems = fileIds.map((fileId) => buildImageLibraryItem({ file_id: fileId }));
    return { ...result.current, selectedItems, selectedCount: selectedItems.length };
};

const renderToolbar = (fileIds: number[], userAlbumId?: number) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <ImageSelectionToolbar
                selection={buildSelection(fileIds)}
                loadedImages={[]}
                userAlbumId={userAlbumId}
            />
        </QueryClientProvider>
    );

describe('album actions of the selection toolbar (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({
            data: {
                items: [trip, holiday],
                pagination: { page: 1, page_size: 60, has_next: false },
            },
        });
        mockedApi.post.mockResolvedValue({ data: { requested: 2, changed: 2 } });
        mockedApi.put.mockResolvedValue({ data: trip });
        mockedApi.delete.mockResolvedValue({ data: { requested: 2, changed: 2 } });
    });

    it('does not offer remove or cover actions outside an album', () => {
        renderToolbar([11, 12]);

        expect(screen.getByRole('button', { name: 'IMAGES_ALBUM_ADD_TO' })).toBeTruthy();
        expect(screen.queryByRole('button', { name: 'IMAGES_ALBUM_REMOVE_FROM' })).toBeNull();
        expect(screen.queryByRole('button', { name: 'IMAGES_ALBUM_SET_COVER' })).toBeNull();
    });

    it('POSTs the selected file ids to the chosen album and summarizes in a toast', async () => {
        renderToolbar([11, 12]);

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_ADD_TO' }));
        const dialog = await screen.findByRole('dialog');
        fireEvent.click(await within(dialog).findByText('Trip'));

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/image/albums/4/items', {
                file_ids: [11, 12],
            })
        );
        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('IMAGES_ALBUM_ITEMS_ADDED:Trip', {
                variant: 'success',
            })
        );
    });

    it('reports skipped photos when only some were added', async () => {
        mockedApi.post.mockResolvedValue({ data: { requested: 2, changed: 1 } });
        renderToolbar([11, 12]);

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_ADD_TO' }));
        const dialog = await screen.findByRole('dialog');
        fireEvent.click(await within(dialog).findByText('Trip'));

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
                'IMAGES_ALBUM_ITEMS_ADDED_PARTIAL:Trip',
                { variant: 'success' }
            )
        );
    });

    it('filters the album list by the search text', async () => {
        renderToolbar([11]);

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_ADD_TO' }));
        const dialog = await screen.findByRole('dialog');
        await within(dialog).findByText('Holiday');
        fireEvent.change(within(dialog).getByLabelText('IMAGES_ALBUM_SEARCH_PLACEHOLDER'), {
            target: { value: 'hol' },
        });

        expect(within(dialog).queryByText('Trip')).toBeNull();
        expect(within(dialog).getByText('Holiday')).toBeTruthy();

        fireEvent.change(within(dialog).getByLabelText('IMAGES_ALBUM_SEARCH_PLACEHOLDER'), {
            target: { value: 'zzz' },
        });
        expect(within(dialog).getByText('IMAGES_ALBUM_NO_RESULTS')).toBeTruthy();
    });

    it('creates a new album inline and then adds the selection to it', async () => {
        mockedApi.post.mockImplementation((url: string) =>
            Promise.resolve({
                data:
                    url === '/image/albums'
                        ? { ...trip, id: 8, name: 'Fresh' }
                        : { requested: 1, changed: 1 },
            })
        );
        renderToolbar([11]);

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_ADD_TO' }));
        const dialog = await screen.findByRole('dialog');
        fireEvent.change(within(dialog).getByLabelText('IMAGES_ALBUM_NAME_LABEL'), {
            target: { value: 'Fresh' },
        });
        fireEvent.click(within(dialog).getByRole('button', { name: 'IMAGES_ALBUM_CREATE' }));

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/image/albums', { name: 'Fresh' })
        );
        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/image/albums/8/items', {
                file_ids: [11],
            })
        );
    });

    it('DELETEs the selected file ids from the open album', async () => {
        renderToolbar([11, 12], 4);

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_REMOVE_FROM' }));

        await waitFor(() =>
            expect(mockedApi.delete).toHaveBeenCalledWith('/image/albums/4/items', {
                data: { file_ids: [11, 12] },
            })
        );
    });

    it('PUTs the cover only for a single selected photo', async () => {
        const { unmount } = renderToolbar([11, 12], 4);
        expect(screen.getByRole('button', { name: 'IMAGES_ALBUM_SET_COVER' })).toBeDisabled();
        unmount();

        renderToolbar([12], 4);
        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_SET_COVER' }));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/image/albums/4', { cover_file_id: 12 })
        );
    });
});
