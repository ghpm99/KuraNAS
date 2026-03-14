import VideoDomainHeader from '@/components/videos/VideoDomainHeader';
import VideoSidebar from '@/components/videos/VideoSidebar';
import VideoContent from '@/components/videos/videoContent/videoContent';
import VideoLayout from '@/components/videos/videoLayout';
import styles from './videos.module.css';

const VideosPage = () => {
	return (
		<VideoLayout>
			<div className={styles.content}>
				<div className={styles.page}>
					<VideoDomainHeader />
					<div className={styles.domainContent}>
						<VideoSidebar />
						<VideoContent />
					</div>
				</div>
			</div>
		</VideoLayout>
	);
};

export default VideosPage;
