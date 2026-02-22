import VideoContent from '@/components/videos/videoContent/videoContent';
import './videos.css';
import VideoLayout from '@/components/videos/videoLayout';

const VideosPage = () => {
	return (
		<VideoLayout>
			<div className='content'>
				<VideoContent />
			</div>
		</VideoLayout>
	);
};

export default VideosPage;
