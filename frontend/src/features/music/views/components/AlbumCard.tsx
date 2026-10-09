import { Box, Card, CardActionArea, CardContent, IconButton, Typography } from '@mui/material';
import { Disc, Play } from 'lucide-react';
import CoverArt from '@/features/music/components/CoverArt';
import CollectionContextMenu from '@/features/music/components/contextMenu/CollectionContextMenu';
import { createAlbumPlaybackContext } from '@/features/music/components/playbackContext';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { getAlbumCoverUrl } from '@/service/musicCover';
import { MusicAlbum } from '@/types/music';
import { handleKeyboardActivation } from '../shared';
import { loadAlbumTracks } from '../libraryQueueLoaders';

type AlbumCardProps = {
    album: MusicAlbum;
    onSelect: (album: MusicAlbum) => void;
};

export default function AlbumCard({ album, onSelect }: AlbumCardProps) {
    const { replaceQueue } = useGlobalMusic();

    const handlePlayAlbum = async (event: React.MouseEvent) => {
        event.stopPropagation();
        const tracks = await loadAlbumTracks(album.key);
        if (tracks.length > 0) {
            replaceQueue(tracks, 0, createAlbumPlaybackContext(album.album));
        }
    };

    return (
        <CollectionContextMenu
            collectionName={album.album}
            playbackContext={createAlbumPlaybackContext(album.album)}
            loadTracks={() => loadAlbumTracks(album.key)}
            layout="card"
        >
            <Card
                sx={{
                    bgcolor: 'background.paper',
                    transition: 'all 0.2s ease',
                    '&:hover': {
                        bgcolor: 'rgba(var(--app-color-ink-rgb), 0.04)',
                    },
                    '&:hover .play-overlay': {
                        opacity: 1,
                        transform: 'translateY(0)',
                    },
                }}
            >
                <CardActionArea
                    component="div"
                    role="button"
                    tabIndex={0}
                    onClick={() => onSelect(album)}
                    onKeyDown={(event) => handleKeyboardActivation(event, () => onSelect(album))}
                    sx={{ position: 'relative' }}
                >
                    <Box
                        sx={{
                            height: 140,
                            overflow: 'hidden',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            bgcolor: 'secondary.dark',
                            background: 'linear-gradient(135deg, #7c3aed 0%, #4f46e5 100%)',
                        }}
                    >
                        <CoverArt
                            src={getAlbumCoverUrl(album.key, 256)}
                            fallback={<Disc size={48} opacity={0.5} />}
                        />
                    </Box>
                    <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
                        <Typography variant="subtitle2" fontWeight={600} noWrap>
                            {album.album}
                        </Typography>
                        <Typography variant="caption" color="text.secondary" noWrap component="div">
                            {album.artist} {album.year ? `· ${album.year}` : ''}
                        </Typography>
                    </CardContent>
                    <IconButton
                        className="play-overlay"
                        onClick={(event) => void handlePlayAlbum(event)}
                        sx={{
                            position: 'absolute',
                            bottom: 50,
                            right: 8,
                            bgcolor: 'primary.main',
                            color: 'white',
                            width: 36,
                            height: 36,
                            opacity: 0,
                            transform: 'translateY(8px)',
                            transition: 'all 0.2s ease',
                            boxShadow: '0 4px 12px rgba(var(--app-color-primary-rgb), 0.4)',
                            '&:hover': {
                                bgcolor: 'primary.light',
                                transform: 'translateY(0) scale(1.05)',
                            },
                        }}
                    >
                        <Play size={16} fill="white" />
                    </IconButton>
                </CardActionArea>
            </Card>
        </CollectionContextMenu>
    );
}
