import { getMusicTitle, getMusicArtist, formatMusicDuration, musicMetadata } from './music';
import type {
    IMusicData,
    IMusicMetadata,
} from '@/features/music/providers/musicProvider/musicProvider';

const createTrack = (overrides: Partial<IMusicData> = {}): IMusicData => ({
    id: 1,
    name: 'track-file.mp3',
    path: '/music/track-file.mp3',
    type: 2,
    format: 'mp3',
    size: 5242880,
    updated_at: '',
    created_at: '',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    ...overrides,
});

const createMetadata = (overrides: Partial<IMusicMetadata> = {}): IMusicMetadata => ({
    id: 1,
    fileId: 1,
    path: '/music/track-file.mp3',
    format: 'mp3',
    title: 'My Song',
    artist: 'Cool Band',
    album: 'Greatest Hits',
    year: 2024,
    genre: 'Rock',
    track: 1,
    disc: 1,
    duration: 245,
    bitrate: 320,
    sampleRate: 44100,
    channels: 2,
    createdAt: '',
    ...overrides,
});

describe('getMusicTitle', () => {
    it('returns metadata title when present', () => {
        const track = createTrack({
            metadata: createMetadata({ title: 'Awesome Track' }),
        });
        expect(getMusicTitle(track)).toBe('Awesome Track');
    });

    it('falls back to file name when metadata has no title', () => {
        const track = createTrack({ metadata: createMetadata({ title: '' }) });
        expect(getMusicTitle(track)).toBe('track-file.mp3');
    });

    it('falls back to file name when metadata is undefined', () => {
        const track = createTrack({ metadata: undefined });
        expect(getMusicTitle(track)).toBe('track-file.mp3');
    });
});

describe('getMusicArtist', () => {
    it('returns metadata artist when present', () => {
        const track = createTrack({
            metadata: createMetadata({ artist: 'The Artist' }),
        });
        expect(getMusicArtist(track)).toBe('The Artist');
    });

    it('returns "Unknown Artist" when metadata has no artist', () => {
        const track = createTrack({ metadata: createMetadata({ artist: '' }) });
        expect(getMusicArtist(track)).toBe('Unknown Artist');
    });

    it('returns "Unknown Artist" when metadata is undefined', () => {
        const track = createTrack({ metadata: undefined });
        expect(getMusicArtist(track)).toBe('Unknown Artist');
    });
});

describe('formatMusicDuration', () => {
    it('formats 0 seconds as 0:00', () => {
        expect(formatMusicDuration(0)).toBe('0:00');
    });

    it('formats seconds less than a minute', () => {
        expect(formatMusicDuration(5)).toBe('0:05');
    });

    it('formats single-digit seconds with leading zero', () => {
        expect(formatMusicDuration(9)).toBe('0:09');
    });

    it('formats exactly one minute', () => {
        expect(formatMusicDuration(60)).toBe('1:00');
    });

    it('formats minutes and seconds correctly', () => {
        expect(formatMusicDuration(125)).toBe('2:05');
    });

    it('formats large durations', () => {
        expect(formatMusicDuration(3661)).toBe('61:01');
    });

    it('floors fractional seconds', () => {
        expect(formatMusicDuration(62.9)).toBe('1:02');
    });
});

describe('musicMetadata', () => {
    it('builds full metadata string with format, size and duration', () => {
        const result = musicMetadata({
            format: 'mp3',
            size: 5242880,
            metadata: createMetadata({ duration: 245 }),
        });
        expect(result).toBe('mp3 - 5.00 MB - 4:05');
    });

    it('omits duration when metadata is undefined', () => {
        const result = musicMetadata({
            format: 'flac',
            size: 1048576,
            metadata: undefined,
        });
        expect(result).toBe('flac - 1.00 MB');
    });

    it('omits duration when metadata has no duration', () => {
        const result = musicMetadata({
            format: 'wav',
            size: 2048,
            metadata: createMetadata({ duration: 0 }),
        });
        expect(result).toBe('wav - 2.00 KB');
    });

    it('omits format prefix when format is empty', () => {
        const result = musicMetadata({
            format: '',
            size: 512,
            metadata: createMetadata({ duration: 30 }),
        });
        expect(result).toBe('512 B - 0:30');
    });

    it('handles empty format and no duration', () => {
        const result = musicMetadata({
            format: '',
            size: 1024,
            metadata: undefined,
        });
        expect(result).toBe('1.00 KB');
    });
});
