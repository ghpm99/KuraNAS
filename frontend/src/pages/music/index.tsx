import DomainPageLayout from '@/components/layout/DomainPageLayout';
import MusicDomainHeader from '@/components/music/MusicDomainHeader';
import MusicLayout from '@/components/music/musicLayout';
import MusicContent from '@/components/musicContent';
import MusicSidebar from '@/components/music/MusicSidebar';

const MusicPage = () => {
	return (
		<MusicLayout>
			<DomainPageLayout
				header={<MusicDomainHeader />}
				sidebar={<MusicSidebar />}
			>
				<MusicContent />
			</DomainPageLayout>
		</MusicLayout>
	);
};

export default MusicPage;
