import { useQuery } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { apiBase } from '@/service';
import type { IMusicData } from '@/types/music';
import TrackContextMenu from '../contextMenu/TrackContextMenu';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), put: jest.fn() },
}));

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useOptionalGlobalMusic: () => undefined,
    useGlobalMusic: () => ({ playNext: jest.fn(), addToQueue: jest.fn() }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

interface ActiveQueriesProps {
    fetchMusicList: () => Promise<unknown[]>;
    fetchGlobalSearch: () => Promise<unknown[]>;
    fetchUnrelated: () => Promise<unknown[]>;
}

const ActiveQueries = ({
    fetchMusicList,
    fetchGlobalSearch,
    fetchUnrelated,
}: ActiveQueriesProps) => {
    useQuery({ queryKey: ['music-albums', 'list'], queryFn: fetchMusicList });
    useQuery({ queryKey: ['global-search', 'rock'], queryFn: fetchGlobalSearch });
    useQuery({ queryKey: ['vacuum', 'status'], queryFn: fetchUnrelated });
    return null;
};

const track = { id: 12, name: 'twelve.mp3', metadata: { genre: 'Pop' } } as IMusicData;

describe('TrackMetadataDialog (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({ data: { items: [] } });
        mockedApi.put.mockResolvedValue({ data: track });
    });

    it('puts the edited genres and refetches the music queries', async () => {
        const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        const fetchMusicList = jest.fn().mockResolvedValue([]);
        const fetchGlobalSearch = jest.fn().mockResolvedValue([]);
        const fetchUnrelated = jest.fn().mockResolvedValue([]);

        render(
            <QueryClientProvider client={queryClient}>
                <ActiveQueries
                    fetchMusicList={fetchMusicList}
                    fetchGlobalSearch={fetchGlobalSearch}
                    fetchUnrelated={fetchUnrelated}
                />
                <TrackContextMenu
                    track={track}
                    position={{ top: 1, left: 1 }}
                    onClose={jest.fn()}
                    onAddToPlaylist={jest.fn()}
                />
            </QueryClientProvider>
        );
        await waitFor(() => expect(fetchUnrelated).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_EDIT_INFO' }));
        const genresInput = screen.getByLabelText('MUSIC_METADATA_FIELD_GENRES');
        fireEvent.keyDown(genresInput, { key: 'Backspace' });
        fireEvent.change(genresInput, { target: { value: 'Rock' } });
        fireEvent.keyDown(genresInput, { key: 'Enter' });
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_METADATA_SAVE' }));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/music/tracks/12/metadata', {
                genres: ['Rock'],
            })
        );
        await waitFor(() => expect(fetchMusicList).toHaveBeenCalledTimes(2));
        await waitFor(() => expect(fetchGlobalSearch).toHaveBeenCalledTimes(2));
        expect(fetchUnrelated).toHaveBeenCalledTimes(1);
    });
});
