import { Button } from '@mui/material';
import { Play } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { type VideoCatalogItemDto } from '@/service/videoPlayback';
import styles from '../videoContent.module.css';

type VideoCatalogRailProps = {
	titleKey: string;
	descriptionKey: string;
	items: VideoCatalogItemDto[];
	onPlayVideo: (videoId: number, playlistId?: number | null) => void;
};

const apiBase = `${getApiV1BaseUrl()}/files`;

const getStatusKey = (status: VideoCatalogItemDto['status']) => {
	switch (status) {
		case 'completed':
			return 'VIDEO_STATUS_COMPLETED';
		case 'in_progress':
			return 'VIDEO_STATUS_IN_PROGRESS';
		default:
			return 'VIDEO_STATUS_NOT_STARTED';
	}
};

export default function VideoCatalogRail({ titleKey, descriptionKey, items, onPlayVideo }: VideoCatalogRailProps) {
	const { t } = useI18n();

	if (items.length === 0) {
		return null;
	}

	return (
		<section className={styles.sectionBlock}>
			<div className={styles.sectionHeader}>
				<h2>{t(titleKey)}</h2>
				<p>{t(descriptionKey)}</p>
			</div>
			<div className={styles.catalogRail}>
				{items.map((item) => (
					<article key={item.video.id} className={styles.catalogCard}>
						<div className={styles.catalogThumb}>
							<img
								loading='lazy'
								src={`${apiBase}/video-thumbnail/${item.video.id}?width=480&height=270`}
								alt={item.video.name}
							/>
						</div>
						<div className={styles.catalogBody}>
							<div className={styles.catalogMeta}>
								<span className={styles.statusBadge}>{t(getStatusKey(item.status))}</span>
								<span className={styles.catalogPath}>{item.video.parent_path}</span>
							</div>
							<h3 className={styles.catalogTitle}>{item.video.name}</h3>
							<div className={styles.catalogFooter}>
								<p className={styles.catalogFormat}>{item.video.format.toUpperCase()}</p>
								<Button
									variant='contained'
									size='small'
									startIcon={<Play size={14} />}
									onClick={() => onPlayVideo(item.video.id, null)}
								>
									{t('VIDEO_PLAY')}
								</Button>
							</div>
						</div>
					</article>
				))}
			</div>
		</section>
	);
}
