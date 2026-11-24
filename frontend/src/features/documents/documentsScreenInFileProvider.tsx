import FileProvider from '@/features/files/providers/fileProvider';
import DocumentsScreen from './DocumentsScreen';

const DocumentsScreenInFileProvider = () => (
    <FileProvider>
        <DocumentsScreen />
    </FileProvider>
);

export default DocumentsScreenInFileProvider;
