import { MusicProvider } from '../hooks/musicProvider/musicProvider';
import Layout from '../layout/Layout';

const MusicLayout = ({ children }: { children: React.ReactNode }) => {
	return (
		<Layout>
			<MusicProvider>{children}</MusicProvider>
		</Layout>
	);
};

export default MusicLayout;
