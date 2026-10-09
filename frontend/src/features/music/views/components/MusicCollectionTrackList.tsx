import { List } from '@mui/material';
import { useState } from 'react';
import AddToPlaylistMenu from '@/features/music/components/AddToPlaylistMenu';
import TrackListItem from '@/features/music/components/TrackListItem';
import { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import MusicCollectionListFeedback from './MusicCollectionListFeedback';

type MusicCollectionTrackListProps = {
    tracks: IMusicData[];
    isLoading: boolean;
    isError: boolean;
    errorMessage?: string;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    showArtist?: boolean;
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
    showArtist = true,
    onPlayTrack,
    onRetry,
    fetchNextPage,
}: MusicCollectionTrackListProps) {
    const [menuAnchor, setMenuAnchor] = useState<{
        el: HTMLElement;
        fileId: number;
    } | null>(null);

    return (
        <>
            <MusicCollectionListFeedback
                isLoading={isLoading}
                isError={isError}
                errorMessage={errorMessage}
                isEmpty={tracks.length === 0}
                emptyTitleKey="MUSIC_COLLECTION_TRACKS_EMPTY"
                errorTitleKey="MUSIC_TRACKS_ERROR_TITLE"
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                onRetry={onRetry}
                fetchNextPage={fetchNextPage}
            >
                <List sx={{ width: '100%' }}>
                    {tracks.map((track, index) => (
                        <TrackListItem
                            key={track.id}
                            track={track}
                            index={index}
                            onPlay={onPlayTrack}
                            onAddToPlaylist={(event, fileId) =>
                                setMenuAnchor({
                                    el: event.currentTarget as HTMLElement,
                                    fileId,
                                })
                            }
                            showArtist={showArtist}
                        />
                    ))}
                </List>
            </MusicCollectionListFeedback>

            <AddToPlaylistMenu
                fileId={menuAnchor?.fileId ?? 0}
                anchorEl={menuAnchor?.el ?? null}
                onClose={() => setMenuAnchor(null)}
            />
        </>
    );
}
