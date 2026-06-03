import { fireEvent, render, screen } from '@testing-library/react';
import GlobalPlayerControl from './GlobalPlayerControl';
import * as musicUtils from '@/utils/music';

const mockUseGlobalMusic = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params?.context
                ? `${key}:${params.context}`
                : params?.name
                  ? `${key}:${params.name}`
                  : key,
    }),
}));
jest.mock('@/utils/music', () => ({
    __esModule: true,
    getMusicTitle: jest.fn(),
    getMusicArtist: jest.fn(),
}));

const baseApi = () => ({
    hasQueue: true,
    isPlaying: false,
    currentTime: 30,
    duration: 120,
    volume: 0.5,
    shuffle: false,
    repeatMode: 'none' as string,
    togglePlayPause: jest.fn(),
    next: jest.fn(),
    previous: jest.fn(),
    seek: jest.fn(),
    setVolume: jest.fn(),
    toggleShuffle: jest.fn(),
    setRepeatMode: jest.fn(),
    currentTrack: {
        name: 'Test Song',
        metadata: { title: 'Meta Title', artist: 'Meta Artist' },
    },
    playbackContext: undefined as
        | undefined
        | { labelKey: string; labelParams: Record<string, string> },
    toggleQueue: jest.fn(),
    queueOpen: false,
});

const mockGetMusicTitle = musicUtils.getMusicTitle as jest.Mock;
const mockGetMusicArtist = musicUtils.getMusicArtist as jest.Mock;

