import { MusicProvider } from '@/features/music/providers/musicProvider/musicProvider';
import Layout from '../layout/Layout';

const MusicLayout = ({ children }: { children: React.ReactNode }) => {
    return (
        <Layout>
            <MusicProvider>{children}</MusicProvider>
        </Layout>
    );
};

export default MusicLayout;
