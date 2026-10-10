import { Check, EyeOff } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from '../videoContent.module.css';

type VideoWatchedToggleButtonProps = {
    isWatched: boolean;
    onToggle: () => void;
};

export default function VideoWatchedToggleButton({
    isWatched,
    onToggle,
}: VideoWatchedToggleButtonProps) {
    const { t } = useI18n();
    const label = isWatched ? t('VIDEO_MARK_UNWATCHED') : t('VIDEO_MARK_WATCHED');

    return (
        <button
            type="button"
            className={styles.iconBtn}
            aria-label={label}
            title={label}
            onClick={onToggle}
        >
            {isWatched ? <EyeOff size={14} /> : <Check size={14} />}
        </button>
    );
}
