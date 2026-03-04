import VideoContent from '@/components/videos/videoContent/videoContent';
import VideoLayout from '@/components/videos/videoLayout';
import styles from './videos.module.css';

const VideosPage = () => {
	return (
		<VideoLayout>
			<div className={styles.content}>
				<div className={styles.page}>
					<VideoContent />
				</div>
			</div>
		</VideoLayout>
	);
};

export default VideosPage;
