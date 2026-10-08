import { useQuery } from '@tanstack/react-query';
import { getFileAncestors } from '@/service/files';

const ancestorsStaleTimeMs = 30_000;

const useFileAncestors = (fileId: number | null) =>
    useQuery({
        queryKey: ['files-ancestors', fileId],
        queryFn: () => getFileAncestors(fileId as number),
        enabled: fileId !== null,
        staleTime: ancestorsStaleTimeMs,
        retry: false,
    });

export default useFileAncestors;
