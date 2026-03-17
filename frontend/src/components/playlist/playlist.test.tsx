import { fireEvent, render, screen } from '@testing-library/react';
import Playlist from './playlist';

const mockUseGlobalMusic = jest.fn();

jest.mock('../providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));
jest.mock('@/utils/music', () => ({
    getMusicTitle: (m: any) => m.name ?? m.metadata?.title ?? '',
    getMusicArtist: (m: any) => m.metadata?.artist ?? 'Unknown Artist',
    musicMetadata: () => 'meta',
    formatMusicDuration: (s: number) =>
        `${Math.floor(s / 60)}:${String(Math.floor(s % 60)).padStart(2, '0')}`,
}));

describe('playlist', () => {
    it('renders queue and plays selected track', () => {
        const playTrackFromQueue = jest.fn();
        mockUseGlobalMusic.mockReturnValue({
            queue: [
                { id: 1, name: 'A' },
                { id: 2, name: 'B' },
            ],
            playTrackFromQueue,
        });

        render(<Playlist />);
        expect(screen.getByText('A')).toBeInTheDocument();
        fireEvent.click(screen.getByLabelText('play A'));
        expect(playTrackFromQueue).toHaveBeenCalledWith(0);
    });
});
