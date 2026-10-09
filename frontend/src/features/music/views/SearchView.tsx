import { useMemo, useState } from 'react';
import { Box, CircularProgress, List, Typography } from '@mui/material';
import { useInfiniteQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import AddToPlaylistMenu from '@/features/music/components/AddToPlaylistMenu';
import TrackListItem from '@/features/music/components/TrackListItem';
import { createSearchPlaybackContext } from '@/features/music/components/playbackContext';
import { useOptionalGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { useIntersectionObserver } from '@/components/hooks/IntersectionObserver/useIntersectionObserver';
import useI18n from '@/components/i18n/provider/i18nContext';
import { searchMusicTracks } from '@/service/music';
import { MUSIC_COLLECTION_PAGE_SIZE } from './shared';

export default function SearchView() {
    const { t } = useI18n();
    const [searchParams] = useSearchParams();
    const searchText = (searchParams.get('q') ?? '').trim();
    const replaceQueue = useOptionalGlobalMusic()?.replaceQueue;
    const [playlistMenuTarget, setPlaylistMenuTarget] = useState<{
        anchorElement: HTMLElement;
        fileId: number;
    } | null>(null);

    const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['music-search', searchText],
        queryFn: ({ pageParam = 1 }) =>
            searchMusicTracks(searchText, pageParam, MUSIC_COLLECTION_PAGE_SIZE),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
        enabled: searchText !== '',
    });

    const tracks = useMemo(() => data?.pages.flatMap((page) => page.items) ?? [], [data]);

    const { ref: lastTrackRef } = useIntersectionObserver<HTMLDivElement>({
        enabled: hasNextPage && !isFetchingNextPage,
        rootMargin: '400px',
        onIntersect: () => {
            if (hasNextPage && !isFetchingNextPage) {
                void fetchNextPage();
            }
        },
    });

    if (searchText === '') {
        return (
            <Typography color="text.secondary" sx={{ p: 3 }}>
                {t('MUSIC_SEARCH_PROMPT')}
            </Typography>
        );
    }

    if (isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                <CircularProgress />
            </Box>
        );
    }

    const playbackContext = createSearchPlaybackContext(searchText);

    return (
        <Box sx={{ p: 2 }}>
            <Typography variant="h6" fontWeight={700} sx={{ pb: 1 }}>
                {t('MUSIC_SEARCH_TITLE', { query: searchText })}
            </Typography>

            {tracks.length === 0 && (
                <Typography color="text.secondary" sx={{ py: 2 }}>
                    {t('MUSIC_SEARCH_EMPTY')}
                </Typography>
            )}

            <List sx={{ width: '100%', px: 1 }}>
                {tracks.map((track, index) => (
                    <Box key={track.id} ref={index === tracks.length - 1 ? lastTrackRef : null}>
                        <TrackListItem
                            track={track}
                            index={index}
                            onPlay={(_, trackIndex) =>
                                replaceQueue?.(tracks, trackIndex, playbackContext)
                            }
                            onAddToPlaylist={(event, fileId) =>
                                setPlaylistMenuTarget({
                                    anchorElement: event.currentTarget as HTMLElement,
                                    fileId,
                                })
                            }
                        />
                    </Box>
                ))}
            </List>

            <AddToPlaylistMenu
                fileId={playlistMenuTarget?.fileId ?? 0}
                anchorEl={playlistMenuTarget?.anchorElement ?? null}
                onClose={() => setPlaylistMenuTarget(null)}
            />

            {isFetchingNextPage && (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                    <CircularProgress size={32} />
                </Box>
            )}
        </Box>
    );
}
