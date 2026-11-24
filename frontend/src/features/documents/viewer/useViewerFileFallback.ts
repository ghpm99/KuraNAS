import { useQuery } from '@tanstack/react-query';
import { getFileByPath, getFileLocation } from '@/service/files';
import { documentViewerFileQueryKey } from '../documentsQueryKeys';

const useViewerFileFallback = (fileId: number, isEnabled: boolean) =>
    useQuery({
        queryKey: documentViewerFileQueryKey(fileId),
        queryFn: async () => {
            const location = await getFileLocation(fileId);
            const file = await getFileByPath(location.logical_path);
            if (!file) throw new Error('document not found');
            return file;
        },
        enabled: isEnabled && Number.isInteger(fileId),
        retry: false,
    });

export default useViewerFileFallback;
