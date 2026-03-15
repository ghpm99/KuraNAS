import { CircularProgress } from '@mui/material';
import { CalendarDays } from 'lucide-react';
import { formatSize } from '@/utils';
import type { IImageData } from '@/components/providers/imageProvider/imageProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import styles from '../ImageContent.module.css';

const thumbnailWidth = 960;
const thumbnailHeight = 720;

const thumbnailUrl = (id: number) =>
	`${getApiV1BaseUrl()}/files/thumbnail/${id}?width=${thumbnailWidth}&height=${thumbnailHeight}`;

type ImageGroup = { label: string; items: IImageData[] };

type ImageGroupsGridProps = {
	groups: ImageGroup[];
	totalImages: number;
	isFetchingNextPage: boolean;
	hasNextPage: boolean;
	lastVisibleImageId?: number;
	loadMoreRef: (node: HTMLButtonElement | null) => void;
	onOpenImage: (id: number) => void;
};

const imageMetadataSummary = (image: IImageData): string => {
	const format = image.format ? `${image.format} - ` : '';
	return `${format}${formatSize(image.size)}`;
};

export default function ImageGroupsGrid({
	groups,
	totalImages,
	isFetchingNextPage,
	hasNextPage,
	lastVisibleImageId,
	loadMoreRef,
	onOpenImage,
}: ImageGroupsGridProps) {
	const { t } = useI18n();

	if (groups.length === 0 && !isFetchingNextPage) {
		return (
			<div className={styles.empty}>
				<h3>{t('IMAGES_EMPTY_TITLE')}</h3>
				<p>{t('IMAGES_EMPTY_DESC')}</p>
			</div>
		);
	}

	return (
		<>
			<div className={styles.sections}>
				{groups.map((group) => (
					<section key={group.label} className={styles.group}>
						<header className={styles.groupHeader}>
							<CalendarDays size={16} />
							<h3>{group.label}</h3>
							<span>{t('IMAGES_PHOTOS_COUNT', { count: String(group.items.length) })}</span>
						</header>
						<div className={styles.grid}>
							{group.items.map((item) => {
								const width = item.metadata?.width ?? 1;
								const height = item.metadata?.height ?? 1;
								const orientationClass = height > width ? styles.portrait : styles.landscape;
								const ref = item.id === lastVisibleImageId ? loadMoreRef : undefined;

								return (
									<button
										type='button'
										key={item.id}
										ref={ref}
										className={`${styles.photoCard} ${orientationClass}`}
										onClick={() => onOpenImage(item.id)}
										aria-label={t('IMAGES_OPEN_IMAGE_ARIA', { name: item.name })}
									>
										<img className={styles.thumbnail} src={thumbnailUrl(item.id)} alt={item.name} loading='lazy' />
										<div className={styles.photoOverlay}>
											<p>{item.name}</p>
											<span>{imageMetadataSummary(item)}</span>
										</div>
									</button>
								);
							})}
						</div>
					</section>
				))}
			</div>
			{isFetchingNextPage && (
				<div className={styles.loadingIndicator}>
					<CircularProgress size={40} />
				</div>
			)}
			{!hasNextPage && totalImages > 0 && <div className={styles.endMessage}>{t('IMAGES_END_MESSAGE')}</div>}
		</>
	);
}
