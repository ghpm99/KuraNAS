import DomainPageLayout from '@/components/layout/DomainPageLayout';
import VideoDomainHeader from '@/components/videos/VideoDomainHeader';
import VideoSidebar from '@/components/videos/VideoSidebar';
import VideoContent from '@/components/videos/videoContent/videoContent';
import VideoLayout from '@/components/videos/videoLayout';

const VideosPage = () => {
	return (
		<VideoLayout>
			<DomainPageLayout
				header={<VideoDomainHeader />}
				sidebar={<VideoSidebar />}
			>
				<VideoContent />
			</DomainPageLayout>
		</VideoLayout>
	);
};

export default VideosPage;
