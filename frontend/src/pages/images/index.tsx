import ImageContent from '@/components/imageContent';
import styles from './images.module.css';
import ImagesLayout from '@/components/images/imagesLayout';
import ImageDomainHeader from '@/components/images/ImageDomainHeader';
import ImageSidebar from '@/components/images/ImageSidebar';

const ImagesPage = () => {
	return (
		<ImagesLayout>
			<div className={styles.page}>
				<ImageDomainHeader />
				<div className={styles.content}>
					<ImageSidebar />
					<div className={styles.main}>
						<ImageContent />
					</div>
				</div>
			</div>
		</ImagesLayout>
	);
};

export default ImagesPage;
