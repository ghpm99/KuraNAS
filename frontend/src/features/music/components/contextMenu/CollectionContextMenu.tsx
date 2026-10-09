import { Box, IconButton } from '@mui/material';
import { ListEnd, ListStart, MoreVertical } from 'lucide-react';
import { useSnackbar } from 'notistack';
import type { ReactNode } from 'react';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import type { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import type { MusicPlaybackContext } from '@/features/music/components/playbackContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import MusicContextMenu from './MusicContextMenu';
import useMenuPosition from './useMenuPosition';

type CollectionContextMenuProps = {
    collectionName: string;
    playbackContext: MusicPlaybackContext;
    loadTracks: () => Promise<IMusicData[]>;
    layout: 'card' | 'row';
    children: ReactNode;
};

const ICON_SIZE = 18;

const buttonLayoutSx = {
    card: {
        top: 4,
        right: 4,
        color: 'white',
        bgcolor: 'rgba(0, 0, 0, 0.45)',
        '&:hover': { bgcolor: 'rgba(0, 0, 0, 0.65)' },
    },
    row: {
        top: '50%',
        right: 8,
        transform: 'translateY(-50%)',
        color: 'text.secondary',
        '&:hover': { color: 'text.primary' },
    },
} as const;

export default function CollectionContextMenu({
    collectionName,
    playbackContext,
    loadTracks,
    layout,
    children,
}: CollectionContextMenuProps) {
    const { playNext, addToQueue } = useGlobalMusic();
    const { enqueueSnackbar } = useSnackbar();
    const { t } = useI18n();
    const { position, openFromButton, openFromContextMenuEvent, close } = useMenuPosition();

    const enqueueCollection = async (
        enqueue: (tracks: IMusicData[], context?: MusicPlaybackContext) => void
    ) => {
        try {
            enqueue(await loadTracks(), playbackContext);
        } catch {
            enqueueSnackbar(t('MUSIC_COLLECTION_QUEUE_LOAD_FAILED'), { variant: 'warning' });
        }
    };

    return (
        <Box
            component={layout === 'row' ? 'li' : 'div'}
            sx={{ position: 'relative', listStyle: 'none' }}
            onContextMenu={openFromContextMenuEvent}
        >
            {children}
            <IconButton
                size="small"
                aria-label={t('MUSIC_COLLECTION_MORE_ACTIONS', { name: collectionName })}
                onClick={openFromButton}
                sx={{ position: 'absolute', ...buttonLayoutSx[layout] }}
            >
                <MoreVertical size={16} />
            </IconButton>
            <MusicContextMenu
                position={position}
                onClose={close}
                actions={[
                    {
                        key: 'play-next',
                        label: t('MUSIC_PLAY_NEXT'),
                        icon: <ListStart size={ICON_SIZE} />,
                        onSelect: () => void enqueueCollection(playNext),
                    },
                    {
                        key: 'add-to-queue',
                        label: t('MUSIC_ADD_TO_QUEUE'),
                        icon: <ListEnd size={ICON_SIZE} />,
                        onSelect: () => void enqueueCollection(addToQueue),
                    },
                ]}
            />
        </Box>
    );
}
