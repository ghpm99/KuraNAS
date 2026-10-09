import { useCallback, useContext, useSyncExternalStore } from 'react';
import { QueryClientContext, type Query } from '@tanstack/react-query';
import { toggleStarredFile } from '@/service/files';
import {
    getStarredOverride,
    setStarredOverride,
    subscribeToStarredOverrides,
} from './trackStarOverrides';

interface StarrableTrack {
    id: number;
    starred?: boolean;
}

const starredDependentQueryPrefixes = [
    'music',
    'automatic-playlists',
    'playlists',
    'playlist-tracks',
];

const isStarredDependentQuery = (query: Query): boolean => {
    const queryRoot = query.queryKey[0];
    return (
        typeof queryRoot === 'string' &&
        starredDependentQueryPrefixes.some((prefix) => queryRoot.startsWith(prefix))
    );
};

export default function useTrackStar(track: StarrableTrack) {
    const starredOverride = useSyncExternalStore(subscribeToStarredOverrides, () =>
        getStarredOverride(track.id)
    );
    const queryClient = useContext(QueryClientContext);
    const isStarred = starredOverride ?? Boolean(track.starred);

    const toggleStar = useCallback(async () => {
        const wasStarred = getStarredOverride(track.id) ?? Boolean(track.starred);
        setStarredOverride(track.id, !wasStarred);
        try {
            await toggleStarredFile(track.id);
        } catch {
            setStarredOverride(track.id, wasStarred);
            return;
        }
        await queryClient?.invalidateQueries({ predicate: isStarredDependentQuery });
    }, [track.id, track.starred, queryClient]);

    return { isStarred, toggleStar };
}
