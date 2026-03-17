import { Box, IconButton, ListItem, ListItemButton, Typography } from '@mui/material';
import { ListPlus, Pause, Play } from 'lucide-react';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import { getMusicTitle, getMusicArtist, formatMusicDuration } from '@/utils/music';
import { IMusicData } from '@/components/providers/musicProvider/musicProvider';

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
    const duration = track.metadata?.duration;
    const trackTitle = getMusicTitle(track);
    const trackArtist = getMusicArtist(track);

    return (
        <ListItem disablePadding sx={{ px: 0 }}>
            <ListItemButton
                onClick={() => onPlay(track, index)}
                aria-label={`play ${trackTitle}`}
                sx={{
                    borderRadius: 1,
                    py: 0.5,
                    px: 1,
                    gap: 1.5,
                    bgcolor: isCurrentTrack ? 'rgba(99, 102, 241, 0.08)' : 'transparent',
                    '&:hover': {
                        bgcolor: isCurrentTrack ? 'rgba(99, 102, 241, 0.12)' : undefined,
                    },
                    '&:hover .track-index': { display: 'none' },
                    '&:hover .track-play-icon': { display: 'flex' },
                }}
            >
                {/* Track number / playing indicator */}
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
                        <Pause size={14} color="#6366f1" />
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

                {/* Track info */}
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

                {/* Actions */}
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

                {/* Duration */}
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
            </ListItemButton>
        </ListItem>
    );
};

export default TrackListItem;
