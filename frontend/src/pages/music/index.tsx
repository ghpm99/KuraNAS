import DomainPageLayout from '@/components/layout/DomainPageLayout';
import MusicDomainHeader from '@/features/music/components/MusicDomainHeader';
import MusicLayout from '@/features/music/components/musicLayout';
import MusicContent from '@/features/music/components/musicContent';
import MusicSidebar from '@/features/music/components/MusicSidebar';

const MusicPage = () => {
    return (
        <MusicLayout>
            <DomainPageLayout header={<MusicDomainHeader />} sidebar={<MusicSidebar />}>
                <MusicContent />
            </DomainPageLayout>
        </MusicLayout>
    );
};

export default MusicPage;
