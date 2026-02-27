import { AboutProvider } from '../hooks/AboutProvider';
import Layout from '../layout/Layout';

const AboutLayout = ({ children }: { children: React.ReactNode }) => {
	return (
		<AboutProvider>
			<Layout>{children}</Layout>
		</AboutProvider>
	);
};

export default AboutLayout;
