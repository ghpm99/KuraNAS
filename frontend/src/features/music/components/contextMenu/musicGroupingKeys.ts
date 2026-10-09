import type { IMusicData } from '@/types/music';

const normalizeLookupKey = (rawText: string) =>
    rawText
        .trim()
        .toLowerCase()
        .replace(/[_\-.]/g, ' ')
        .replace(/\s+/g, ' ');

const collapseSpaces = (rawText: string | undefined) => (rawText ?? '').trim().replace(/\s+/g, ' ');

const resolveArtistLabel = (track: IMusicData) =>
    collapseSpaces(track.metadata?.album_artist) || collapseSpaces(track.metadata?.artist);

export const getTrackArtistKey = (track: IMusicData): string => {
    const artistLabel = resolveArtistLabel(track);
    return artistLabel === '' ? '' : normalizeLookupKey(artistLabel);
};

export const getTrackAlbumKey = (track: IMusicData): string => {
    const artistLabel = resolveArtistLabel(track);
    const albumLabel = collapseSpaces(track.metadata?.album);
    if (artistLabel === '' || albumLabel === '') return '';
    return normalizeLookupKey(`${artistLabel}::${albumLabel}`);
};

export const getArtistKeyFromLabel = (artistLabel: string): string => {
    const collapsedLabel = collapseSpaces(artistLabel);
    return collapsedLabel === '' ? '' : normalizeLookupKey(collapsedLabel);
};
