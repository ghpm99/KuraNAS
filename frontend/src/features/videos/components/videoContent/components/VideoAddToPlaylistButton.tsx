import { useState } from 'react';
import { ListPlus } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import VideoAddToPlaylistMenu from './VideoAddToPlaylistMenu';
import styles from '../videoContent.module.css';

type VideoAddToPlaylistButtonProps = {
    videoId: number;
    isAddingToPlaylist: boolean;
    onAddVideo: (videoId: number, playlistId: number) => void;
};

export default function VideoAddToPlaylistButton({
    videoId,
    isAddingToPlaylist,
    onAddVideo,
}: VideoAddToPlaylistButtonProps) {
    const { t } = useI18n();
    const [menuAnchor, setMenuAnchor] = useState<HTMLElement | null>(null);

    return (
        <>
            <button
                type="button"
                className={styles.actionBtn}
                disabled={isAddingToPlaylist}
                onClick={(event) => setMenuAnchor(event.currentTarget)}
            >
                <ListPlus size={14} />
                {t('VIDEO_ADD_TO_PLAYLIST')}
            </button>
            <VideoAddToPlaylistMenu
                videoId={videoId}
                anchorEl={menuAnchor}
                onClose={() => setMenuAnchor(null)}
                onSelectPlaylist={(playlistId) => {
                    setMenuAnchor(null);
                    onAddVideo(videoId, playlistId);
                }}
            />
        </>
    );
}
