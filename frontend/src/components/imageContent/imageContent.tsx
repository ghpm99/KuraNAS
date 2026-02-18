import { formatSize } from '@/utils';
import FileCard from '../fileCard';
import { IImageData, useImage } from '../hooks/imageProvider/imageProvider';
import './imageContent.css';

const ImageContent = () => {
	const { images, status } = useImage();
	console.log('Images status', status);

	const imageMetadata = (image: IImageData): string => {
		const format = image.format ? `${image.format} - ` : '';
		const fileSize = formatSize(image.size);

		return `${format}${fileSize}`;
	};

	const thumbnailUrl = (id: number) => `${import.meta.env.VITE_API_URL}/api/v1/files/thumbnail/${id}`;

	return (
		<div className='file-content'>
			{images?.pages.map((page, index) => (
				<div key={index} className='file-grid'>
					{page.items.map((image) => (
						<FileCard
							key={image.id}
							title={image.name}
							starred={image.starred}
							metadata={imageMetadata(image)}
							thumbnail={thumbnailUrl(image.id)}
							onClick={() => console.log('Clicked image', image.id)}
							onClickStar={() => console.log('Clicked star for image', image.id)}
						/>
					))}
				</div>
			))}
		</div>
	);
};

export default ImageContent;
