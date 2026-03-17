import { render, screen } from '@testing-library/react';
import MusicContent from './musicContent';

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
jest.mock('react-router-dom', () => ({
    Outlet: () => <div>OutletView</div>,
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

describe('musicContent', () => {
    it('renders routed content and queue panel branch', () => {
        mockUseGlobalMusic.mockReturnValue({
            hasQueue: true,
            queue: [],
            currentIndex: undefined,
            queueOpen: false,
            setQueueOpen: jest.fn(),
            playTrackFromQueue: jest.fn(),
            removeFromQueue: jest.fn(),
            clearQueue: jest.fn(),
            isPlaying: false,
        });

        render(<MusicContent />);
        expect(screen.getByText('OutletView')).toBeInTheDocument();
        expect(screen.getAllByText('MUSIC_QUEUE').length).toBeGreaterThan(0);

        mockUseGlobalMusic.mockReturnValue({
            hasQueue: false,
            queue: [],
            currentIndex: undefined,
            queueOpen: false,
            setQueueOpen: jest.fn(),
            playTrackFromQueue: jest.fn(),
            removeFromQueue: jest.fn(),
            clearQueue: jest.fn(),
            isPlaying: false,
        });
        render(<MusicContent />);
        expect(screen.getAllByText('OutletView').length).toBeGreaterThan(0);
    });
});
