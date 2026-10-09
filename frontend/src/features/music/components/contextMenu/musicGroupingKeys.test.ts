import type { IMusicData } from '@/types/music';
import { getTrackAlbumKey, getTrackArtistKey } from './musicGroupingKeys';

const trackWith = (metadata: IMusicData['metadata']) => ({ id: 1, metadata }) as IMusicData;

describe('musicGroupingKeys', () => {
    it('mirrors the backend normalization preferring the album artist', () => {
        const track = trackWith({
            artist: 'Solo',
            album_artist: ' The  Band_Name ',
            album: 'Big-Hits',
        });

        expect(getTrackArtistKey(track)).toBe('the band name');
        expect(getTrackAlbumKey(track)).toBe('the band name::big hits');
    });

    it('falls back to the track artist and returns empty keys when data is missing', () => {
        expect(getTrackArtistKey(trackWith({ artist: 'Solo' }))).toBe('solo');
        expect(getTrackAlbumKey(trackWith({ artist: 'Solo' }))).toBe('');
        expect(getTrackArtistKey(trackWith(undefined))).toBe('');
        expect(getTrackAlbumKey(trackWith(undefined))).toBe('');
    });
});
