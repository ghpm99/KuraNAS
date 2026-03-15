import { Play } from 'lucide-react';
import { Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { type VideoDetailItem } from '../useVideoPlaylistDetail';
import styles from '../videoContent.module.css';

type VideoDetailListItemProps = {
	item: VideoDetailItem;
	onOpenVideo: (videoId: number) => void;
};

const apiBase = `${getApiV1BaseUrl()}/files`;

const statusKeyMap = {
	not_started: 'VIDEO_STATUS_NOT_STARTED',
	in_progress: 'VIDEO_STATUS_IN_PROGRESS',
	completed: 'VIDEO_STATUS_COMPLETED',
} as const;

export default function VideoDetailListItem({ item, onOpenVideo }: VideoDetailListItemProps) {
	const { t } = useI18n();

	return (
		<button type='button' className={styles.detailListItem} onClick={() => onOpenVideo(item.video.id)}>
			<div className={styles.detailListThumb}>
				<img
					loading='lazy'
					src={`${apiBase}/video-thumbnail/${item.video.id}?width=320&height=180`}
					alt={item.video.name}
				/>
			</div>
			<div className={styles.detailListBody}>
				<div className={styles.detailListHeader}>
					{item.sequenceLabel && <span className={styles.sequenceTag}>{item.sequenceLabel}</span>}
					<span className={styles.statusBadge}>{t(statusKeyMap[item.status])}</span>
				</div>
				<Typography className={styles.detailListTitle}>{item.displayTitle || item.video.name}</Typography>
				<Typography className={styles.detailListMeta}>
					{item.video.format.toUpperCase()} · {item.source_kind === 'manual' ? t('VIDEO_SOURCE_MANUAL') : t('VIDEO_SOURCE_AUTO')}
				</Typography>
				{item.progress_pct > 0 && item.status !== 'not_started' && (
					<div className={styles.progressTrack} aria-hidden='true'>
						<div className={styles.progressFill} style={{ width: `${item.progress_pct}%` }} />
					</div>
				)}
			</div>
			<div className={styles.detailPlay}>
				<Play size={16} />
			</div>
		</button>
	);
}
