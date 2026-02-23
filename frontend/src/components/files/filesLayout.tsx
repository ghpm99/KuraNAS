import FileProvider from '../hooks/fileProvider';
import Layout from '../layout/Layout';

const FilesLayout = ({ children }: { children: React.ReactNode }) => {
	return (
		<FileProvider>
			<Layout>{children}</Layout>
		</FileProvider>
	);
};

export default FilesLayout;
