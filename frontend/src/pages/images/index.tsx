import DomainPageLayout from '@/components/layout/DomainPageLayout';
import ImageContent from '@/components/imageContent';
import ImagesLayout from '@/components/images/imagesLayout';
import ImageDomainHeader from '@/components/images/ImageDomainHeader';
import ImageDomainNav from '@/components/images/ImageDomainNav';

const ImagesPage = () => {
    return (
        <ImagesLayout>
            <DomainPageLayout header={<ImageDomainHeader />} nav={<ImageDomainNav />}>
                <ImageContent />
            </DomainPageLayout>
        </ImagesLayout>
    );
};

export default ImagesPage;
