import { findStartIndex, queueEntryToTrack, queueToTracks } from './musicQueueTracks';

const entry = {
    file_id: 7,
    name: 'song.mp3',
    path: '/music/song.mp3',
    format: '.mp3',
    title: 'Song',
    artist: 'Artist',
    album: 'Album',
    length: 95,
};

describe('musicQueueTracks', () => {
    it('maps a queue entry to the track shape the player reads', () => {
        expect(queueEntryToTrack(entry)).toMatchObject({
            id: 7,
            name: 'song.mp3',
            path: '/music/song.mp3',
            format: '.mp3',
            metadata: { title: 'Song', artist: 'Artist', album: 'Album', length: 95 },
        });
    });

    it('survives an absent or partial queue payload', () => {
        expect(queueToTracks(undefined)).toEqual([]);
        expect(queueToTracks({} as never)).toEqual([]);
        expect(queueToTracks({ items: [entry], truncated: false })).toHaveLength(1);
    });

    it('finds the start index of a track and falls back to the first position', () => {
        const tracks = queueToTracks({
            items: [entry, { ...entry, file_id: 8 }, { ...entry, file_id: 9 }],
            truncated: false,
        });

        expect(findStartIndex(tracks)).toBe(0);
        expect(findStartIndex(tracks, 9)).toBe(2);
        expect(findStartIndex(tracks, 404)).toBe(0);
    });
});
