import { useCallback, useContext, useSyncExternalStore } from 'react';
import { QueryClientContext, type Query } from '@tanstack/react-query';
import { toggleStarredFile } from '@/service/files';
import { invalidateStarredViews } from '@/shared/queryKeys/invalidateStarredViews';
import {
    clearStarredOverride,
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

const restoreStarredOverride = (trackId: number, previousOverride: boolean | undefined) => {
    if (previousOverride === undefined) {
        clearStarredOverride(trackId);
        return;
    }
    setStarredOverride(trackId, previousOverride);
};

export default function useTrackStar(track: StarrableTrack) {
    const starredOverride = useSyncExternalStore(subscribeToStarredOverrides, () =>
        getStarredOverride(track.id)
    );
    const queryClient = useContext(QueryClientContext);
    const isStarred = starredOverride ?? Boolean(track.starred);

    const toggleStar = useCallback(async () => {
        const previousOverride = getStarredOverride(track.id);
        const wasStarred = previousOverride ?? Boolean(track.starred);
        const optimisticStarred = !wasStarred;
        const isOverrideOwnedByThisToggle = () =>
            getStarredOverride(track.id) === optimisticStarred;

        setStarredOverride(track.id, optimisticStarred);
        try {
            await toggleStarredFile(track.id);
        } catch {
            if (isOverrideOwnedByThisToggle()) {
                restoreStarredOverride(track.id, previousOverride);
            }
            return;
        }
        if (!queryClient) {
            return;
        }
        await Promise.all([
            queryClient.invalidateQueries({ predicate: isStarredDependentQuery }),
            invalidateStarredViews(queryClient),
        ]);
        if (isOverrideOwnedByThisToggle()) {
            clearStarredOverride(track.id);
        }
    }, [track.id, track.starred, queryClient]);

    return { isStarred, toggleStar };
}
