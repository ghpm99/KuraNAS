import { TextField, Typography } from '@mui/material';
import { ArrowDown, ArrowLeft, ArrowUp, Play, Trash2 } from 'lucide-react';
import { useMemo, useState } from 'react';
import { VideoPlaylistDto } from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import styles from '../videoContent.module.css';

type VideoPlaylistDetailViewProps = {
	playlist: VideoPlaylistDto;
	isRenaming: boolean;
	isRemoving: boolean;
	isReordering: boolean;
	onBack: () => void;
	onOpenVideo: (videoId: number) => void;
	onRename: (name: string) => void;
	onRemoveVideo: (videoId: number) => void;
	onMoveItem: (index: number, direction: -1 | 1) => void;
};

const apiBase = `${getApiV1BaseUrl()}/files`;

export default function VideoPlaylistDetailView({
	playlist,
	isRenaming,
	isRemoving,
	isReordering,
	onBack,
	onOpenVideo,
	onRename,
	onRemoveVideo,
	onMoveItem,
}: VideoPlaylistDetailViewProps) {
	const { t } = useI18n();
	const [nameDraft, setNameDraft] = useState(playlist.name);
	const coverId = playlist.cover_video_id || playlist.items[0]?.video.id;

	const orderedItems = useMemo(
		() => [...playlist.items].sort((a, b) => a.order_index - b.order_index || a.id - b.id),
		[playlist.items],
	);

	return (
		<div className={styles.page}>
			<section className={styles.hero}>
				{coverId && (
					<img className={styles.heroImage} src={`${apiBase}/video-thumbnail/${coverId}?width=1280&height=720`} alt={playlist.name} />
				)}
				<div className={styles.heroShade} />
				<div className={styles.heroContent}>
					<button type='button' className={styles.backBtn} onClick={onBack}>
						<ArrowLeft size={16} />
						<span>{t('VIDEO_BACK_TO_VIDEOS')}</span>
					</button>
					<p className={styles.heroEyebrow}>{(playlist.classification || 'personal').toUpperCase()}</p>
					<h1 className={styles.heroTitle}>{playlist.name}</h1>
					<p className={styles.heroMeta}>{t('VIDEO_PLAYLIST_META', { count: String(playlist.item_count) })}</p>
				</div>
			</section>

			<section className={styles.managementPanel}>
				<div className={styles.panelHeader}>
					<Typography variant='h6'>{t('VIDEO_EDIT_PLAYLIST')}</Typography>
				</div>
				<div className={styles.renameRow}>
					<TextField
						size='small'
						value={nameDraft}
						onChange={(event) => setNameDraft(event.target.value)}
						placeholder={t('VIDEO_DISPLAY_NAME_PLACEHOLDER')}
					/>
					<button
						type='button'
						className={styles.actionBtn}
						onClick={() => onRename(nameDraft)}
						disabled={isRenaming || !nameDraft.trim()}
					>
						{t('VIDEO_SAVE_NAME')}
					</button>
				</div>
			</section>

			<section className={styles.detailList}>
				{orderedItems.map((item, index) => (
					<div key={item.video.id} className={styles.detailItem}>
						<button type='button' className={styles.detailMainButton} onClick={() => onOpenVideo(item.video.id)}>
							<div className={styles.detailThumb}>
								<img
									loading='lazy'
									src={`${apiBase}/video-thumbnail/${item.video.id}?width=320&height=180`}
									alt={item.video.name}
								/>
							</div>
							<div className={styles.detailMeta}>
								<Typography className={styles.detailTitle}>{item.video.name}</Typography>
								<Typography className={styles.detailSub}>
									{item.video.format.toUpperCase()} ·{' '}
									{item.source_kind === 'manual' ? t('VIDEO_SOURCE_MANUAL') : t('VIDEO_SOURCE_AUTO')}
								</Typography>
							</div>
							<div className={styles.detailPlay}>
								<Play size={16} />
							</div>
						</button>
						<div className={styles.itemActions}>
							<button
								type='button'
								className={styles.iconBtn}
								onClick={() => onMoveItem(index, -1)}
								disabled={index === 0 || isReordering}
							>
								<ArrowUp size={14} />
							</button>
							<button
								type='button'
								className={styles.iconBtn}
								onClick={() => onMoveItem(index, 1)}
								disabled={index === orderedItems.length - 1 || isReordering}
							>
								<ArrowDown size={14} />
							</button>
							<button
								type='button'
								className={styles.iconBtnDanger}
								onClick={() => onRemoveVideo(item.video.id)}
								disabled={isRemoving}
							>
								<Trash2 size={14} />
							</button>
						</div>
					</div>
				))}
			</section>
		</div>
	);
}
