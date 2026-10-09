import { createEvent, fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { Playlist, PlaylistTrack } from '@/types/playlist';
import PlaylistDetailSection from './PlaylistDetailSection';

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ replaceQueue: jest.fn(), currentTrack: null, isPlaying: false }),
}));

const userPlaylist = {
    id: 1,
    name: 'Mix',
    description: 'desc',
    is_system: false,
    is_auto: false,
    is_ai_generated: false,
    track_count: 3,
} as Playlist;

const trackNamed = (id: number, name: string) =>
    ({
        id,
        position: id,
        added_at: '',
        file: { id: id * 10, name, metadata: { title: name, artist: 'Band' } },
    }) as unknown as PlaylistTrack;

const tracks = [trackNamed(1, 'Alpha'), trackNamed(2, 'Beta'), trackNamed(3, 'Gamma')];

const renderDetail = (
    overrides: Partial<React.ComponentProps<typeof PlaylistDetailSection>> = {}
) => {
    const props = {
        playlist: userPlaylist,
        tracks,
        isLoading: false,
        hasNextPage: false,
        isFetchingNextPage: false,
        onBack: jest.fn(),
        onRemoveTrack: jest.fn(),
        onLoadMore: jest.fn(),
        onRenamePlaylist: jest.fn(),
        onMoveTrack: jest.fn(),
        ...overrides,
    };
    renderWithoutBackend(<PlaylistDetailSection {...props} />);
    return props;
};

const dragAndDrop = (sourceTitle: string, targetTitle: string) => {
    const sourceItem = screen.getByText(sourceTitle).closest('li') as HTMLElement;
    const targetItem = screen.getByText(targetTitle).closest('li') as HTMLElement;
    fireEvent.dragStart(sourceItem);
    fireEvent.dragOver(targetItem);
    fireEvent.drop(targetItem);
};

describe('PlaylistDetailSection editing', () => {
    it('renders without edit handlers and without any backend', () => {
        expect(() =>
            renderDetail({ onRenamePlaylist: undefined, onMoveTrack: undefined })
        ).not.toThrow();
        expect(screen.queryByRole('button', { name: 'MUSIC_PLAYLIST_EDIT' })).not.toBeInTheDocument();
        expect(screen.queryAllByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_UP' })).toHaveLength(0);
    });

    it('renames through the edit dialog and closes it once saved', () => {
        const onRenamePlaylist = jest.fn((_name: string, _description: string, onSaved: () => void) =>
            onSaved()
        );
        renderDetail({ onRenamePlaylist });

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_EDIT' }));
        expect(screen.getByLabelText('NAME')).toHaveValue('Mix');
        fireEvent.change(screen.getByLabelText('NAME'), { target: { value: 'Road trip' } });
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_SAVE' }));

        expect(onRenamePlaylist).toHaveBeenCalledWith('Road trip', 'desc', expect.any(Function));
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('moves a track one position with the arrow buttons', () => {
        const { onMoveTrack } = renderDetail();
        const moveUpButtons = screen.getAllByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_UP' });
        const moveDownButtons = screen.getAllByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_DOWN' });

        expect(moveUpButtons[0]).toBeDisabled();
        expect(moveDownButtons[2]).toBeDisabled();

        fireEvent.click(moveDownButtons[0]!);
        expect(onMoveTrack).toHaveBeenLastCalledWith(10, 2);

        fireEvent.click(moveUpButtons[2]!);
        expect(onMoveTrack).toHaveBeenLastCalledWith(30, 2);
    });

    it('keeps the last loaded track movable down while more pages exist', () => {
        const { onMoveTrack } = renderDetail({ hasNextPage: true });
        const lastMoveDown = screen.getAllByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_DOWN' })[2]!;

        fireEvent.click(lastMoveDown);

        expect(onMoveTrack).toHaveBeenCalledWith(30, 4);
    });

    it('reorders by dragging a track onto another one', () => {
        const { onMoveTrack } = renderDetail();

        dragAndDrop('Alpha', 'Gamma');

        expect(onMoveTrack).toHaveBeenCalledWith(10, 3);
    });

    it('ignores a drop onto the dragged track itself or without a drag in progress', () => {
        const { onMoveTrack } = renderDetail();

        dragAndDrop('Beta', 'Beta');
        const targetItem = screen.getByText('Gamma').closest('li') as HTMLElement;
        const dragOver = createEvent.dragOver(targetItem);
        fireEvent(targetItem, dragOver);
        fireEvent.drop(targetItem);

        expect(dragOver.defaultPrevented).toBe(false);
        expect(onMoveTrack).not.toHaveBeenCalled();
    });

    it('does not offer editing or reordering for suggestions and automatic playlists', () => {
        renderDetail({ playlist: { ...userPlaylist, is_ai_generated: true } });

        expect(screen.queryByRole('button', { name: 'MUSIC_PLAYLIST_EDIT' })).not.toBeInTheDocument();
        expect(screen.queryAllByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_UP' })).toHaveLength(0);
    });
});
