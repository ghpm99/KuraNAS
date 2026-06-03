import FileProvider from '@/features/files/providers/fileProvider';
import Layout from '@/components/layout/Layout';

const FilesLayout = ({ children }: { children: React.ReactNode }) => {
    return (
        <FileProvider>
            <Layout>{children}</Layout>
        </FileProvider>
    );
};

export default FilesLayout;
