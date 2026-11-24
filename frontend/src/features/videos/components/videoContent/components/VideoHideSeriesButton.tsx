import { EyeOff } from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import VideoConfirmDialog from './VideoConfirmDialog';
import styles from '../videoContent.module.css';

type VideoHideSeriesButtonProps = {
    seriesName: string;
    isChanging: boolean;
    onHide?: () => void;
};

export default function VideoHideSeriesButton({
    seriesName,
    isChanging,
    onHide,
}: VideoHideSeriesButtonProps) {
    const { t } = useI18n();
    const [isConfirmOpen, setIsConfirmOpen] = useState(false);

    if (!onHide) return null;

    return (
        <>
            <button
                type="button"
                className={styles.actionBtn}
                disabled={isChanging}
                onClick={() => setIsConfirmOpen(true)}
            >
                <EyeOff size={14} />
                {t('VIDEO_HIDE_SERIES')}
            </button>
            <VideoConfirmDialog
                isOpen={isConfirmOpen}
                title={t('VIDEO_HIDE_SERIES_CONFIRM_TITLE')}
                message={t('VIDEO_HIDE_SERIES_CONFIRM_MESSAGE', { name: seriesName })}
                confirmLabel={t('VIDEO_HIDE_SERIES')}
                onConfirm={onHide}
                onClose={() => setIsConfirmOpen(false)}
            />
        </>
    );
}
