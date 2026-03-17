import { fireEvent, render, screen } from '@testing-library/react';
import TrackListItem from './TrackListItem';

const mockUseGlobalMusic = jest.fn();
const mockOnPlay = jest.fn();
const mockOnAddToPlaylist = jest.fn();

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/utils/music', () => ({
    getMusicTitle: (track: any) => track.name,
    getMusicArtist: () => 'artist-9',
    musicMetadata: () => 'meta',
    formatMusicDuration: (duration: number) => `dur-${duration}`,
}));

const baseTrack: any = {
    id: 9,
    name: 'track-9',
    format: 'mp3',
    size: 1000,
    metadata: {
        duration: 180,
    },
};

describe('TrackListItem', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseGlobalMusic.mockReturnValue({
            currentTrack: { id: 9 },
            isPlaying: true,
        });
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
});
