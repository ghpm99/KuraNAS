import { ImageProvider } from '../providers/imageProvider/imageProvider';
import Layout from '../layout/Layout';

const ImagesLayout = ({ children }: { children: React.ReactNode }) => {
    return (
        <ImageProvider>
            <Layout>{children}</Layout>
        </ImageProvider>
    );
};

export default ImagesLayout;
