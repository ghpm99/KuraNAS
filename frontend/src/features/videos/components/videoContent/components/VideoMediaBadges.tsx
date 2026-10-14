import useI18n from '@/components/i18n/provider/i18nContext';
import type { VideoFileDto, VideoLibraryResolution } from '@/service/videoPlayback';
import { classifyVideoResolution } from '../../../providers/videoContentProvider/videoResolutionClass';
import { formatVideoDuration } from './formatVideoDuration';
import styles from './VideoMediaBadges.module.css';

type VideoMediaBadgesProps = {
    video?: Pick<VideoFileDto, 'duration_seconds' | 'height'> | null;
};

const resolutionBadgeKeys: Record<VideoLibraryResolution, string> = {
    sd: 'VIDEO_BADGE_RESOLUTION_SD',
    hd: 'VIDEO_BADGE_RESOLUTION_HD',
    fhd: 'VIDEO_BADGE_RESOLUTION_FHD',
    uhd: 'VIDEO_BADGE_RESOLUTION_UHD',
};

export default function VideoMediaBadges({ video }: VideoMediaBadgesProps) {
    const { t } = useI18n();
    const durationLabel = formatVideoDuration(video?.duration_seconds);
    const resolutionClass = classifyVideoResolution(video?.height);

    return (
        <>
            {resolutionClass && (
                <span className={`${styles.badge} ${styles.resolutionBadge}`}>
                    {t(resolutionBadgeKeys[resolutionClass])}
                </span>
            )}
            {durationLabel && (
                <span className={`${styles.badge} ${styles.durationBadge}`}>{durationLabel}</span>
            )}
        </>
    );
}
