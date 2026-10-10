import ErrorState from '@/components/errorState/errorState';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { VideoQueryFailure } from '@/features/videos/providers/videoContentProvider/videoQueryFailure';
import styles from '../videoContent.module.css';

type VideoSectionErrorProps = {
    sectionTitleKey: string;
    failure: VideoQueryFailure;
};

export default function VideoSectionError({ sectionTitleKey, failure }: VideoSectionErrorProps) {
    const { t } = useI18n();

    return (
        <section className={styles.sectionBlock}>
            <ErrorState
                title={t('VIDEO_SECTION_LOAD_ERROR', { section: t(sectionTitleKey) })}
                backendMessage={failure.message}
                onRetry={failure.retry}
            />
        </section>
    );
}
