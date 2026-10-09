import { Disc, ListEnd, ListPlus, ListStart, User } from 'lucide-react';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import type { IMusicData } from '@/types/music';
import { getMusicRoute } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import MusicContextMenu, { type MusicContextMenuAction } from './MusicContextMenu';
import { getTrackAlbumKey, getTrackArtistKey } from './musicGroupingKeys';
import useOptionalNavigate from './useOptionalNavigate';
import type { MenuPosition } from './useMenuPosition';

type TrackContextMenuProps = {
    track: IMusicData;
    position: MenuPosition | null;
    onClose: () => void;
    onAddToPlaylist: () => void;
};

const ICON_SIZE = 18;

const buildMusicSectionRoute = (
    section: 'albums' | 'artists',
    queryParam: 'album' | 'artist',
    key: string
) => `${getMusicRoute(section)}?${new URLSearchParams({ [queryParam]: key }).toString()}`;

export default function TrackContextMenu({
    track,
    position,
    onClose,
    onAddToPlaylist,
}: TrackContextMenuProps) {
    const { playNext, addToQueue } = useGlobalMusic();
    const { t } = useI18n();
    const navigate = useOptionalNavigate();

    const albumKey = getTrackAlbumKey(track);
    const artistKey = getTrackArtistKey(track);

    const actions: MusicContextMenuAction[] = [
        {
            key: 'play-next',
            label: t('MUSIC_PLAY_NEXT'),
            icon: <ListStart size={ICON_SIZE} />,
            onSelect: () => playNext([track]),
        },
        {
            key: 'add-to-queue',
            label: t('MUSIC_ADD_TO_QUEUE'),
            icon: <ListEnd size={ICON_SIZE} />,
            onSelect: () => addToQueue([track]),
        },
        {
            key: 'add-to-playlist',
            label: t('MUSIC_ADD_TO_PLAYLIST'),
            icon: <ListPlus size={ICON_SIZE} />,
            onSelect: onAddToPlaylist,
        },
    ];
    if (albumKey !== '') {
        actions.push({
            key: 'go-to-album',
            label: t('MUSIC_GO_TO_ALBUM'),
            icon: <Disc size={ICON_SIZE} />,
            onSelect: () => navigate(buildMusicSectionRoute('albums', 'album', albumKey)),
        });
    }
    if (artistKey !== '') {
        actions.push({
            key: 'go-to-artist',
            label: t('MUSIC_GO_TO_ARTIST'),
            icon: <User size={ICON_SIZE} />,
            onSelect: () => navigate(buildMusicSectionRoute('artists', 'artist', artistKey)),
        });
    }

    return <MusicContextMenu position={position} actions={actions} onClose={onClose} />;
}
