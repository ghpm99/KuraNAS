import DomainPageLayout from '@/components/layout/DomainPageLayout';
import MusicDomainHeader from '@/features/music/components/MusicDomainHeader';
import MusicLayout from '@/features/music/components/musicLayout';
import MusicContent from '@/features/music/components/musicContent';
import MusicDomainNav from '@/features/music/components/MusicDomainNav';

const MusicPage = () => {
    return (
        <MusicLayout>
            <DomainPageLayout header={<MusicDomainHeader />} nav={<MusicDomainNav />}>
                <MusicContent />
            </DomainPageLayout>
        </MusicLayout>
    );
};

export default MusicPage;
