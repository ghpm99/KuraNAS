import { formatSize } from '@/utils';
import { CircularProgress, IconButton, ImageList, ImageListItem, ImageListItemBar } from '@mui/material';
import { InfoIcon } from 'lucide-react';
import { useImage } from '../hooks/imageProvider/imageProvider';
import { useIntersectionObserver } from '../hooks/IntersectionObserver/useIntersectionObserver';
import './imageContent.css';

const thumbnailWidth = 760;
const thumbnailHeight = 760;

const MusicContent = () => {
	const { images, fetchNextPage, hasNextPage, isFetchingNextPage } = useImage();
	const { ref: lastItemRef } = useIntersectionObserver<HTMLLIElement>({
		enabled: hasNextPage && !isFetchingNextPage,
		rootMargin: '400px',
		onIntersect: () => {
			if (hasNextPage && !isFetchingNextPage) {
				fetchNextPage();
			}
		},
	});

	const imageMetadata = (image: { format: string; size: number }): string => {
		const format = image.format ? `${image.format} - ` : '';
		const fileSize = formatSize(image.size);
		return `${format}${fileSize}`;
	};

	const thumbnailUrl = (id: number) =>
		`${import.meta.env.VITE_API_URL}/api/v1/files/thumbnail/${id}?width=${thumbnailWidth}&height=${thumbnailHeight}`;

	return (
		<div className='file-content'>
			<ImageList cols={3} gap={8} rowHeight={thumbnailHeight}>
				{images.map((item, index) => {
					const isLastItem = index === images.length - 1;
					return (
						<ImageListItem key={item.id} ref={isLastItem ? lastItemRef : null}>
							<img className='thumbnail-img' src={thumbnailUrl(item.id)} alt={item.name} loading='lazy' />
							<ImageListItemBar
								title={item.name}
								subtitle={imageMetadata(item)}
								actionIcon={
									<IconButton sx={{ color: 'rgba(255, 255, 255, 0.54)' }} aria-label={`info about ${item.name}`}>
										<InfoIcon />
									</IconButton>
								}
							/>
						</ImageListItem>
					);
				})}
			</ImageList>

			{isFetchingNextPage && (
				<div className='loading-indicator'>
					<CircularProgress size={40} />
				</div>
			)}

			{!hasNextPage && images.length > 0 && <div className='end-message'>Todas as imagens carregadas</div>}
		</div>
	);
};

export default MusicContent;
