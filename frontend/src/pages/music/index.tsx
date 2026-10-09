import DomainPageLayout from '@/components/layout/DomainPageLayout';
import MusicDomainHeader from '@/features/music/components/MusicDomainHeader';
import MusicContent from '@/features/music/components/musicContent';
import MusicDomainNav from '@/features/music/components/MusicDomainNav';

const MusicPage = () => {
    return (
        <DomainPageLayout header={<MusicDomainHeader />} nav={<MusicDomainNav />} width="full">
            <MusicContent />
        </DomainPageLayout>
    );
};

export default MusicPage;
