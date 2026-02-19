import { formatSize } from '@/utils';
import { IImageData, useImage } from '../hooks/imageProvider/imageProvider';
import { useIntersectionObserver } from '../hooks/IntersectionObserver/useIntersectionObserver';
import './imageContent.css';
import {
	IconButton,
	ImageList,
	ImageListItem,
	ImageListItemBar,
	ListSubheader,
	CircularProgress,
	CardMedia,
} from '@mui/material';
import { InfoIcon } from 'lucide-react';

const ImageContent = () => {
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

	const imageMetadata = (image: IImageData): string => {
		const format = image.format ? `${image.format} - ` : '';
		const fileSize = formatSize(image.size);
		return `${format}${fileSize}`;
	};

	const thumbnailUrl = (id: number) => `${import.meta.env.VITE_API_URL}/api/v1/files/thumbnail/${id}`;

	return (
		<div className='file-content'>
			<ImageList cols={3} rowHeight={489}>
				<ImageListItem key='Subheader' cols={3}>
					<ListSubheader component='div'>Images</ListSubheader>
				</ImageListItem>
				{images.map((item, index) => {
					const isLastItem = index === images.length - 1;
					return (
						<ImageListItem key={item.id} ref={isLastItem ? lastItemRef : null}>
							<CardMedia
								component='img'
								width={652}
								height={489}
								image={thumbnailUrl(item.id)}
								alt={item.name}
								loading='lazy'
								sx={{ objectFit: 'cover' }}
							/>
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

export default ImageContent;
