import FileProvider from '@/features/files/providers/fileProvider';

const FilesLayout = ({ children }: { children: React.ReactNode }) => {
    return <FileProvider>{children}</FileProvider>;
};

export default FilesLayout;
