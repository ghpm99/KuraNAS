import { fireEvent, render, screen } from '@testing-library/react';
import TrackListItem from './TrackListItem';

const mockUseGlobalMusic = jest.fn();
const mockOnPlay = jest.fn();
const mockOnAddToPlaylist = jest.fn();
const mockPlayNext = jest.fn();
const mockAddToQueue = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/utils/music', () => ({
    getMusicTitle: (track: any) => track.name,
    getMusicArtist: () => 'artist-9',
    musicMetadata: () => 'meta',
    getTrackDurationSeconds: (metadata?: any) => metadata?.length ?? 0,
    formatMusicDuration: (duration: number) => `dur-${duration}`,
}));

const baseTrack: any = {
    id: 9,
    name: 'track-9',
    format: 'mp3',
    size: 1000,
    metadata: {
        length: 180,
    },
};

describe('TrackListItem', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseGlobalMusic.mockReturnValue({
            currentTrack: { id: 9 },
            isPlaying: true,
            playNext: mockPlayNext,
            addToQueue: mockAddToQueue,
        });
    });

    it('opens the context menu from the more actions button and from right click', () => {
        render(<TrackListItem track={baseTrack} index={0} onPlay={mockOnPlay} />);

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_TRACK_MORE_ACTIONS' }));
        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_PLAY_NEXT' }));
        expect(mockPlayNext).toHaveBeenCalledWith([baseTrack]);
        expect(mockOnPlay).not.toHaveBeenCalled();

        fireEvent.contextMenu(screen.getByRole('button', { name: 'play track-9' }));
        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_ADD_TO_QUEUE' }));
        expect(mockAddToQueue).toHaveBeenCalledWith([baseTrack]);
    });

    it('renders playing state and playlist action', () => {
        render(
            <TrackListItem
                track={baseTrack}
                index={2}
                onPlay={mockOnPlay}
                onAddToPlaylist={mockOnAddToPlaylist}
            />
        );

        expect(screen.getByText('track-9')).toBeInTheDocument();
        expect(screen.getByText('artist-9')).toBeInTheDocument();
        expect(screen.getByText('dur-180')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'play track-9' }));
        expect(mockOnPlay).toHaveBeenCalledWith(baseTrack, 2);

        fireEvent.click(screen.getByRole('button', { name: 'add track-9 to playlist' }));
        expect(mockOnAddToPlaylist).toHaveBeenCalledWith(expect.any(Object), 9);
    });

    it('renders paused current track and fallback layout without artist or duration', () => {
        mockUseGlobalMusic.mockReturnValue({
            currentTrack: { id: 9 },
            isPlaying: false,
        });

        const { rerender } = render(
            <TrackListItem track={baseTrack} index={0} onPlay={mockOnPlay} showArtist={false} />
        );

        expect(screen.queryByText('artist-9')).not.toBeInTheDocument();

        mockUseGlobalMusic.mockReturnValue({
            currentTrack: { id: 1 },
            isPlaying: false,
        });

        rerender(
            <TrackListItem
                track={{ ...baseTrack, id: 10, name: 'track-10', metadata: undefined }}
                index={4}
                onPlay={mockOnPlay}
            />
        );

        expect(screen.getByText('5')).toBeInTheDocument();
        expect(screen.getByText('artist-9')).toBeInTheDocument();
        expect(screen.queryByText(/dur-/)).not.toBeInTheDocument();
    });

    describe('selection mode', () => {
        it('shows a checkbox named after the track and toggles instead of playing on row click', () => {
            const mockOnSelect = jest.fn();
            render(
                <TrackListItem
                    track={baseTrack}
                    index={0}
                    onPlay={mockOnPlay}
                    selectionMode={{
                        isSelected: false,
                        onSelect: mockOnSelect,
                        onToggle: jest.fn(),
                    }}
                />
            );

            expect(screen.getByRole('checkbox', { name: 'MUSIC_SELECT_TRACK' })).not.toBeChecked();

            fireEvent.click(screen.getByRole('button', { name: 'play track-9' }));
            fireEvent.click(screen.getByRole('button', { name: 'play track-9' }), {
                ctrlKey: true,
            });

            expect(mockOnSelect).toHaveBeenCalledTimes(2);
            expect(mockOnSelect).toHaveBeenNthCalledWith(1, baseTrack, { isRangeClick: false });
            expect(mockOnSelect).toHaveBeenNthCalledWith(2, baseTrack, { isRangeClick: false });
            expect(mockOnPlay).not.toHaveBeenCalled();
        });

        it('reports a range click when the row is clicked with shift', () => {
            const mockOnSelect = jest.fn();
            render(
                <TrackListItem
                    track={baseTrack}
                    index={0}
                    onPlay={mockOnPlay}
                    selectionMode={{
                        isSelected: false,
                        onSelect: mockOnSelect,
                        onToggle: jest.fn(),
                    }}
                />
            );

            fireEvent.click(screen.getByRole('button', { name: 'play track-9' }), {
                shiftKey: true,
            });

            expect(mockOnSelect).toHaveBeenCalledWith(baseTrack, { isRangeClick: true });
            expect(mockOnPlay).not.toHaveBeenCalled();
        });

        it('toggles once when the checkbox itself is clicked', () => {
            const mockOnToggle = jest.fn();
            render(
                <TrackListItem
                    track={baseTrack}
                    index={0}
                    onPlay={mockOnPlay}
                    selectionMode={{
                        isSelected: true,
                        onSelect: jest.fn(),
                        onToggle: mockOnToggle,
                    }}
                />
            );

            const checkbox = screen.getByRole('checkbox', { name: 'MUSIC_SELECT_TRACK' });
            expect(checkbox).toBeChecked();
            fireEvent.click(checkbox);

            expect(mockOnToggle).toHaveBeenCalledTimes(1);
            expect(mockOnPlay).not.toHaveBeenCalled();
        });

        it('has no checkbox and still plays outside the selection mode', () => {
            render(<TrackListItem track={baseTrack} index={2} onPlay={mockOnPlay} />);

            expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
            fireEvent.click(screen.getByRole('button', { name: 'play track-9' }));

            expect(mockOnPlay).toHaveBeenCalledWith(baseTrack, 2);
        });
    });
});
