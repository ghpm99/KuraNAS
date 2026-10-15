import { Star } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoStarToggle } from '../hooks/useVideoStarToggle/useVideoStarToggle';
import styles from './VideoStarButton.module.css';

type VideoStarButtonProps = {
    videoId: number;
    videoName: string;
    isStarred?: boolean;
    placement?: 'inline' | 'overlay';
};

export default function VideoStarButton({
    videoId,
    videoName,
    isStarred = false,
    placement = 'inline',
}: VideoStarButtonProps) {
    const { t } = useI18n();
    const { toggleStar } = useVideoStarToggle();
    const label = t(isStarred ? 'FILES_UNFAVORITE_ITEM' : 'FILES_FAVORITE_ITEM', {
        name: videoName,
    });

    return (
        <button
            type="button"
            className={placement === 'overlay' ? styles.overlayStar : styles.inlineStar}
            aria-label={label}
            title={label}
            aria-pressed={isStarred}
            onClick={(event) => {
                event.stopPropagation();
                toggleStar(videoId, isStarred);
            }}
        >
            <Star size={14} fill={isStarred ? 'currentColor' : 'none'} />
        </button>
    );
}
