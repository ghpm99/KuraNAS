import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import VideoLibraryPlaylistPicker from './VideoLibraryPlaylistPicker';
import type { VideoPlaylistDto } from '@/service/videoPlayback';

const buildPlaylist = (id: number, name: string): VideoPlaylistDto =>
    ({ id, name, type: 'custom' }) as VideoPlaylistDto;

const playlists = [buildPlaylist(1, 'Season 1'), buildPlaylist(2, 'Favorites')];

const withQueryClient = (children: ReactNode) => (
    <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
        {children}
    </QueryClientProvider>
);

describe('features/videos/VideoLibraryPlaylistPicker', () => {
    it('renders without any provider or service mock and without a backend', () => {
        render(
            withQueryClient(
                <VideoLibraryPlaylistPicker
                    videoId={5}
                    playlists={[]}
                    isAddingToPlaylist={false}
                    onSelectPlaylist={jest.fn()}
                    onAddVideo={jest.fn()}
                />
            )
        );

        expect(screen.getByRole('combobox')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'VIDEO_ADD' })).toBeDisabled();
    });

    it('defaults to the first playlist and forwards selection and add', () => {
        const onSelectPlaylist = jest.fn();
        const onAddVideo = jest.fn();
        render(
            withQueryClient(
                <VideoLibraryPlaylistPicker
                    videoId={5}
                    playlists={playlists}
                    isAddingToPlaylist={false}
                    onSelectPlaylist={onSelectPlaylist}
                    onAddVideo={onAddVideo}
                />
            )
        );

        expect(screen.getByRole('combobox')).toHaveValue('1');
        fireEvent.change(screen.getByRole('combobox'), { target: { value: '2' } });
        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_ADD' }));

        expect(onSelectPlaylist).toHaveBeenCalledWith(5, 2);
        expect(onAddVideo).toHaveBeenCalledWith(5);
    });
});
