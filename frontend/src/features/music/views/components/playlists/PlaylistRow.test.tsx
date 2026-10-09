import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { Playlist } from '@/types/playlist';
import PlaylistRow from './PlaylistRow';

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ replaceQueue: jest.fn() }),
}));

const playlist = {
    id: 4,
    name: 'Mix',
    description: 'desc',
    is_system: false,
    is_auto: false,
    track_count: 3,
} as Playlist;

describe('PlaylistRow', () => {
    it('renders without any backend', () => {
        expect(() =>
            renderWithoutBackend(
                <PlaylistRow
                    playlist={playlist}
                    canDelete
                    onSelect={jest.fn()}
                    onDeleteRequest={jest.fn()}
                />
            )
        ).not.toThrow();
    });

    it('selects with the keyboard and requests deletion without selecting', () => {
        const onSelect = jest.fn();
        const onDeleteRequest = jest.fn();
        renderWithoutBackend(
            <PlaylistRow
                playlist={playlist}
                canDelete
                onSelect={onSelect}
                onDeleteRequest={onDeleteRequest}
            />
        );

        fireEvent.keyDown(screen.getByText('Mix'), { key: 'Enter' });
        expect(onSelect).toHaveBeenCalledWith(playlist);
        onSelect.mockClear();

        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));

        expect(onDeleteRequest).toHaveBeenCalledWith(playlist);
        expect(onSelect).not.toHaveBeenCalled();
    });

    it('hides the delete action when the playlist cannot be deleted', () => {
        renderWithoutBackend(
            <PlaylistRow
                playlist={playlist}
                canDelete={false}
                onSelect={jest.fn()}
                onDeleteRequest={jest.fn()}
            />
        );

        expect(screen.queryByRole('button', { name: 'DELETE' })).not.toBeInTheDocument();
    });
});