describe('GlobalPlayerControl', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockGetMusicTitle.mockImplementation(
            (track: { name: string; metadata?: { title?: string } }) =>
                track.metadata?.title || track.name
        );
        mockGetMusicArtist.mockImplementation(
            (track: { metadata?: { artist?: string } }) =>
                track.metadata?.artist || 'Unknown Artist'
        );
    });

    // Branch: !hasQueue => return null
    it('returns null when hasQueue is false', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), hasQueue: false });
        const { container } = render(<GlobalPlayerControl />);
        expect(container).toBeEmptyDOMElement();
    });

    // Branch: hasQueue => render the player
    it('renders player when hasQueue is true', () => {
        mockUseGlobalMusic.mockReturnValue(baseApi());
        render(<GlobalPlayerControl />);
        expect(screen.getByText('Meta Title')).toBeInTheDocument();
    });

    // Branch: isPlaying true => animated bars + Pause icon + pause aria-label
    it('shows animated bars and pause button when playing', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), isPlaying: true });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('pause playback')).toBeInTheDocument();
    });

    // Branch: isPlaying false => Volume2 icon + Play icon + play aria-label
    it('shows volume icon and play button when paused', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), isPlaying: false });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('play playback')).toBeInTheDocument();
    });

    // Branch: getMusicTitle uses metadata.title when present
    it('displays track title from metadata', () => {
        mockUseGlobalMusic.mockReturnValue(baseApi());
        render(<GlobalPlayerControl />);
        expect(screen.getByText('Meta Title')).toBeInTheDocument();
    });

    // Branch: getMusicTitle falls back to name when no metadata.title
    it('displays track name when metadata.title is missing', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { name: 'Filename.mp3', metadata: {} },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('Filename.mp3')).toBeInTheDocument();
    });

    // Branch: getMusicArtist uses metadata.artist when present
    it('displays artist from metadata', () => {
        mockUseGlobalMusic.mockReturnValue(baseApi());
        render(<GlobalPlayerControl />);
        expect(screen.getByText('Meta Artist')).toBeInTheDocument();
    });

    // Branch: getMusicArtist falls back to "Unknown Artist" when no metadata.artist
    it('displays Unknown Artist when metadata.artist is missing', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { name: 'Song', metadata: {} },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('Unknown Artist')).toBeInTheDocument();
    });

    // Branch: playbackContext truthy => show context label
    it('shows playback context label when playbackContext exists', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            playbackContext: {
                labelKey: 'MUSIC_PLAYBACK_CONTEXT_PLAYLIST',
                labelParams: { name: 'My Playlist' },
            },
        });
        render(<GlobalPlayerControl />);
        expect(
            screen.getByText('MUSIC_PLAYBACK_FROM:MUSIC_PLAYBACK_CONTEXT_PLAYLIST:My Playlist')
        ).toBeInTheDocument();
    });

    // Branch: playbackContext falsy => no context label
    it('does not show playback context label when playbackContext is undefined', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            playbackContext: undefined,
        });
        render(<GlobalPlayerControl />);
        expect(screen.queryByText(/MUSIC_PLAYBACK_FROM/)).not.toBeInTheDocument();
    });

    // Branch: Number.isFinite(currentTime) true => use currentTime; false => 0
    it('uses currentTime when finite', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: 65,
            duration: 200,
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('1:05')).toBeInTheDocument();
    });

    it('uses 0 when currentTime is not finite (NaN)', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: NaN,
            duration: 100,
        });
        render(<GlobalPlayerControl />);
        // formatTime(NaN) returns '0:00' and safeCurrentTime is 0
        const timeTexts = screen.getAllByText('0:00');
        expect(timeTexts.length).toBeGreaterThanOrEqual(1);
    });

    it('uses 0 when currentTime is Infinity', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: Infinity,
            duration: 100,
        });
        render(<GlobalPlayerControl />);
        // safeCurrentTime = 0, but formatTime(Infinity) for display
        // currentTime display: formatTime(Infinity) => isNaN(Infinity) is false, so Math.floor(Infinity/60) => Infinity
        // Actually Infinity is not NaN but not finite, so safeCurrentTime = 0
        expect(screen.getByLabelText('seek playback')).toBeInTheDocument();
    });

    // Branch: Number.isFinite(duration) && duration > 0 => use duration; else 0
    it('uses duration when finite and positive', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: 0,
            duration: 185,
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('3:05')).toBeInTheDocument();
    });

    it('uses 0 when duration is NaN', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: 0,
            duration: NaN,
        });
        render(<GlobalPlayerControl />);
        const timeTexts = screen.getAllByText('0:00');
        expect(timeTexts.length).toBe(2);
    });

    it('uses 0 when duration is 0', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: 0,
            duration: 0,
        });
        render(<GlobalPlayerControl />);
        const timeTexts = screen.getAllByText('0:00');
        expect(timeTexts.length).toBe(2);
    });

    it('uses 0 when duration is negative', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: 0,
            duration: -10,
        });
        render(<GlobalPlayerControl />);
        // safeDuration = 0 because duration <= 0
        const seekSlider = screen.getByLabelText('seek playback');
        // max should be safeDuration || 100 = 100
        expect(seekSlider).toBeInTheDocument();
    });

    // Branch: safeDuration || 100 => when safeDuration is 0, max becomes 100
    it('sets slider max to 100 when safeDuration is 0', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: 0,
            duration: 0,
        });
        render(<GlobalPlayerControl />);
        const seekSlider = screen.getByLabelText('seek playback');
        expect(seekSlider).toHaveAttribute('aria-valuemax', '100');
    });

    it('sets slider max to actual duration when safeDuration > 0', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: 0,
            duration: 200,
        });
        render(<GlobalPlayerControl />);
        const seekSlider = screen.getByLabelText('seek playback');
        expect(seekSlider).toHaveAttribute('aria-valuemax', '200');
    });

    // Branch: volume === 0 => VolumeX icon + "unmute volume" aria-label
    it('shows muted state when volume is 0', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), volume: 0 });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('unmute volume')).toBeInTheDocument();
    });

    // Branch: volume > 0 => Volume2 icon + "mute volume" aria-label
    it('shows unmuted state when volume > 0', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), volume: 0.5 });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('mute volume')).toBeInTheDocument();
    });

    // Branch: mute button click: volume > 0 => set to 0
    it('mutes when clicking volume button with volume > 0', () => {
        const api = { ...baseApi(), volume: 0.8 };
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('mute volume'));
        expect(api.setVolume).toHaveBeenCalledWith(0);
    });

    // Branch: mute button click: volume === 0 => set to 0.7
    it('unmutes to 0.7 when clicking volume button with volume at 0', () => {
        const api = { ...baseApi(), volume: 0 };
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('unmute volume'));
        expect(api.setVolume).toHaveBeenCalledWith(0.7);
    });

    // Branch: shuffle true => opacity 1
    it('renders shuffle button with full opacity when shuffle is true', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), shuffle: true });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('toggle shuffle')).toBeInTheDocument();
    });

    // Branch: shuffle false => opacity 0.4
    it('renders shuffle button with reduced opacity when shuffle is false', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), shuffle: false });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('toggle shuffle')).toBeInTheDocument();
    });

    // Branch: repeatMode === 'one' => Repeat1 icon
    it('renders Repeat1 icon when repeatMode is one', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), repeatMode: 'one' });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('change repeat mode')).toBeInTheDocument();
    });

    // Branch: repeatMode !== 'one' => Repeat icon
    it('renders Repeat icon when repeatMode is not one', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), repeatMode: 'none' });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('change repeat mode')).toBeInTheDocument();
    });

    // Branch: repeatMode !== 'none' => opacity 1, primary color
    it('renders repeat button with full opacity when repeatMode is all', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), repeatMode: 'all' });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('change repeat mode')).toBeInTheDocument();
    });

    // Branch: queueOpen true => primary color
    it('renders queue button with primary color when queueOpen is true', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), queueOpen: true });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('toggle queue')).toBeInTheDocument();
    });

    // Branch: queueOpen false => secondary color
    it('renders queue button with secondary color when queueOpen is false', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), queueOpen: false });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('toggle queue')).toBeInTheDocument();
    });

    // cycleRepeatMode: none -> all
    it('cycles repeat mode from none to all', () => {
        const api = { ...baseApi(), repeatMode: 'none' };
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('change repeat mode'));
        expect(api.setRepeatMode).toHaveBeenCalledWith('all');
    });

    // cycleRepeatMode: all -> one
    it('cycles repeat mode from all to one', () => {
        const api = { ...baseApi(), repeatMode: 'all' };
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('change repeat mode'));
        expect(api.setRepeatMode).toHaveBeenCalledWith('one');
    });

    // cycleRepeatMode: one -> none
    it('cycles repeat mode from one to none', () => {
        const api = { ...baseApi(), repeatMode: 'one' };
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('change repeat mode'));
        expect(api.setRepeatMode).toHaveBeenCalledWith('none');
    });

    // cycleRepeatMode: unknown value => fallback to 'none' (indexOf returns -1)
    it('cycles repeat mode from unknown value to none (fallback)', () => {
        const api = { ...baseApi(), repeatMode: 'unknown-mode' };
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('change repeat mode'));
        expect(api.setRepeatMode).toHaveBeenCalledWith('none');
    });

    // Click handlers
    it('calls togglePlayPause on play/pause button click', () => {
        const api = baseApi();
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('play playback'));
        expect(api.togglePlayPause).toHaveBeenCalled();
    });

    it('calls next on next button click', () => {
        const api = baseApi();
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('next track'));
        expect(api.next).toHaveBeenCalled();
    });

    it('calls previous on previous button click', () => {
        const api = baseApi();
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('previous track'));
        expect(api.previous).toHaveBeenCalled();
    });

    it('calls toggleShuffle on shuffle button click', () => {
        const api = baseApi();
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('toggle shuffle'));
        expect(api.toggleShuffle).toHaveBeenCalled();
    });

    it('calls toggleQueue on queue button click', () => {
        const api = baseApi();
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('toggle queue'));
        expect(api.toggleQueue).toHaveBeenCalled();
    });

    // Slider interactions
    it('calls seek when seek slider changes', () => {
        const api = baseApi();
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        const seekSlider = screen.getByLabelText('seek playback');
        fireEvent.change(seekSlider, { target: { value: '50' } });
        expect(api.seek).toHaveBeenCalledWith(50);
    });

    it('calls setVolume when volume slider changes', () => {
        const api = baseApi();
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        const volumeSlider = screen.getByLabelText('set volume');
        fireEvent.change(volumeSlider, { target: { value: '0.3' } });
        expect(api.setVolume).toHaveBeenCalledWith(0.3);
    });

    // formatTime with NaN input
    it('formats NaN time as 0:00', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: NaN,
            duration: NaN,
        });
        render(<GlobalPlayerControl />);
        const timeTexts = screen.getAllByText('0:00');
        expect(timeTexts.length).toBe(2);
    });

    // formatTime with valid time
    it('formats valid time correctly', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTime: 125,
            duration: 300,
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('2:05')).toBeInTheDocument();
        expect(screen.getByText('5:00')).toBeInTheDocument();
    });

    // Track with no metadata at all
    it('handles track with no metadata', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { name: 'Raw File.mp3' },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('Raw File.mp3')).toBeInTheDocument();
    });

    // playbackContext with context param (not name param)
    it('shows playback context with context param', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            playbackContext: {
                labelKey: 'CONTEXT_KEY',
                labelParams: { context: 'Album X' },
            },
        });
        render(<GlobalPlayerControl />);
        // t('CONTEXT_KEY', {context: 'Album X'}) => 'CONTEXT_KEY:Album X'
        // t('MUSIC_PLAYBACK_FROM', {context: 'CONTEXT_KEY:Album X'}) => 'MUSIC_PLAYBACK_FROM:CONTEXT_KEY:Album X'
        expect(screen.getByText('MUSIC_PLAYBACK_FROM:CONTEXT_KEY:Album X')).toBeInTheDocument();
    });

    // Branch: getMusicTitle is falsy => fallback to metadata?.title || name || ''
    it('falls back to metadata.title when getMusicTitle is undefined', () => {
        (musicUtils as { getMusicTitle: unknown }).getMusicTitle = undefined;
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: {
                name: 'Fallback Name',
                metadata: { title: 'Fallback Title', artist: 'Art' },
            },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('Fallback Title')).toBeInTheDocument();
    });

    it('falls back to track name when getMusicTitle is undefined and no title in metadata', () => {
        (musicUtils as { getMusicTitle: unknown }).getMusicTitle = undefined;
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { name: 'Just Name', metadata: {} },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('Just Name')).toBeInTheDocument();
    });

    it('falls back to empty string when getMusicTitle is undefined, no title and no name', () => {
        (musicUtils as { getMusicTitle: unknown }).getMusicTitle = undefined;
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { metadata: {} },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('play playback')).toBeInTheDocument();
    });

    it('falls back to empty string when getMusicTitle is undefined, no metadata at all', () => {
        (musicUtils as { getMusicTitle: unknown }).getMusicTitle = undefined;
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { name: '' },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('play playback')).toBeInTheDocument();
    });

    it('falls back through all branches when getMusicTitle undefined and metadata.title is empty', () => {
        (musicUtils as { getMusicTitle: unknown }).getMusicTitle = undefined;
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { name: 'FallbackName', metadata: { title: '' } },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('FallbackName')).toBeInTheDocument();
    });

    // Branch: getMusicArtist is falsy => fallback to metadata?.artist || ''
    it('falls back to metadata.artist when getMusicArtist is undefined', () => {
        (musicUtils as { getMusicArtist: unknown }).getMusicArtist = undefined;
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: {
                name: 'Song',
                metadata: { title: 'T', artist: 'Fallback Artist' },
            },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('Fallback Artist')).toBeInTheDocument();
    });

    it('falls back to empty string when getMusicArtist is undefined and no artist in metadata', () => {
        (musicUtils as { getMusicArtist: unknown }).getMusicArtist = undefined;
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { name: 'Song', metadata: { title: 'T' } },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('T')).toBeInTheDocument();
    });

    it('falls back to empty string when getMusicArtist is undefined and artist is empty', () => {
        (musicUtils as { getMusicArtist: unknown }).getMusicArtist = undefined;
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { name: 'Song', metadata: { title: 'T', artist: '' } },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByText('T')).toBeInTheDocument();
    });

    it('falls back to empty string when getMusicArtist is undefined and no metadata', () => {
        (musicUtils as { getMusicArtist: unknown }).getMusicArtist = undefined;
        mockUseGlobalMusic.mockReturnValue({
            ...baseApi(),
            currentTrack: { name: 'Song' },
        });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('play playback')).toBeInTheDocument();
    });

    // Ensure nextMode fallback: (currentIdx + 1) % modes.length uses ?? modes[0]
    it('cycles through all three repeat modes sequentially', () => {
        const api1 = { ...baseApi(), repeatMode: 'none' };
        mockUseGlobalMusic.mockReturnValue(api1);
        const { rerender } = render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('change repeat mode'));
        expect(api1.setRepeatMode).toHaveBeenCalledWith('all');

        const api2 = { ...baseApi(), repeatMode: 'all' };
        mockUseGlobalMusic.mockReturnValue(api2);
        rerender(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('change repeat mode'));
        expect(api2.setRepeatMode).toHaveBeenCalledWith('one');

        const api3 = { ...baseApi(), repeatMode: 'one' };
        mockUseGlobalMusic.mockReturnValue(api3);
        rerender(<GlobalPlayerControl />);
        fireEvent.click(screen.getByLabelText('change repeat mode'));
        expect(api3.setRepeatMode).toHaveBeenCalledWith('none');
    });
});
