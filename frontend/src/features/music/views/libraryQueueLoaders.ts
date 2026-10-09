import { queueToTracks } from '@/features/music/components/musicQueueTracks';
import { getMusicQueueByAlbum, getMusicQueueByArtist } from '@/service/music';

export const loadAlbumTracks = (albumKey: string) =>
    getMusicQueueByAlbum(albumKey).then(queueToTracks);

export const loadArtistTracks = (artistKey: string) =>
    getMusicQueueByArtist(artistKey).then(queueToTracks);
