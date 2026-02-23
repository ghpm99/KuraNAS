import ImageContent from '@/components/imageContent';
import './images.css';
import ImagesLayout from '@/components/images/imagesLayout';

const ImagesPage = () => {
	return (
		<ImagesLayout>
			<div className='content'>
				<ImageContent />
			</div>
		</ImagesLayout>
	);
};

export default ImagesPage;
