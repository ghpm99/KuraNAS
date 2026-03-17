import type { IMusicData } from '@/components/providers/musicProvider/musicProvider';

export interface MusicArtistHighlight {
    artist: string;
    trackCount: number;
    albumCount: number;
}

export interface MusicAlbumHighlight {
    album: string;
    artist: string;
    year?: string;
    trackCount: number;
}

const getTrackTimestamp = (track: IMusicData) => {
    const comparableDate = track.created_at || track.updated_at || track.last_interaction;
    const parsedDate = comparableDate ? Date.parse(comparableDate) : Number.NaN;

    return Number.isNaN(parsedDate) ? 0 : parsedDate;
};

export const buildMusicArtistHighlights = (
    tracks: IMusicData[],
    limit: number
): MusicArtistHighlight[] => {
    const groupedArtists = new Map<
        string,
        {
            artist: string;
            trackCount: number;
            albums: Set<string>;
            latestTimestamp: number;
        }
    >();

    for (const track of tracks) {
        const artist = track.metadata?.artist?.trim();

        if (!artist) {
            continue;
        }

        const existingArtist = groupedArtists.get(artist) ?? {
            artist,
            trackCount: 0,
            albums: new Set<string>(),
            latestTimestamp: 0,
        };

        existingArtist.trackCount += 1;
        if (track.metadata?.album?.trim()) {
            existingArtist.albums.add(track.metadata.album.trim());
        }
        existingArtist.latestTimestamp = Math.max(
            existingArtist.latestTimestamp,
            getTrackTimestamp(track)
        );

        groupedArtists.set(artist, existingArtist);
    }

    return [...groupedArtists.values()]
        .sort(
            (left, right) =>
                right.latestTimestamp - left.latestTimestamp ||
                left.artist.localeCompare(right.artist)
        )
        .slice(0, limit)
        .map(({ artist, trackCount, albums }) => ({
            artist,
            trackCount,
            albumCount: albums.size,
        }));
};

export const buildMusicAlbumHighlights = (
    tracks: IMusicData[],
    limit: number
): MusicAlbumHighlight[] => {
    const groupedAlbums = new Map<
        string,
        {
            album: string;
            artist: string;
            year?: string;
            trackCount: number;
            latestTimestamp: number;
        }
    >();

    for (const track of tracks) {
        const album = track.metadata?.album?.trim();
        const artist = track.metadata?.artist?.trim();

        if (!album || !artist) {
            continue;
        }

        const albumKey = `${artist}::${album}`;
        const existingAlbum = groupedAlbums.get(albumKey) ?? {
            album,
            artist,
            year: track.metadata?.year ? String(track.metadata.year) : undefined,
            trackCount: 0,
            latestTimestamp: 0,
        };

        existingAlbum.trackCount += 1;
        existingAlbum.latestTimestamp = Math.max(
            existingAlbum.latestTimestamp,
            getTrackTimestamp(track)
        );

        groupedAlbums.set(albumKey, existingAlbum);
    }

    return [...groupedAlbums.values()]
        .sort(
            (left, right) =>
                right.latestTimestamp - left.latestTimestamp ||
                left.album.localeCompare(right.album)
        )
        .slice(0, limit)
        .map(({ album, artist, year, trackCount }) => ({
            album,
            artist,
            year,
            trackCount,
        }));
};
