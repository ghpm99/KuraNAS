import { VideoPlayerProvider } from '../hooks/videoPlayerProvider/videoPlayerProvider';
import Layout from '../layout/Layout';

const VideoLayout = ({ children }: { children: React.ReactNode }) => {
	return (
		<Layout>
			<VideoPlayerProvider>{children}</VideoPlayerProvider>
		</Layout>
	);
};

export default VideoLayout;
