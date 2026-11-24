const ALBUM_KEY_SEPARATOR = '::';

export const formatAlbumKeyLabel = (albumKey: string): string => {
    const [artistPart, ...albumParts] = albumKey.split(ALBUM_KEY_SEPARATOR);
    const albumPart = albumParts.join(ALBUM_KEY_SEPARATOR);
    if (!albumPart) return albumKey;
    if (!artistPart) return albumPart;
    return `${albumPart} - ${artistPart}`;
};
