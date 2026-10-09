import { getAlbumCoverUrl, getTrackCoverArtwork, getTrackCoverUrl } from './musicCover';

describe('music cover urls', () => {
    it('builds the track cover url on the api base', () => {
        expect(getTrackCoverUrl(7, 96)).toBe('/api/v1/music/tracks/7/cover?size=96');
    });

    it('encodes the album key', () => {
        expect(getAlbumCoverUrl('a b/c', 256)).toBe(
            '/api/v1/music/library/albums/a%20b%2Fc/cover?size=256'
        );
    });
});

describe('media session artwork', () => {
    it('describes the 96, 256 and 512 renditions of the track cover', () => {
        expect(getTrackCoverArtwork(7)).toEqual([
            { src: '/api/v1/music/tracks/7/cover?size=96', sizes: '96x96', type: 'image/jpeg' },
            { src: '/api/v1/music/tracks/7/cover?size=256', sizes: '256x256', type: 'image/jpeg' },
            { src: '/api/v1/music/tracks/7/cover?size=512', sizes: '512x512', type: 'image/jpeg' },
        ]);
    });
});
