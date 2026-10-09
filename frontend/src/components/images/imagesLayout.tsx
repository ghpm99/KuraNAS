import { ImageProvider } from '../providers/imageProvider/imageProvider';

const ImagesLayout = ({ children }: { children: React.ReactNode }) => {
    return <ImageProvider>{children}</ImageProvider>;
};

export default ImagesLayout;
