import VideoContent from '@/components/videos/videoContent/videoContent';
import VideoLayout from '@/components/videos/videoLayout';
import styles from './videos.module.css';

const VideosPage = () => {
	return (
		<VideoLayout>
			<div className={styles.videosPage}>
				<div className={styles.videosViewport}>
				<VideoContent />
				</div>
			</div>
		</VideoLayout>
	);
};

export default VideosPage;
