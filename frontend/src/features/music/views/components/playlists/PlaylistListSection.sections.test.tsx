import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { Playlist } from '@/types/playlist';
import PlaylistListSection from './PlaylistListSection';

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ replaceQueue: jest.fn() }),
}));

const ownPlaylist = {
    id: 1,
    name: 'Road trip',
    description: '',
    is_system: false,
    is_auto: false,
    is_ai_generated: false,
    track_count: 2,
} as Playlist;

const suggestedPlaylist = {
    ...ownPlaylist,
    id: 2,
    name: 'Chill mix',
    is_ai_generated: true,
} as Playlist;

const renderList = (playlists: Playlist[], onDelete = jest.fn()) => {
    renderWithoutBackend(
        <PlaylistListSection
            playlists={playlists}
            isLoading={false}
            hasNextPage={false}
            isFetchingNextPage={false}
            onSelect={jest.fn()}
            onDelete={onDelete}
            onLoadMore={jest.fn()}
            onCreateOpen={jest.fn()}
        />
    );
    return onDelete;
};

describe('PlaylistListSection sections and deletion', () => {
    it('shows AI suggestions in their own section without a delete action', () => {
        renderList([ownPlaylist, suggestedPlaylist]);

        const suggestions = screen.getByRole('region', { name: 'MUSIC_PLAYLISTS_SUGGESTIONS' });
        expect(within(suggestions).getByText('Chill mix')).toBeInTheDocument();
        expect(within(suggestions).queryByRole('button', { name: 'DELETE' })).not.toBeInTheDocument();
        expect(screen.getByText('Road trip').closest('section')).toBeNull();
        expect(screen.getAllByRole('button', { name: 'DELETE' })).toHaveLength(1);
    });

    it('omits the suggestions section when there are none', () => {
        renderList([ownPlaylist]);

        expect(screen.queryByRole('region', { name: 'MUSIC_PLAYLISTS_SUGGESTIONS' })).toBeNull();
    });

    it('shows the empty message when only suggestions exist', () => {
        renderList([suggestedPlaylist]);

        expect(screen.getByText('MUSIC_NO_PLAYLISTS_MSG')).toBeInTheDocument();
    });

    it('asks for confirmation before deleting and deletes only after confirming', async () => {
        const onDelete = renderList([ownPlaylist]);

        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
        expect(screen.getByText('MUSIC_PLAYLIST_DELETE_CONFIRM_TITLE')).toBeInTheDocument();
        expect(onDelete).not.toHaveBeenCalled();

        fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'DELETE' }));

        expect(onDelete).toHaveBeenCalledWith(1);
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('keeps the playlist when the confirmation is cancelled', async () => {
        const onDelete = renderList([ownPlaylist]);

        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));

        expect(onDelete).not.toHaveBeenCalled();
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });
});
