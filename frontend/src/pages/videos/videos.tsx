import DomainPageLayout from '@/components/layout/DomainPageLayout';
import VideoDomainHeader from '@/features/videos/components/VideoDomainHeader';
import VideoSidebar from '@/features/videos/components/VideoSidebar';
import VideoContent from '@/features/videos/components/videoContent/videoContent';
import VideoLayout from '@/features/videos/components/videoLayout';

const VideosPage = () => {
    return (
        <VideoLayout>
            <DomainPageLayout header={<VideoDomainHeader />} sidebar={<VideoSidebar />}>
                <VideoContent />
            </DomainPageLayout>
        </VideoLayout>
    );
};

export default VideosPage;
