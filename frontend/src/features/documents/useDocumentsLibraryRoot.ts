import { useQuery } from '@tanstack/react-query';
import { getFileByDiskPath } from '@/service/files';
import { getLibraries } from '@/service/libraries';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';

export type DocumentsLibraryRootStatus = 'loading' | 'error' | 'not-configured' | 'ready';

export const documentsLibraryRootQueryKey = ['documents', 'library-root'] as const;

const resolveDocumentsLibraryRoot = async (): Promise<FileData | null> => {
    const libraries = await getLibraries();
    const documentsLibrary = (Array.isArray(libraries) ? libraries : []).find(
        (library) => library?.category === 'documents'
    );
    if (!documentsLibrary?.path) return null;
    return getFileByDiskPath(documentsLibrary.path);
};

const useDocumentsLibraryRoot = () => {
    const rootQuery = useQuery({
        queryKey: documentsLibraryRootQueryKey,
        queryFn: resolveDocumentsLibraryRoot,
        retry: false,
    });

    const status: DocumentsLibraryRootStatus = (() => {
        if (rootQuery.isError) return 'error';
        if (rootQuery.isPending) return 'loading';
        return rootQuery.data ? 'ready' : 'not-configured';
    })();

    return {
        status,
        root: rootQuery.data ?? undefined,
        retry: () => {
            void rootQuery.refetch();
        },
    };
};

export default useDocumentsLibraryRoot;
