import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { apiBase } from '@/service';
import ImageUserAlbumsSection from './ImageUserAlbumsSection';

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
    cover_file_id: 9,
    item_count: 3,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
};

const renderSection = (onOpenAlbum = jest.fn()) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <ImageUserAlbumsSection onOpenAlbum={onOpenAlbum} />
        </QueryClientProvider>
    );

describe('ImageUserAlbumsSection (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({
            data: { items: [trip], pagination: { page: 1, page_size: 60, has_next: false } },
        });
        mockedApi.post.mockResolvedValue({ data: { ...trip, id: 5, name: 'New' } });
        mockedApi.put.mockResolvedValue({ data: { ...trip, name: 'Renamed' } });
        mockedApi.delete.mockResolvedValue({ data: undefined });
    });

    it('lists the albums and opens one by id', async () => {
        const onOpenAlbum = jest.fn();
        renderSection(onOpenAlbum);

        fireEvent.click(await screen.findByRole('button', { name: 'IMAGES_COLLECTION_OPEN:Trip' }));

        expect(mockedApi.get).toHaveBeenCalledWith('/image/albums', {
            params: { page: 1, page_size: 60 },
        });
        expect(onOpenAlbum).toHaveBeenCalledWith(4);
    });

    it('POSTs /image/albums with the typed name', async () => {
        renderSection();
        await screen.findByText('Trip');

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_CREATE' }));
        const dialog = await screen.findByRole('dialog');
        fireEvent.change(within(dialog).getByRole('textbox'), { target: { value: '  New ' } });
        fireEvent.click(
            within(dialog).getByRole('button', { name: 'IMAGES_ALBUM_CREATE_CONFIRM' })
        );

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/image/albums', { name: 'New' })
        );
        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('IMAGES_ALBUM_CREATED:New', {
                variant: 'success',
            })
        );
    });

    it('PUTs /image/albums/:id with the new name from the card menu', async () => {
        renderSection();
        await screen.findByText('Trip');

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_MENU_ARIA:Trip' }));
        fireEvent.click(await screen.findByRole('menuitem', { name: 'RENAME' }));
        const dialog = await screen.findByRole('dialog');
        const nameInput = within(dialog).getByRole('textbox');
        expect(nameInput).toHaveValue('Trip');
        fireEvent.change(nameInput, { target: { value: 'Renamed' } });
        fireEvent.click(
            within(dialog).getByRole('button', { name: 'IMAGES_ALBUM_RENAME_CONFIRM' })
        );

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/image/albums/4', { name: 'Renamed' })
        );
    });

    it('DELETEs /image/albums/:id only after the confirmation', async () => {
        renderSection();
        await screen.findByText('Trip');

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_MENU_ARIA:Trip' }));
        fireEvent.click(await screen.findByRole('menuitem', { name: 'DELETE' }));
        const dialog = await screen.findByRole('dialog');
        expect(within(dialog).getByText('IMAGES_ALBUM_DELETE_CONFIRM:Trip')).toBeTruthy();
        expect(mockedApi.delete).not.toHaveBeenCalled();

        fireEvent.click(within(dialog).getByRole('button', { name: 'DELETE' }));

        await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/image/albums/4'));
    });

    it('shows the backend message verbatim when an action fails', async () => {
        mockedApi.post.mockRejectedValue({ response: { data: { error: 'Name already used' } } });
        renderSection();
        await screen.findByText('Trip');

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_ALBUM_CREATE' }));
        const dialog = await screen.findByRole('dialog');
        fireEvent.change(within(dialog).getByRole('textbox'), { target: { value: 'Trip' } });
        fireEvent.click(
            within(dialog).getByRole('button', { name: 'IMAGES_ALBUM_CREATE_CONFIRM' })
        );

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Name already used', {
                variant: 'error',
            })
        );
    });

    it('shows the empty state when there are no albums', async () => {
        mockedApi.get.mockResolvedValue({ data: { items: [], pagination: { has_next: false } } });
        renderSection();

        expect(await screen.findByText('IMAGES_USER_ALBUMS_EMPTY_TITLE')).toBeTruthy();
    });
});
