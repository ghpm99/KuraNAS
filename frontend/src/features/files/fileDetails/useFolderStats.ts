import { getFolderStats } from '@/service/files';
import type { FolderStats } from '@/types/folderStats';
import { useQuery } from '@tanstack/react-query';

const folderStatsStaleTimeMs = 30_000;

const useFolderStats = (folderId: number | undefined) =>
    useQuery<FolderStats>({
        queryKey: ['files', 'folder-stats', folderId],
        queryFn: () => getFolderStats(folderId as number),
        enabled: typeof folderId === 'number',
        staleTime: folderStatsStaleTimeMs,
        retry: false,
    });

export default useFolderStats;
