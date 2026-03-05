import ImageContent from '@/components/imageContent';
import styles from './images.module.css';
import ImagesLayout from '@/components/images/imagesLayout';

const ImagesPage = () => {
	return (
		<ImagesLayout>
			<div className={styles['content']}>
				<ImageContent />
			</div>
		</ImagesLayout>
	);
};

export default ImagesPage;
