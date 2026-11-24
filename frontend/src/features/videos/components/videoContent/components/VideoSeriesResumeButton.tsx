import { Play } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { VideoSeriesResumeDto } from '@/service/videoPlayback';
import styles from '../videoContent.module.css';

type VideoSeriesResumeButtonProps = {
    resume: VideoSeriesResumeDto | null;
    onOpenVideo: (videoId: number) => void;
};

export default function VideoSeriesResumeButton({
    resume,
    onOpenVideo,
}: VideoSeriesResumeButtonProps) {
    const { t } = useI18n();

    const resolveLabel = (currentResume: VideoSeriesResumeDto): string => {
        if (currentResume.mode === 'rewatch') {
            return t('VIDEO_SERIES_REWATCH');
        }
        if (currentResume.mode === 'start') {
            return t('VIDEO_SERIES_WATCH');
        }
        if (currentResume.season == null || currentResume.episode == null) {
            return t('VIDEO_DETAIL_RESUME_ACTION');
        }
        return t('VIDEO_SERIES_CONTINUE', {
            season: String(currentResume.season),
            episode: String(currentResume.episode),
        });
    };

    return (
        <button
            type="button"
            className={styles.actionBtnPrimary}
            onClick={() => resume && onOpenVideo(resume.video_id)}
            disabled={!resume}
        >
            <Play size={16} />
            <span>{resume ? resolveLabel(resume) : t('VIDEO_SERIES_WATCH')}</span>
        </button>
    );
}
