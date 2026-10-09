import {
    buildPlayableTrack,
    buildTrackPlaybackContext,
    formatTrackDuration,
} from './searchTrackPlayback';

const track = {
    file_id: 7,
    title: 'Time',
    artist: 'Pink Floyd',
    album: 'The Dark Side',
    album_key: 'pink floyd::the dark side',
    duration: 413.5,
    path: '/Music/Time.MP3',
};

describe('components/search/searchTrackPlayback', () => {
    it('builds a playable track keyed by the file id with search metadata', () => {
        const playable = buildPlayableTrack(track);

        expect(playable.id).toBe(7);
        expect(playable.name).toBe('Time.MP3');
        expect(playable.format).toBe('.mp3');
        expect(playable.metadata).toEqual(
            expect.objectContaining({
                title: 'Time',
                artist: 'Pink Floyd',
                album: 'The Dark Side',
                length: 413.5,
            })
        );
    });

    it('falls back to the title and empty format when the path has no name or extension', () => {
        const playable = buildPlayableTrack({ ...track, path: '', title: 'Untitled' });

        expect(playable.name).toBe('Untitled');
        expect(playable.format).toBe('');
    });

    it('uses the album playback context when the track has an album', () => {
        expect(buildTrackPlaybackContext(track).kind).toBe('album');
    });

    it('uses the all tracks playback context when the track has no album', () => {
        expect(buildTrackPlaybackContext({ ...track, album: '' }).kind).toBe('all-tracks');
    });

    it.each([
        { seconds: 413.5, formatted: '6:54' },
        { seconds: 59.4, formatted: '0:59' },
        { seconds: 0, formatted: '' },
        { seconds: Number.NaN, formatted: '' },
    ])('formats $seconds seconds as "$formatted"', ({ seconds, formatted }) => {
        expect(formatTrackDuration(seconds)).toBe(formatted);
    });
});
