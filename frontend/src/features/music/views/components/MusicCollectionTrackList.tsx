import { List, Typography } from '@mui/material';
import { useState } from 'react';
import AddToPlaylistMenu from '@/features/music/components/AddToPlaylistMenu';
import TrackListItem from '@/features/music/components/TrackListItem';
import { IMusicData } from '@/types/music';
import useI18n from '@/components/i18n/provider/i18nContext';
import { parseTrackNumber } from '@/utils/music';
import TrackSelectionControls from '@/features/music/selection/TrackSelectionControls';
import useTrackSelection from '@/features/music/selection/useTrackSelection';
import { groupTracksByDisc } from '../groupTracksByDisc';
import MusicCollectionListFeedback from './MusicCollectionListFeedback';

type MusicCollectionTrackListProps = {
    tracks: IMusicData[];
    isLoading: boolean;
    isError: boolean;
    errorMessage?: string;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    isNextPageError: boolean;
    showArtist?: boolean;
    albumDiscCount?: number | null;
    isAlbumLayout?: boolean;
    selectionScopeKey: string;
    onPlayTrack: (track: IMusicData) => void;
    onRetry: () => void;
    fetchNextPage: () => void;
};

export default function MusicCollectionTrackList({
    tracks,
    isLoading,
    isError,
    errorMessage,
    hasNextPage,
    isFetchingNextPage,
    isNextPageError,
    showArtist = true,
    albumDiscCount = null,
    isAlbumLayout = false,
    selectionScopeKey,
    onPlayTrack,
    onRetry,
    fetchNextPage,
}: MusicCollectionTrackListProps) {
    const { t } = useI18n();
    const [menuAnchor, setMenuAnchor] = useState<{
        el: HTMLElement;
        fileId: number;
    } | null>(null);

    const selection = useTrackSelection(selectionScopeKey);

    const discGroups = isAlbumLayout
        ? groupTracksByDisc(tracks)
        : [{ discNumber: 1, tracks: tracks.map((track, position) => ({ track, position })) }];
    const displayedTracks = discGroups.flatMap((discGroup) =>
        discGroup.tracks.map(({ track }) => track)
    );
    const hasMultipleDiscs = isAlbumLayout && (albumDiscCount ?? discGroups.length) > 1;

    return (
        <>
            <TrackSelectionControls selection={selection} loadedTracks={tracks} />
            <MusicCollectionListFeedback
                isLoading={isLoading}
                isError={isError}
                errorMessage={errorMessage}
                isEmpty={tracks.length === 0}
                emptyTitleKey="MUSIC_COLLECTION_TRACKS_EMPTY"
                errorTitleKey="MUSIC_TRACKS_ERROR_TITLE"
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                isNextPageError={isNextPageError}
                onRetry={onRetry}
                fetchNextPage={fetchNextPage}
            >
                {discGroups.map((discGroup) => (
                    <List
                        key={discGroup.discNumber}
                        sx={{ width: '100%' }}
                        subheader={
                            hasMultipleDiscs ? (
                                <Typography
                                    component="h3"
                                    variant="subtitle2"
                                    color="text.secondary"
                                    sx={{ px: 1, pt: 1 }}
                                >
                                    {t('MUSIC_DISC_HEADING', {
                                        number: String(discGroup.discNumber),
                                    })}
                                </Typography>
                            ) : undefined
                        }
                    >
                        {discGroup.tracks.map(({ track, position }) => (
                            <TrackListItem
                                key={track.id}
                                track={track}
                                index={position}
                                trackNumber={
                                    isAlbumLayout ? parseTrackNumber(track.metadata) : undefined
                                }
                                onPlay={onPlayTrack}
                                onAddToPlaylist={(event, fileId) =>
                                    setMenuAnchor({
                                        el: event.currentTarget as HTMLElement,
                                        fileId,
                                    })
                                }
                                showArtist={showArtist}
                                selectionMode={selection.resolveRowSelectionMode(
                                    track,
                                    displayedTracks
                                )}
                            />
                        ))}
                    </List>
                ))}
            </MusicCollectionListFeedback>

            <AddToPlaylistMenu
                fileId={menuAnchor?.fileId ?? 0}
                anchorEl={menuAnchor?.el ?? null}
                onClose={() => setMenuAnchor(null)}
            />
        </>
    );
}
