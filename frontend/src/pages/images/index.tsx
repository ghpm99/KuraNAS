import DomainPageLayout from '@/components/layout/DomainPageLayout';
import ImageContent from '@/components/imageContent';
import ImagesLayout from '@/components/images/imagesLayout';
import ImageDomainHeader from '@/components/images/ImageDomainHeader';
import ImageSidebar from '@/components/images/ImageSidebar';

const ImagesPage = () => {
	return (
		<ImagesLayout>
			<DomainPageLayout
				header={<ImageDomainHeader />}
				sidebar={<ImageSidebar />}
			>
				<ImageContent />
			</DomainPageLayout>
		</ImagesLayout>
	);
};

export default ImagesPage;
