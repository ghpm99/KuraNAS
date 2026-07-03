import DomainPageLayout from '@/components/layout/DomainPageLayout';
import VideoDomainHeader from '@/features/videos/components/VideoDomainHeader';
import VideoDomainNav from '@/features/videos/components/VideoDomainNav';
import VideoContent from '@/features/videos/components/videoContent/videoContent';

const VideosPage = () => {
    return (
        <DomainPageLayout header={<VideoDomainHeader />} nav={<VideoDomainNav />}>
            <VideoContent />
        </DomainPageLayout>
    );
};

export default VideosPage;
