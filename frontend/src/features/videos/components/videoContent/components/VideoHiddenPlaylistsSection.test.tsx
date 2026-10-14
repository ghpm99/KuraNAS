import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { VideoPlaylistDto } from '@/service/videoPlayback';
import VideoHiddenPlaylistsSection from './VideoHiddenPlaylistsSection';

const createHiddenPlaylist = (id: number, name: string): VideoPlaylistDto => ({
    id,
    type: 'custom',
    source_path: `/${name}`,
    name,
    is_hidden: true,
    is_auto: false,
    group_mode: 'single',
    classification: 'personal',
    item_count: 3,
    cover_video_id: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    last_played_at: null,
    items: [],
});

describe('VideoHiddenPlaylistsSection', () => {
    it('mounts without any backend and shows the empty state', () => {
        renderWithoutBackend(<VideoHiddenPlaylistsSection playlists={[]} />);

        expect(screen.getByText('VIDEO_SECTION_HIDDEN_EMPTY')).toBeInTheDocument();
    });

    it('tolerates a partial playlist without the show handler', () => {
        renderWithoutBackend(
            <VideoHiddenPlaylistsSection
                playlists={[{ id: 1, name: 'So nome' } as VideoPlaylistDto]}
            />
        );

        expect(screen.getByText('So nome')).toBeInTheDocument();
        expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('lists the hidden playlists and shows one through its action', () => {
        const onShowPlaylist = jest.fn();
        const alpha = createHiddenPlaylist(1, 'Alfa');
        renderWithoutBackend(
            <VideoHiddenPlaylistsSection
                playlists={[alpha, createHiddenPlaylist(2, 'Beta')]}
                onShowPlaylist={onShowPlaylist}
            />
        );

        expect(screen.getByText('Alfa')).toBeInTheDocument();
        expect(screen.getByText('Beta')).toBeInTheDocument();
        fireEvent.click(screen.getAllByRole('button', { name: /VIDEO_SHOW_PLAYLIST/ })[0]!);

        expect(onShowPlaylist).toHaveBeenCalledWith(alpha);
    });

    it('disables the show actions while a visibility change is in flight', () => {
        renderWithoutBackend(
            <VideoHiddenPlaylistsSection
                playlists={[createHiddenPlaylist(1, 'Alfa')]}
                isChangingVisibility
                onShowPlaylist={jest.fn()}
            />
        );

        expect(screen.getByRole('button', { name: /VIDEO_SHOW_PLAYLIST/ })).toBeDisabled();
    });

    it('offers loading more hidden playlists when there is a next page', () => {
        const fetchNextPage = jest.fn();
        renderWithoutBackend(
            <VideoHiddenPlaylistsSection
                playlists={[createHiddenPlaylist(1, 'Alfa')]}
                hasNextPage
                fetchNextPage={fetchNextPage}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));

        expect(fetchNextPage).toHaveBeenCalled();
    });
});
