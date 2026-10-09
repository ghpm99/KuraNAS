import FileProvider from '@/features/files/providers/fileProvider';
import UploadQueueProvider from '@/features/files/upload/uploadQueueProvider';

const FilesLayout = ({ children }: { children: React.ReactNode }) => {
    return (
        <FileProvider>
            <UploadQueueProvider>{children}</UploadQueueProvider>
        </FileProvider>
    );
};

export default FilesLayout;
