import { ChevronLeft, ChevronRight, Expand, Info, Minus, Plus, X } from 'lucide-react';
import { formatSize } from '@/utils';
import type { IImageData } from '@/components/providers/imageProvider/imageProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';

const thumbnailWidth = 960;
const thumbnailHeight = 720;

const thumbnailUrl = (id: number) =>
	`${getApiV1BaseUrl()}/files/thumbnail/${id}?width=${thumbnailWidth}&height=${thumbnailHeight}`;
const blobUrl = (id: number) => `${getApiV1BaseUrl()}/files/blob/${id}`;

type ImageViewerModalProps = {
	activeImage: IImageData;
	activeIndex: number;
	activeImageDate: Date | null;
	dateFormatter: Intl.DateTimeFormat;
	filteredImages: IImageData[];
	zoom: number;
	showDetails: boolean;
	onToggleDetails: () => void;
	onDecreaseZoom: () => void;
	onResetZoom: () => void;
	onIncreaseZoom: () => void;
	onClose: () => void;
	onPrevious: () => void;
	onNext: () => void;
	onOpenImage: (id: number) => void;
};

export default function ImageViewerModal({
	activeImage,
	activeIndex,
	activeImageDate,
	dateFormatter,
	filteredImages,
	zoom,
	showDetails,
	onToggleDetails,
	onDecreaseZoom,
	onResetZoom,
	onIncreaseZoom,
	onClose,
	onPrevious,
	onNext,
	onOpenImage,
}: ImageViewerModalProps) {
	const { t } = useI18n();

	return (
		<div className='image-viewer-overlay' role='dialog' aria-modal='true'>
			<div className='image-viewer-topbar'>
				<div>
					<strong>{activeImage.name}</strong>
					<p>{activeImageDate ? dateFormatter.format(activeImageDate) : t('IMAGES_DATE_UNAVAILABLE')}</p>
				</div>
				<div className='viewer-actions'>
					<button type='button' onClick={onToggleDetails} aria-label={t('IMAGES_TOGGLE_DETAILS')}>
						<Info size={16} />
					</button>
					<button type='button' onClick={onDecreaseZoom} aria-label={t('IMAGES_DECREASE_ZOOM')}>
						<Minus size={16} />
					</button>
					<button type='button' onClick={onResetZoom} aria-label={t('IMAGES_RESET_ZOOM')}>
						<Expand size={16} />
					</button>
					<button type='button' onClick={onIncreaseZoom} aria-label={t('IMAGES_INCREASE_ZOOM')}>
						<Plus size={16} />
					</button>
					<button type='button' onClick={onClose} aria-label={t('IMAGES_CLOSE_VIEWER')}>
						<X size={16} />
					</button>
				</div>
			</div>

			<div
				className='image-viewer-stage'
				onWheel={(event) => {
					event.preventDefault();
					if (event.deltaY < 0) onIncreaseZoom();
					if (event.deltaY > 0) onDecreaseZoom();
				}}
			>
				<button type='button' className='viewer-nav left' onClick={onPrevious} aria-label={t('IMAGES_PREVIOUS')}>
					<ChevronLeft size={24} />
				</button>
				<img src={blobUrl(activeImage.id)} alt={activeImage.name} className='viewer-image' style={{ transform: `scale(${zoom})` }} />
				<button type='button' className='viewer-nav right' onClick={onNext} aria-label={t('IMAGES_NEXT')}>
					<ChevronRight size={24} />
				</button>
			</div>

			<div className='image-viewer-bottom'>
				<span>
					{t('IMAGES_ZOOM_LABEL')}: {Math.round(zoom * 100)}%
				</span>
				<span>
					{activeIndex + 1} / {filteredImages.length}
				</span>
			</div>

			{showDetails && (
				<aside className='image-viewer-details'>
					<h4>{t('IMAGES_DETAILS_TITLE')}</h4>
					<p>
						<strong>{t('IMAGES_DETAIL_NAME')}:</strong> {activeImage.name}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_FOLDER')}:</strong> {activeImage.path}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_FORMAT')}:</strong> {activeImage.format || t('COMMON_NOT_AVAILABLE')}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_SIZE')}:</strong> {formatSize(activeImage.size)}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_DIMENSIONS')}:</strong>{' '}
						{activeImage.metadata?.width && activeImage.metadata?.height
							? `${activeImage.metadata.width}x${activeImage.metadata.height}`
							: t('COMMON_NOT_AVAILABLE')}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_CAMERA')}:</strong>{' '}
						{activeImage.metadata?.make || activeImage.metadata?.model || t('COMMON_NOT_AVAILABLE')}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_LENS')}:</strong> {activeImage.metadata?.lens_model || t('COMMON_NOT_AVAILABLE')}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_ISO')}:</strong> {activeImage.metadata?.iso || t('COMMON_NOT_AVAILABLE')}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_FOCAL')}:</strong> {activeImage.metadata?.focal_length || t('COMMON_NOT_AVAILABLE')}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_APERTURE')}:</strong> {activeImage.metadata?.f_number || t('COMMON_NOT_AVAILABLE')}
					</p>
					<p>
						<strong>{t('IMAGES_DETAIL_EXPOSURE')}:</strong> {activeImage.metadata?.exposure_time || t('COMMON_NOT_AVAILABLE')}
					</p>
				</aside>
			)}

			<div className='viewer-filmstrip'>
				{filteredImages.slice(Math.max(0, activeIndex - 8), activeIndex + 9).map((item) => (
					<button
						type='button'
						key={item.id}
						onClick={() => onOpenImage(item.id)}
						className={`filmstrip-item ${activeImage.id === item.id ? 'is-active' : ''}`}
						aria-label={t('IMAGES_OPEN_IMAGE_ARIA', { name: item.name })}
					>
						<img src={thumbnailUrl(item.id)} alt={item.name} loading='lazy' />
					</button>
				))}
			</div>
		</div>
	);
}
