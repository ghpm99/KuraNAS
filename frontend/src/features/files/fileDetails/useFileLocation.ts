import { getFileLocation } from '@/service/files';
import type { FileLocation } from '@/types/fileLocation';
import { useQuery } from '@tanstack/react-query';

const useFileLocation = (fileId: number | undefined) => {
    const { data } = useQuery<FileLocation>({
        queryKey: ['files', 'location', fileId],
        queryFn: () => getFileLocation(fileId as number),
        enabled: typeof fileId === 'number',
        retry: false,
    });

    return data ?? null;
};

export default useFileLocation;
