import {
    getMusicTitle,
    getMusicArtist,
    formatMusicDuration,
    musicMetadata,
    getTrackDurationSeconds,
    formatTotalDuration,
    parseDiscNumber,
    parseTrackNumber,
} from './music';
import type { IMusicData } from '@/types/music';
import type { IMusicMetadata } from '@/types/music';

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
    mime: 'audio/mpeg',
    title: 'My Song',
    artist: 'Cool Band',
    album: 'Greatest Hits',
    year: '2024',
    genre: 'Rock',
    track_number: '1',
    disc_number: '1',
    length: 245,
    bitrate: 320,
    sample_rate: 44100,
    channels: 2,
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
            metadata: createMetadata({ length: 245 }),
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
            metadata: createMetadata({ length: 0 }),
        });
        expect(result).toBe('wav - 2.00 KB');
    });

    it('omits format prefix when format is empty', () => {
        const result = musicMetadata({
            format: '',
            size: 512,
            metadata: createMetadata({ length: 30 }),
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

describe('backend audio metadata contract', () => {
    const backendAudioMetadataJson = `{
        "mime": "audio/flac",
        "length": 213.4,
        "bitrate": 960000,
        "sample_rate": 44100,
        "channels": 2,
        "bitrate_mode": 0,
        "encoder_info": "reference libFLAC",
        "bit_depth": 16,
        "title": "Time",
        "artist": "Pink Floyd",
        "album": "The Dark Side of the Moon",
        "album_artist": "Pink Floyd",
        "track_number": "4",
        "genre": "Rock",
        "composer": "Wright",
        "year": "1973",
        "recording_date": "1973",
        "encoder": "",
        "publisher": "",
        "original_release_date": "",
        "original_artist": "",
        "lyricist": "",
        "lyrics": "Ticking away",
        "disc_number": "1"
    }`;

    const metadata: IMusicMetadata = JSON.parse(backendAudioMetadataJson);

    it('reads the track duration from the length field', () => {
        expect(getTrackDurationSeconds(metadata)).toBe(213.4);
    });

    it('formats the duration shown beside the file summary', () => {
        const summary = musicMetadata({ format: 'flac', size: 1024, metadata });

        expect(summary).toContain('3:33');
    });

    it('returns zero for missing, negative or non finite lengths', () => {
        expect(getTrackDurationSeconds(undefined)).toBe(0);
        expect(getTrackDurationSeconds({ length: -3 })).toBe(0);
        expect(getTrackDurationSeconds({ length: Number.NaN })).toBe(0);
    });
});

describe('track and disc numbers', () => {
    it('reads the numeric part of "n/total" track numbers', () => {
        expect(parseTrackNumber({ track_number: '2/12' })).toBe(2);
        expect(parseTrackNumber({ track_number: ' 07 ' })).toBe(7);
        expect(parseTrackNumber({ track_number: '3' })).toBe(3);
    });

    it('returns undefined when there is no track number', () => {
        expect(parseTrackNumber(undefined)).toBeUndefined();
        expect(parseTrackNumber({ track_number: '' })).toBeUndefined();
        expect(parseTrackNumber({ track_number: 'A1' })).toBeUndefined();
    });

    it('defaults the disc to 1 and reads "n/total"', () => {
        expect(parseDiscNumber(undefined)).toBe(1);
        expect(parseDiscNumber({ disc_number: '' })).toBe(1);
        expect(parseDiscNumber({ disc_number: '2/3' })).toBe(2);
    });
});

describe('formatTotalDuration', () => {
    const translate = (key: string, options?: Record<string, string>) =>
        `${key}:${JSON.stringify(options)}`;

    it('uses minutes only below one hour', () => {
        expect(formatTotalDuration(125, translate)).toBe('MUSIC_DURATION_MINUTES:{"minutes":"2"}');
        expect(formatTotalDuration(-5, translate)).toBe('MUSIC_DURATION_MINUTES:{"minutes":"0"}');
    });

    it('uses hours and padded minutes from one hour on', () => {
        expect(formatTotalDuration(3600 + 5 * 60, translate)).toBe(
            'MUSIC_DURATION_HOURS_MINUTES:{"hours":"1","minutes":"05"}'
        );
    });
});
