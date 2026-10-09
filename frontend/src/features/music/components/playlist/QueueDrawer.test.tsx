import {
    fireEvent,
    render,
    screen,
    waitForElementToBeRemoved,
    within,
} from '@testing-library/react';
import QueueDrawer from './QueueDrawer';

const mockUseGlobalMusic = jest.fn();
const mockSetQueueOpen = jest.fn();
const mockPlayTrackFromQueue = jest.fn();
const mockRemoveFromQueue = jest.fn();
const mockClearQueue = jest.fn();
const mockMoveQueueItem = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/utils/music', () => ({
    getMusicTitle: (track: any) => `title-${track.id}`,
    getMusicArtist: (track: any) => `artist-${track.id}`,
    musicMetadata: () => 'meta',
    getTrackDurationSeconds: (metadata?: any) => metadata?.length ?? 0,
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

const queueOf = (ids: number[]) =>
    ids.map((id) => ({ id, queueEntryId: `entry-${id}`, metadata: { length: id * 60 } }));

const mockQueueState = (overrides: Record<string, unknown> = {}) =>
    mockUseGlobalMusic.mockReturnValue({
        queue: queueOf([1, 2, 3]),
        currentIndex: 0,
        queueOpen: true,
        setQueueOpen: mockSetQueueOpen,
        playTrackFromQueue: mockPlayTrackFromQueue,
        removeFromQueue: mockRemoveFromQueue,
        moveQueueItem: mockMoveQueueItem,
        clearQueue: mockClearQueue,
        isPlaying: true,
        playbackContext: {
            labelKey: 'MUSIC_PLAYBACK_CONTEXT_PLAYLIST',
            labelParams: { name: 'Roadtrip' },
        },
        ...overrides,
    });

describe('QueueDrawer', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockQueueState();
    });

    it('renders the current track, playback context and only the upcoming queue', () => {
        render(<QueueDrawer />);

        expect(screen.getByText('MUSIC_NOW_PLAYING')).toBeInTheDocument();
        expect(screen.getByText('title-1')).toBeInTheDocument();
        expect(screen.getByText('dur-60')).toBeInTheDocument();
        expect(
            screen.getByText('MUSIC_PLAYBACK_FROM:MUSIC_PLAYBACK_CONTEXT_PLAYLIST:Roadtrip')
        ).toBeInTheDocument();
        expect(screen.getByText('MUSIC_NEXT_IN_QUEUE')).toBeInTheDocument();
        expect(screen.getByText('title-2')).toBeInTheDocument();
        expect(screen.getByText('title-3')).toBeInTheDocument();
        expect(screen.queryByText(/MUSIC_QUEUE_PLAYED/)).not.toBeInTheDocument();

        fireEvent.click(screen.getByText('title-2'));
        expect(mockPlayTrackFromQueue).toHaveBeenCalledWith(1);

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_QUEUE_CLOSE' }));
        expect(mockSetQueueOpen).toHaveBeenCalledWith(false);
    });

    it('removes an upcoming item by queue entry id', () => {
        render(<QueueDrawer />);

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_QUEUE_REMOVE_TRACK:title-3' }));

        expect(mockRemoveFromQueue).toHaveBeenCalledWith('entry-3');
    });

    it('keeps played tracks in a collapsed section out of the upcoming list', () => {
        mockQueueState({ queue: queueOf([1, 2, 3]), currentIndex: 1 });
        render(<QueueDrawer />);

        expect(screen.queryByText('title-1')).not.toBeInTheDocument();
        const playedToggle = screen.getByRole('button', { name: /MUSIC_QUEUE_PLAYED/ });
        expect(playedToggle).toHaveAttribute('aria-expanded', 'false');

        fireEvent.click(playedToggle);
        expect(playedToggle).toHaveAttribute('aria-expanded', 'true');
        fireEvent.click(screen.getByText('title-1'));
        expect(mockPlayTrackFromQueue).toHaveBeenCalledWith(0);
        expect(
            screen.queryByRole('button', { name: 'MUSIC_QUEUE_MOVE_UP:title-1' })
        ).not.toBeInTheDocument();
    });

    it('moves upcoming items with the keyboard accessible buttons within the upcoming range', () => {
        mockQueueState({ queue: queueOf([1, 2, 3, 4]), currentIndex: 0 });
        render(<QueueDrawer />);

        expect(screen.getByRole('button', { name: 'MUSIC_QUEUE_MOVE_UP:title-2' })).toBeDisabled();
        expect(
            screen.getByRole('button', { name: 'MUSIC_QUEUE_MOVE_DOWN:title-4' })
        ).toBeDisabled();

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_QUEUE_MOVE_DOWN:title-2' }));
        expect(mockMoveQueueItem).toHaveBeenLastCalledWith(1, 2);
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_QUEUE_MOVE_UP:title-4' }));
        expect(mockMoveQueueItem).toHaveBeenLastCalledWith(3, 2);
    });

    it('reorders with native drag and drop', () => {
        mockQueueState({ queue: queueOf([1, 2, 3, 4]), currentIndex: 0 });
        render(<QueueDrawer />);
        const draggedRow = screen.getByText('title-2').closest('li') as HTMLElement;
        const targetRow = screen.getByText('title-4').closest('li') as HTMLElement;

        fireEvent.dragStart(draggedRow);
        fireEvent.dragOver(targetRow);
        fireEvent.drop(targetRow);

        expect(mockMoveQueueItem).toHaveBeenCalledWith(1, 3);
    });

    it('ignores a drop that did not start from a queue row', () => {
        render(<QueueDrawer />);

        fireEvent.drop(screen.getByText('title-3').closest('li') as HTMLElement);

        expect(mockMoveQueueItem).not.toHaveBeenCalled();
    });

    it('asks for confirmation before clearing the queue', async () => {
        render(<QueueDrawer />);

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_QUEUE_CLEAR' }));
        expect(mockClearQueue).not.toHaveBeenCalled();

        const dialog = screen.getByRole('dialog');
        fireEvent.click(within(dialog).getByText('ACTION_CANCEL'));
        await waitForElementToBeRemoved(() => screen.queryByRole('dialog'));
        expect(mockClearQueue).not.toHaveBeenCalled();

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_QUEUE_CLEAR' }));
        const reopenedDialog = screen.getByRole('dialog');
        fireEvent.click(within(reopenedDialog).getByRole('button', { name: 'MUSIC_QUEUE_CLEAR' }));
        expect(mockClearQueue).toHaveBeenCalledTimes(1);
    });

    it('handles empty states and hides optional blocks', () => {
        mockQueueState({
            queue: [],
            currentIndex: undefined,
            queueOpen: false,
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
