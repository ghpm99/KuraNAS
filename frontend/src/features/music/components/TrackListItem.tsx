import { Box, IconButton, ListItem, ListItemButton, Typography } from '@mui/material';
import { ListPlus, MoreVertical, Music, Pause, Play } from 'lucide-react';
import { useRef, useState } from 'react';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import {
    getMusicTitle,
    getMusicArtist,
    formatMusicDuration,
    getTrackDurationSeconds,
} from '@/utils/music';
import { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import { getTrackCoverUrl } from '@/service/musicCover';
import CoverArt from './CoverArt';
import useI18n from '@/components/i18n/provider/i18nContext';
import AddToPlaylistMenu from './AddToPlaylistMenu';
import TrackContextMenu from './contextMenu/TrackContextMenu';
import useMenuPosition from './contextMenu/useMenuPosition';

interface TrackListItemProps {
    track: IMusicData;
    index: number;
    onPlay: (track: IMusicData, index: number) => void;
    onAddToPlaylist?: (e: React.MouseEvent<HTMLElement>, fileId: number) => void;
    showArtist?: boolean;
}

const TrackListItem = ({
    track,
    index,
    onPlay,
    onAddToPlaylist,
    showArtist = true,
}: TrackListItemProps) => {
    const { currentTrack, isPlaying } = useGlobalMusic();
    const isCurrentTrack = currentTrack?.id === track.id;
    const duration = getTrackDurationSeconds(track.metadata);
    const trackTitle = getMusicTitle(track);
    const trackArtist = getMusicArtist(track);
    const { t } = useI18n();
    const rowRef = useRef<HTMLLIElement>(null);
    const { position, openFromButton, openFromContextMenuEvent, close } = useMenuPosition();
    const [hasOpenedPlaylistMenu, setHasOpenedPlaylistMenu] = useState(false);
    const [playlistMenuAnchor, setPlaylistMenuAnchor] = useState<HTMLElement | null>(null);

    const openPlaylistMenu = () => {
        setHasOpenedPlaylistMenu(true);
        setPlaylistMenuAnchor(rowRef.current);
    };

    return (
        <ListItem
            ref={rowRef}
            disablePadding
            sx={{
                px: 0,
                contentVisibility: 'auto',
                containIntrinsicBlockSize: 'auto var(--app-intrinsic-track-row-height)',
            }}
        >
            <ListItemButton
                onClick={() => onPlay(track, index)}
                onContextMenu={openFromContextMenuEvent}
                aria-label={`play ${trackTitle}`}
                sx={{
                    borderRadius: 1,
                    py: 0.5,
                    px: 1,
                    gap: 1.5,
                    bgcolor: isCurrentTrack
                        ? 'rgba(var(--app-color-primary-rgb), 0.08)'
                        : 'transparent',
                    '&:hover': {
                        bgcolor: isCurrentTrack
                            ? 'rgba(var(--app-color-primary-rgb), 0.12)'
                            : undefined,
                    },
                    '&:hover .track-index': { display: 'none' },
                    '&:hover .track-play-icon': { display: 'flex' },
                }}
            >
                <Box
                    sx={{
                        width: 32,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        flexShrink: 0,
                    }}
                >
                    {isCurrentTrack && isPlaying ? (
                        <Box
                            sx={{
                                display: 'flex',
                                alignItems: 'flex-end',
                                gap: '2px',
                                height: 16,
                            }}
                        >
                            {[1, 2, 3].map((bar) => (
                                <Box
                                    key={bar}
                                    sx={{
                                        width: 3,
                                        bgcolor: 'primary.main',
                                        borderRadius: 1,
                                        height: '10px',
                                        animation: `equalizer ${0.4 + bar * 0.15}s ease-in-out infinite alternate`,
                                        '@keyframes equalizer': {
                                            '0%': { height: '4px' },
                                            '100%': { height: '14px' },
                                        },
                                    }}
                                />
                            ))}
                        </Box>
                    ) : isCurrentTrack ? (
                        <Pause size={14} color="var(--app-color-primary)" />
                    ) : (
                        <>
                            <Typography
                                className="track-index"
                                variant="body2"
                                color="text.secondary"
                                sx={{ fontVariantNumeric: 'tabular-nums' }}
                            >
                                {index + 1}
                            </Typography>
                            <Box
                                className="track-play-icon"
                                sx={{ display: 'none', alignItems: 'center' }}
                            >
                                <Play size={14} />
                            </Box>
                        </>
                    )}
                </Box>

                <Box
                    sx={{
                        width: 36,
                        height: 36,
                        flexShrink: 0,
                        borderRadius: 0.5,
                        overflow: 'hidden',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        bgcolor: 'action.hover',
                    }}
                >
                    <CoverArt
                        src={getTrackCoverUrl(track.id, 96)}
                        fallback={<Music size={16} opacity={0.5} />}
                    />
                </Box>

                <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Typography
                        variant="body2"
                        noWrap
                        fontWeight={isCurrentTrack ? 600 : 400}
                        color={isCurrentTrack ? 'primary.main' : 'text.primary'}
                    >
                        {trackTitle}
                    </Typography>
                    {showArtist && (
                        <Typography variant="caption" color="text.secondary" noWrap component="div">
                            {trackArtist}
                        </Typography>
                    )}
                </Box>

                {onAddToPlaylist && (
                    <IconButton
                        size="small"
                        aria-label={`add ${trackTitle} to playlist`}
                        sx={{
                            color: 'text.secondary',
                            opacity: 0,
                            '.MuiListItemButton-root:hover &': { opacity: 1 },
                            '&:hover': { color: 'primary.main' },
                        }}
                        onClick={(e) => {
                            e.stopPropagation();
                            onAddToPlaylist(e, track.id);
                        }}
                    >
                        <ListPlus size={16} />
                    </IconButton>
                )}

                {duration ? (
                    <Typography
                        variant="caption"
                        color="text.secondary"
                        sx={{
                            flexShrink: 0,
                            fontVariantNumeric: 'tabular-nums',
                            minWidth: 36,
                            textAlign: 'right',
                        }}
                    >
                        {formatMusicDuration(duration)}
                    </Typography>
                ) : (
                    <Box sx={{ width: 36 }} />
                )}

                <IconButton
                    size="small"
                    aria-label={t('MUSIC_TRACK_MORE_ACTIONS', { name: trackTitle })}
                    aria-haspopup="menu"
                    onClick={openFromButton}
                    sx={{ color: 'text.secondary', '&:hover': { color: 'text.primary' } }}
                >
                    <MoreVertical size={16} />
                </IconButton>
            </ListItemButton>

            <TrackContextMenu
                track={track}
                position={position}
                onClose={close}
                onAddToPlaylist={openPlaylistMenu}
            />
            {hasOpenedPlaylistMenu && (
                <AddToPlaylistMenu
                    fileId={track.id}
                    anchorEl={playlistMenuAnchor}
                    onClose={() => setPlaylistMenuAnchor(null)}
                />
            )}
        </ListItem>
    );
};

export default TrackListItem;
