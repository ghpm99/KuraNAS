import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { Playlist } from '@/types/playlist';
import PlaylistDetailSection from './PlaylistDetailSection';
import PlaylistListSection from './PlaylistListSection';

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ replaceQueue: jest.fn(), currentTrack: null, isPlaying: false }),
}));

const playlist = {
    id: 1,
    name: 'Mix',
    description: '',
    is_system: false,
    is_auto: false,
    track_count: 42,
} as Playlist;

describe('playlist sections', () => {
    it('render without backend and without data', () => {
        expect(() =>
            renderWithoutBackend(
                <>
                    <PlaylistListSection
                        playlists={[]}
                        isLoading={false}
                        hasNextPage={false}
                        isFetchingNextPage={false}
                        onSelect={jest.fn()}
                        onDelete={jest.fn()}
                        onLoadMore={jest.fn()}
                        onCreateOpen={jest.fn()}
                    />
                    <PlaylistDetailSection
                        playlist={playlist}
                        tracks={[]}
                        isLoading={false}
                        hasNextPage={false}
                        isFetchingNextPage={false}
                        onBack={jest.fn()}
                        onRemoveTrack={jest.fn()}
                        onLoadMore={jest.fn()}
                    />
                </>
            )
        ).not.toThrow();
    });

    it('list loads the next page through the sentinel button', () => {
        const onLoadMore = jest.fn();
        renderWithoutBackend(
            <PlaylistListSection
                playlists={[playlist]}
                isLoading={false}
                hasNextPage
                isFetchingNextPage={false}
                onSelect={jest.fn()}
                onDelete={jest.fn()}
                onLoadMore={onLoadMore}
                onCreateOpen={jest.fn()}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: /LOAD_MORE|Load more/i }));

        expect(onLoadMore).toHaveBeenCalledTimes(1);
    });

    it('detail loads more tracks through the sentinel and shows the playlist total', () => {
        const onLoadMore = jest.fn();
        renderWithoutBackend(
            <PlaylistDetailSection
                playlist={playlist}
                tracks={[]}
                isLoading={false}
                hasNextPage
                isFetchingNextPage={false}
                onBack={jest.fn()}
                onRemoveTrack={jest.fn()}
                onLoadMore={onLoadMore}
            />
        );

        expect(screen.getByText('42 MUSIC_TRACKS_COUNT')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /LOAD_MORE|Load more/i }));

        expect(onLoadMore).toHaveBeenCalledTimes(1);
    });
});
