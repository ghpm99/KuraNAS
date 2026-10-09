import { MusicProvider } from '@/features/music/providers/musicProvider/musicProvider';

const MusicLayout = ({ children }: { children: React.ReactNode }) => {
    return <MusicProvider>{children}</MusicProvider>;
};

export default MusicLayout;
