import type { IMusicData } from '@/types/music';

export const buildTrack = (id: number): IMusicData =>
    ({
        id,
        name: `track-${id}.mp3`,
        path: `/music/track-${id}.mp3`,
        metadata: { title: `Track ${id}`, artist: 'Artist' },
    }) as IMusicData;

export const buildTracks = (count: number): IMusicData[] =>
    Array.from({ length: count }, (_, position) => buildTrack(position + 1));
