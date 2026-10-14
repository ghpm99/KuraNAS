import { useEffect, useRef } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getVideoThumbnailUrl } from '@/service/videoPlayback';
import styles from './UpNextAutoplayCard.module.css';

const THUMBNAIL_WIDTH = 320;
const THUMBNAIL_HEIGHT = 180;

type UpNextAutoplayCardProps = {
    nextVideoId: number;
    nextVideoTitle: string;
    isCounting: boolean;
    secondsRemaining: number;
    isAutoplayEnabled: boolean;
    onPlayNow: () => void;
    onCancel: () => void;
    onAutoplayEnabledChange: (isEnabled: boolean) => void;
};

const isUserTyping = (): boolean => {
    const focusedElement = document.activeElement;
    if (!(focusedElement instanceof HTMLElement)) return false;
    return (
        ['INPUT', 'TEXTAREA', 'SELECT'].includes(focusedElement.tagName) ||
        focusedElement.isContentEditable ||
        ['', 'true'].includes(focusedElement.getAttribute('contenteditable') ?? 'false')
    );
};

const UpNextAutoplayCard = ({
    nextVideoId,
    nextVideoTitle,
    isCounting,
    secondsRemaining,
    isAutoplayEnabled,
    onPlayNow,
    onCancel,
    onAutoplayEnabledChange,
}: UpNextAutoplayCardProps) => {
    const { t } = useI18n();
    const playNowButtonRef = useRef<HTMLButtonElement>(null);

    useEffect(() => {
        if (!isUserTyping()) playNowButtonRef.current?.focus();
    }, []);

    useEffect(() => {
        const cancelOnEscape = (event: KeyboardEvent) => {
            if (event.key === 'Escape') onCancel();
        };
        document.addEventListener('keydown', cancelOnEscape);
        return () => document.removeEventListener('keydown', cancelOnEscape);
    }, [onCancel]);

    return (
        <div className={styles.card} role="group" aria-label={t('VIDEO_PLAYER_UP_NEXT')}>
            <img
                className={styles.thumbnail}
                src={getVideoThumbnailUrl(nextVideoId, THUMBNAIL_WIDTH, THUMBNAIL_HEIGHT)}
                alt=""
            />
            <div className={styles.body}>
                <span className={styles.eyebrow}>{t('VIDEO_PLAYER_UP_NEXT')}</span>
                <strong className={styles.title}>{nextVideoTitle}</strong>
                <p className={styles.countdown} role="status" aria-live="polite">
                    {isCounting
                        ? t('VIDEO_UP_NEXT_COUNTDOWN', { seconds: String(secondsRemaining) })
                        : ''}
                </p>
                <div className={styles.actions}>
                    <button
                        ref={playNowButtonRef}
                        type="button"
                        className={styles.primaryAction}
                        onClick={onPlayNow}
                    >
                        {t('VIDEO_UP_NEXT_PLAY_NOW')}
                    </button>
                    <button type="button" className={styles.secondaryAction} onClick={onCancel}>
                        {t('VIDEO_UP_NEXT_CANCEL')}
                    </button>
                </div>
                <label className={styles.autoplayToggle}>
                    <input
                        type="checkbox"
                        checked={isAutoplayEnabled}
                        onChange={(event) => onAutoplayEnabledChange(event.target.checked)}
                    />
                    <span>{t('VIDEO_UP_NEXT_AUTOPLAY_TOGGLE')}</span>
                </label>
            </div>
        </div>
    );
};

export default UpNextAutoplayCard;
