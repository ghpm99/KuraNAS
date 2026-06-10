import { MusicProvider } from '@/features/music/providers/musicProvider/musicProvider';
import Layout from '@/components/layout/Layout';

const MusicLayout = ({ children }: { children: React.ReactNode }) => {
    return (
        <Layout>
            <MusicProvider>{children}</MusicProvider>
        </Layout>
    );
};

export default MusicLayout;
