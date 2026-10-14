import { EyeOff } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from '../videoContent.module.css';

type VideoHidePlaylistButtonProps = {
    isChanging: boolean;
    onHide?: () => void;
};

export default function VideoHidePlaylistButton({
    isChanging,
    onHide,
}: VideoHidePlaylistButtonProps) {
    const { t } = useI18n();

    if (!onHide) return null;

    return (
        <button type="button" className={styles.actionBtn} disabled={isChanging} onClick={onHide}>
            <EyeOff size={14} />
            {t('VIDEO_HIDE_PLAYLIST')}
        </button>
    );
}
