import { useState } from 'react';
import { Box, Button, CircularProgress, List, Typography } from '@mui/material';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import { Plus } from 'lucide-react';
import { Playlist } from '@/types/playlist';
import useI18n from '@/components/i18n/provider/i18nContext';
import DeletePlaylistDialog from './DeletePlaylistDialog';
import PlaylistRow from './PlaylistRow';

type PlaylistListSectionProps = {
    playlists: Playlist[];
    isLoading: boolean;
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    onSelect: (playlist: Playlist) => void;
    onDelete: (playlistId: number) => void;
    onLoadMore: () => void;
    onCreateOpen: () => void;
};

export default function PlaylistListSection({
    playlists,
    isLoading,
    hasNextPage,
    isFetchingNextPage,
    onSelect,
    onDelete,
    onLoadMore,
    onCreateOpen,
}: PlaylistListSectionProps) {
    const { t } = useI18n();
    const [playlistPendingDeletion, setPlaylistPendingDeletion] = useState<Playlist | null>(null);

    if (isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                <CircularProgress />
            </Box>
        );
    }

    const ownPlaylists = playlists.filter((playlist) => !playlist.is_ai_generated);
    const suggestedPlaylists = playlists.filter((playlist) => playlist.is_ai_generated);

    const confirmDeletion = () => {
        if (playlistPendingDeletion) onDelete(playlistPendingDeletion.id);
        setPlaylistPendingDeletion(null);
    };

    return (
        <Box sx={{ p: 1 }}>
            <Box
                sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    p: 1,
                    mb: 1,
                }}
            >
                <Typography variant="h6" fontWeight={700}>
                    {t('MUSIC_PLAYLISTS')}
                </Typography>
                <Button
                    startIcon={<Plus size={16} />}
                    size="small"
                    variant="contained"
                    onClick={onCreateOpen}
                >
                    {t('MUSIC_NEW')}
                </Button>
            </Box>

            <List sx={{ width: '100%' }}>
                {ownPlaylists.map((playlist) => (
                    <PlaylistRow
                        key={playlist.id}
                        playlist={playlist}
                        canDelete={!playlist.is_system}
                        onSelect={onSelect}
                        onDeleteRequest={setPlaylistPendingDeletion}
                    />
                ))}
            </List>

            {ownPlaylists.length === 0 && (
                <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ textAlign: 'center', p: 4 }}
                >
                    {t('MUSIC_NO_PLAYLISTS_MSG')}
                </Typography>
            )}

            {suggestedPlaylists.length > 0 && (
                <Box component="section" aria-label={t('MUSIC_PLAYLISTS_SUGGESTIONS')} sx={{ mt: 2 }}>
                    <Typography variant="h6" fontWeight={700} sx={{ p: 1 }}>
                        {t('MUSIC_PLAYLISTS_SUGGESTIONS')}
                    </Typography>
                    <List sx={{ width: '100%' }}>
                        {suggestedPlaylists.map((playlist) => (
                            <PlaylistRow
                                key={playlist.id}
                                playlist={playlist}
                                canDelete={false}
                                onSelect={onSelect}
                                onDeleteRequest={setPlaylistPendingDeletion}
                            />
                        ))}
                    </List>
                </Box>
            )}

            <LoadMoreSentinel
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={onLoadMore}
            />

            <DeletePlaylistDialog
                playlistName={playlistPendingDeletion?.name ?? null}
                onConfirm={confirmDeletion}
                onCancel={() => setPlaylistPendingDeletion(null)}
            />
        </Box>
    );
}
