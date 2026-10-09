import { Box, IconButton, Typography } from '@mui/material';
import { ArrowLeft, Play, Shuffle } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatTotalDuration } from '@/utils/music';

interface CategoryHeaderProps {
    title: string;
    subtitle?: React.ReactNode;
    trackCount?: number;
    totalLengthSeconds?: number;
    iconSize?: number;
    icon: React.ReactNode;
    gradientFrom?: string;
    gradientTo?: string;
    onBack: () => void;
    onPlayAll: () => void;
    onShuffleAll: () => void;
    actions?: React.ReactNode;
}

const DEFAULT_ICON_SIZE = 120;
const COMPACT_ICON_MAX_SIZE = 160;

const CategoryHeader = ({
    title,
    subtitle,
    trackCount,
    totalLengthSeconds,
    iconSize = DEFAULT_ICON_SIZE,
    icon,
    gradientFrom = '#4f46e5',
    gradientTo = '#1a1a24',
    onBack,
    onPlayAll,
    onShuffleAll,
    actions,
}: CategoryHeaderProps) => {
    const { t } = useI18n();

    return (
        <Box
            sx={{
                background: `linear-gradient(180deg, ${gradientFrom}33 0%, ${gradientTo} 100%)`,
                borderRadius: 3,
                p: 2.5,
                mb: 2,
            }}
        >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                <IconButton
                    onClick={onBack}
                    size="small"
                    aria-label={t('MUSIC_GROUP_BACK')}
                    sx={{ color: 'text.primary' }}
                >
                    <ArrowLeft size={20} />
                </IconButton>
            </Box>

            <Box
                sx={{
                    display: 'flex',
                    flexDirection: {
                        xs: iconSize > DEFAULT_ICON_SIZE ? 'column' : 'row',
                        sm: 'row',
                    },
                    alignItems: { xs: 'flex-start', sm: 'flex-end' },
                    gap: 2.5,
                }}
            >
                <Box
                    sx={{
                        width: { xs: Math.min(iconSize, COMPACT_ICON_MAX_SIZE), sm: iconSize },
                        height: { xs: Math.min(iconSize, COMPACT_ICON_MAX_SIZE), sm: iconSize },
                        overflow: 'hidden',
                        borderRadius: 2,
                        bgcolor: `${gradientFrom}44`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        flexShrink: 0,
                        boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
                    }}
                >
                    {icon}
                </Box>
                <Box sx={{ minWidth: 0, width: { xs: '100%', sm: 'auto' }, flex: { sm: 1 } }}>
                    <Typography variant="h5" fontWeight={700} noWrap sx={{ mb: 0.5 }}>
                        {title}
                    </Typography>
                    {subtitle && (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            noWrap
                            component="div"
                            sx={{ mb: 0.5 }}
                        >
                            {subtitle}
                        </Typography>
                    )}
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', columnGap: 1 }}>
                        {trackCount !== undefined && (
                            <Typography variant="caption" color="text.secondary">
                                {trackCount} {t('MUSIC_TRACKS_COUNT')}
                            </Typography>
                        )}
                        {totalLengthSeconds !== undefined && totalLengthSeconds > 0 && (
                            <Typography variant="caption" color="text.secondary">
                                {trackCount !== undefined ? '· ' : ''}
                                {formatTotalDuration(totalLengthSeconds, t)}
                            </Typography>
                        )}
                    </Box>
                </Box>
            </Box>

            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mt: 2 }}>
                <IconButton
                    onClick={onPlayAll}
                    aria-label={t('MUSIC_PLAY_ALL')}
                    sx={{
                        bgcolor: 'primary.main',
                        color: 'white',
                        width: 44,
                        height: 44,
                        '&:hover': { bgcolor: 'primary.light', transform: 'scale(1.05)' },
                        transition: 'all 0.2s ease',
                    }}
                >
                    <Play size={22} fill="white" />
                </IconButton>
                <IconButton
                    onClick={onShuffleAll}
                    aria-label={t('MUSIC_SHUFFLE_ALL')}
                    sx={{
                        color: 'text.secondary',
                        '&:hover': { color: 'text.primary' },
                    }}
                >
                    <Shuffle size={20} />
                </IconButton>
                {actions}
            </Box>
        </Box>
    );
};

export default CategoryHeader;
