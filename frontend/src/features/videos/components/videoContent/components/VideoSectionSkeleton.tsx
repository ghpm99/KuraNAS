import { Skeleton } from '@mui/material';
import styles from '../videoContent.module.css';

type VideoSectionSkeletonProps = {
    cardCount?: number;
    layout?: 'playlistGrid' | 'catalogRail';
};

export default function VideoSectionSkeleton({
    cardCount = 4,
    layout = 'playlistGrid',
}: VideoSectionSkeletonProps) {
    const cardsClassName = layout === 'catalogRail' ? styles.catalogRail : styles.gridCards;
    const skeletonCardKeys = Array.from({ length: cardCount }, (_, position) => position);

    return (
        <section className={styles.sectionBlock} data-testid="video-section-skeleton">
            <div className={styles.sectionHeader}>
                <Skeleton variant="text" width={220} height={32} />
                <Skeleton variant="text" width={320} height={20} />
            </div>
            <div className={cardsClassName}>
                {skeletonCardKeys.map((cardKey) => (
                    <Skeleton key={cardKey} variant="rounded" height={200} animation="wave" />
                ))}
            </div>
        </section>
    );
}
