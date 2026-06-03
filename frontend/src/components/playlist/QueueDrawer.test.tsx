import { fireEvent, render, screen } from '@testing-library/react';
import QueueDrawer from './QueueDrawer';

const mockUseGlobalMusic = jest.fn();
const mockSetQueueOpen = jest.fn();
const mockPlayTrackFromQueue = jest.fn();
const mockRemoveFromQueue = jest.fn();
const mockClearQueue = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/utils/music', () => ({
    getMusicTitle: (track: any) => `title-${track.id}`,
    getMusicArtist: (track: any) => `artist-${track.id}`,
    musicMetadata: () => 'meta',
    formatMusicDuration: (duration: number) => `dur-${duration}`,
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) => {
            if (params?.context) {
                return `${key}:${params.context}`;
            }
            if (params?.name) {
                return `${key}:${params.name}`;
            }
            return key;
        },
    }),
}));

describe('QueueDrawer', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseGlobalMusic.mockReturnValue({
            queue: [
                { id: 1, metadata: { duration: 180 } },
                { id: 2, metadata: { duration: 120 } },
                { id: 3 },
            ],
            currentIndex: 0,
            queueOpen: true,
            setQueueOpen: mockSetQueueOpen,
            playTrackFromQueue: mockPlayTrackFromQueue,
            removeFromQueue: mockRemoveFromQueue,
            clearQueue: mockClearQueue,
            isPlaying: true,
            playbackContext: {
                labelKey: 'MUSIC_PLAYBACK_CONTEXT_PLAYLIST',
                labelParams: { name: 'Roadtrip' },
            },
        });
    });

    it('renders the current track, playback context, upcoming queue, and actions', () => {
        const { container } = render(<QueueDrawer />);

        expect(screen.getByText('MUSIC_QUEUE')).toBeInTheDocument();
        expect(screen.getByText('MUSIC_NOW_PLAYING')).toBeInTheDocument();
        expect(screen.getByText('title-1')).toBeInTheDocument();
        expect(screen.getByText('artist-1')).toBeInTheDocument();
        expect(
            screen.getByText('MUSIC_PLAYBACK_FROM:MUSIC_PLAYBACK_CONTEXT_PLAYLIST:Roadtrip')
        ).toBeInTheDocument();
        expect(screen.getByText('dur-180')).toBeInTheDocument();
        expect(screen.getByText('MUSIC_NEXT_IN_QUEUE')).toBeInTheDocument();
        expect(screen.getByText('title-2')).toBeInTheDocument();
        expect(screen.getByText('title-3')).toBeInTheDocument();
        expect(screen.getByText('dur-120')).toBeInTheDocument();

        fireEvent.click(screen.getByText('title-2'));
        expect(mockPlayTrackFromQueue).toHaveBeenCalledWith(1);

        const buttons = container.querySelectorAll('button');
        fireEvent.click(buttons[0]!);
        fireEvent.click(buttons[1]!);
        expect(mockClearQueue).toHaveBeenCalled();
        expect(mockSetQueueOpen).toHaveBeenCalledWith(false);

        const removeButtons = container.querySelectorAll('svg.lucide-trash2');
        fireEvent.click(removeButtons[1]!.closest('button') as HTMLElement);
        expect(mockRemoveFromQueue).toHaveBeenCalledWith(1);
    });

    it('handles empty states, paused indicator, and hides optional metadata blocks', () => {
        mockUseGlobalMusic.mockReturnValue({
            queue: [],
            currentIndex: undefined,
            queueOpen: false,
            setQueueOpen: mockSetQueueOpen,
            playTrackFromQueue: mockPlayTrackFromQueue,
            removeFromQueue: mockRemoveFromQueue,
            clearQueue: mockClearQueue,
            isPlaying: false,
            playbackContext: undefined,
        });

        render(<QueueDrawer />);

        expect(screen.getByText('MUSIC_QUEUE')).toBeInTheDocument();
        expect(screen.queryByText('MUSIC_NOW_PLAYING')).not.toBeInTheDocument();
        expect(screen.queryByText('MUSIC_NEXT_IN_QUEUE')).not.toBeInTheDocument();
        expect(screen.queryByText(/MUSIC_PLAYBACK_FROM/)).not.toBeInTheDocument();
    });
});
