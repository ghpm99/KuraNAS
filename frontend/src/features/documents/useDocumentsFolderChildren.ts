import { useInfiniteQuery } from '@tanstack/react-query';
import { getFilesTree } from '@/service/files';
import { FileType } from '@/utils';
import { getNextPageNumber } from '@/shared/pagination/getNextPageNumber';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

export const documentsFolderPageSize = 50;

const useDocumentsFolderChildren = (parentFolderId: number) => {
    const childrenQuery = useInfiniteQuery({
        queryKey: ['documents', 'folders', parentFolderId],
        queryFn: ({ pageParam = 1 }) =>
            getFilesTree({
                page: pageParam,
                pageSize: documentsFolderPageSize,
                fileParent: parentFolderId,
                category: 'all',
            }),
        initialPageParam: 1,
        getNextPageParam: getNextPageNumber,
        retry: false,
    });

    const subfolders: FileData[] = (childrenQuery.data?.pages ?? [])
        .flatMap((page) => page?.items ?? [])
        .filter((child) => child?.type === FileType.Directory);

    return {
        subfolders,
        isLoading: childrenQuery.isPending && !childrenQuery.isError,
        isError: childrenQuery.isError,
        hasNextPage: Boolean(childrenQuery.hasNextPage),
        isFetchingNextPage: childrenQuery.isFetchingNextPage,
        fetchNextPage: () => {
            void childrenQuery.fetchNextPage();
        },
        retry: () => {
            void childrenQuery.refetch();
        },
    };
};

export default useDocumentsFolderChildren;
