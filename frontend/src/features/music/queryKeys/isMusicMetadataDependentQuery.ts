import type { Query } from '@tanstack/react-query';

const musicMetadataDependentQueryPrefixes = [
    'music',
    'automatic-playlists',
    'playlists',
    'playlist-tracks',
];

export const isMusicMetadataDependentQuery = (query: Query): boolean => {
    const queryRoot = query.queryKey[0];
    return (
        typeof queryRoot === 'string' &&
        musicMetadataDependentQueryPrefixes.some((prefix) => queryRoot.startsWith(prefix))
    );
};
